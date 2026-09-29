package server

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"lockbox/internal/crypto"
)

type testEnv struct {
	t     *testing.T
	srv   *Server
	ts    *httptest.Server
	token string
	vault string
}

func newTestEnv(t *testing.T, lockTimeout time.Duration) *testEnv {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	vaultPath := filepath.Join(home, "vault.lock")
	t.Setenv("LOCKBOX_VAULT_PATH", vaultPath)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	srv, err := New(Options{Port: port, LockTimeout: lockTimeout})
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewUnstartedServer(srv.Handler())
	ts.Listener.Close()
	ts.Listener = ln
	ts.Start()
	t.Cleanup(func() {
		ts.Close()
		srv.Shutdown()
	})

	return &testEnv{t: t, srv: srv, ts: ts, token: srv.Token(), vault: vaultPath}
}

func (e *testEnv) do(method, path string, body any, token string) *httptest.ResponseRecorder {
	e.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			e.t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, e.ts.URL+path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("X-Lockbox-Token", token)
	}
	rec := httptest.NewRecorder()
	e.ts.Config.Handler.ServeHTTP(rec, req)
	return rec
}

func (e *testEnv) mustDo(method, path string, body any, wantStatus int) map[string]any {
	e.t.Helper()
	rec := e.do(method, path, body, e.token)
	if rec.Code != wantStatus {
		e.t.Fatalf("%s %s: status=%d want=%d body=%s", method, path, rec.Code, wantStatus, rec.Body.String())
	}
	if rec.Body.Len() == 0 {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		e.t.Fatalf("bad json: %s", rec.Body.String())
	}
	return out
}

func (e *testEnv) createVault(password string) {
	e.t.Helper()
	e.mustDo("POST", "/api/unlock", map[string]string{"password": password, "confirm": password}, 200)
}

func TestTokenAndHostEnforcement(t *testing.T) {
	e := newTestEnv(t, time.Minute)

	if rec := e.do("GET", "/api/session", nil, ""); rec.Code != http.StatusForbidden {
		t.Errorf("no token: got %d want 403", rec.Code)
	}
	if rec := e.do("GET", "/api/session", nil, "wrong-token"); rec.Code != http.StatusForbidden {
		t.Errorf("bad token: got %d want 403", rec.Code)
	}
	if rec := e.do("GET", "/api/session", nil, e.token); rec.Code != http.StatusOK {
		t.Errorf("good token: got %d want 200", rec.Code)
	}

	// Host check (DNS rebinding guard).
	req := httptest.NewRequest("GET", "/api/session", nil)
	req.Host = "evil.example.com"
	req.Header.Set("X-Lockbox-Token", e.token)
	rec := httptest.NewRecorder()
	e.ts.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("bad host: got %d want 403", rec.Code)
	}

	// Origin check (CSRF guard).
	req2 := httptest.NewRequest("POST", "/api/unlock", bytes.NewReader([]byte(`{"password":"x","confirm":"x"}`)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Lockbox-Token", e.token)
	req2.Header.Set("Origin", "https://evil.example.com")
	rec2 := httptest.NewRecorder()
	e.ts.Config.Handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusForbidden {
		t.Errorf("bad origin: got %d want 403", rec2.Code)
	}
}

func TestCreateFlowAndUnlock(t *testing.T) {
	e := newTestEnv(t, time.Minute)

	// Session before unlock.
	s := e.mustDo("GET", "/api/session", nil, 200)
	if s["vaultExists"] != false {
		t.Errorf("vaultExists=%v want false", s["vaultExists"])
	}

	// Confirm mismatch rejected.
	e.mustDo("POST", "/api/unlock", map[string]string{"password": "pw123456", "confirm": "other"}, 400)

	// Create.
	e.mustDo("POST", "/api/unlock", map[string]string{"password": "pw123456", "confirm": "pw123456"}, 200)
	if _, err := crypto.GenerateSalt(); err != nil {
		t.Fatal(err)
	}
	s = e.mustDo("GET", "/api/session", nil, 200)
	if s["vaultExists"] != true || s["unlocked"] != true {
		t.Errorf("after create: %v", s)
	}

	// Lock + wrong password rejected.
	e.mustDo("POST", "/api/lock", nil, 200)
	e.mustDo("POST", "/api/unlock", map[string]string{"password": "wrong-password"}, 401)
	e.mustDo("POST", "/api/unlock", map[string]string{"password": "pw123456"}, 200)
}

func TestVaultCRUDAndReorder(t *testing.T) {
	e := newTestEnv(t, time.Minute)
	e.createVault("pw123456")

	add := func(acc string) string {
		res := e.mustDo("POST", "/api/vault", map[string]string{
			"account": acc, "username": "user-" + acc, "password": "pw-" + acc,
		}, http.StatusCreated)
		cred := res["credential"].(map[string]any)
		return cred["id"].(string)
	}
	idA := add("a.example")
	idB := add("b.example")
	idC := add("c.example")

	list := func() []any {
		res := e.mustDo("GET", "/api/vault", nil, 200)
		return res["credentials"].([]any)
	}
	if got := len(list()); got != 3 {
		t.Fatalf("list len=%d want 3", got)
	}

	// Search filter.
	res := e.mustDo("GET", "/api/vault?search=b.example", nil, 200)
	if got := len(res["credentials"].([]any)); got != 1 {
		t.Errorf("search len=%d want 1", got)
	}

	// Edit all fields.
	e.mustDo("PUT", "/api/vault/"+idA, map[string]string{
		"account": "a.example", "username": "newuser", "password": "newpass",
	}, 200)
	first := list()[0].(map[string]any)
	if first["username"] != "newuser" || first["password"] != "newpass" {
		t.Errorf("edit not applied: %v", first)
	}

	// Reorder: move C (position 3) to position 1.
	e.mustDo("POST", "/api/vault/reorder", map[string]any{"id": idC, "to": 1}, 200)
	order := list()
	if order[0].(map[string]any)["id"] != idC {
		t.Errorf("reorder failed, first=%v want idC", order[0])
	}

	// Delete B.
	e.mustDo("DELETE", "/api/vault/"+idB, nil, 200)
	if got := len(list()); got != 2 {
		t.Errorf("after delete len=%d want 2", got)
	}
	// Delete unknown id → 404.
	e.mustDo("DELETE", "/api/vault/does-not-exist", nil, 404)
}

func TestGenerate(t *testing.T) {
	e := newTestEnv(t, time.Minute)
	e.createVault("pw123456")

	res := e.mustDo("POST", "/api/generate", map[string]int{"length": 12}, 200)
	if got := len(res["password"].(string)); got != 12 {
		t.Errorf("len=%d want 12", got)
	}
	res = e.mustDo("POST", "/api/generate", map[string]int{"length": 0}, 200)
	if got := len(res["password"].(string)); got != 12 {
		t.Errorf("default len=%d want 12", got)
	}
	e.mustDo("POST", "/api/generate", map[string]int{"length": 2}, 400)
}

func TestChangeMaster(t *testing.T) {
	e := newTestEnv(t, time.Minute)
	e.createVault("oldpw123456")
	e.mustDo("POST", "/api/vault", map[string]string{"account": "x", "username": "y", "password": "z"}, 201)

	e.mustDo("POST", "/api/change-master", map[string]string{"newPassword": "newpw123456"}, 200)
	e.mustDo("POST", "/api/lock", nil, 200)
	e.mustDo("POST", "/api/unlock", map[string]string{"password": "oldpw123456"}, 401)
	e.mustDo("POST", "/api/unlock", map[string]string{"password": "newpw123456"}, 200)

	res := e.mustDo("GET", "/api/vault", nil, 200)
	if got := len(res["credentials"].([]any)); got != 1 {
		t.Errorf("creds after change-master=%d want 1", got)
	}
}

func TestAutoLock(t *testing.T) {
	e := newTestEnv(t, 50*time.Millisecond)
	e.createVault("pw123456")
	e.mustDo("GET", "/api/vault", nil, 200)

	time.Sleep(120 * time.Millisecond)

	rec := e.do("GET", "/api/vault", nil, e.token)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("after idle: status=%d want 401", rec.Code)
	}
}

func TestConfigEndpoints(t *testing.T) {
	e := newTestEnv(t, time.Minute)

	res := e.mustDo("GET", "/api/config", nil, 200)
	if res["kdfIterations"].(float64) != 100000 {
		t.Errorf("kdfIterations=%v", res["kdfIterations"])
	}

	e.createVault("pw123456")

	// Bad key rejected.
	e.mustDo("POST", "/api/config", map[string]string{"key": "bogus", "value": "1"}, 400)

	// Vault path change invalidates session.
	newPath := filepath.Join(filepath.Dir(e.vault), "vault2.lock")
	res = e.mustDo("POST", "/api/config", map[string]string{"key": "vault_path", "value": newPath}, 200)
	if res["sessionInvalidated"] != true {
		t.Errorf("sessionInvalidated=%v want true", res["sessionInvalidated"])
	}
	rec := e.do("GET", "/api/vault", nil, e.token)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("after vault_path change: status=%d want 401", rec.Code)
	}
}

func writeLegacyVaultFile(t *testing.T, path, password string) {
	t.Helper()
	plaintext := []byte(`[
		{"index":2,"account":"second","username":"u2","password":"p2","saved_at":"2026-06-01T10:00:00Z"},
		{"index":1,"account":"first","username":"u1","password":"p1","saved_at":"2026-06-01T09:00:00Z"}
	]`)
	salt := make([]byte, crypto.SaltSize)
	if _, err := rand.Read(salt); err != nil {
		t.Fatal(err)
	}
	key := crypto.DeriveKey(password, salt, 100000)
	ct, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(salt, ct...), 0600); err != nil {
		t.Fatal(err)
	}
}

func TestLegacyVaultUpgradeOnUnlock(t *testing.T) {
	e := newTestEnv(t, time.Minute)
	writeLegacyVaultFile(t, e.vault, "legacypw1")

	// Unlock migrates: assigns IDs, maps index→SortOrder, rewrites file.
	e.mustDo("POST", "/api/unlock", map[string]string{"password": "legacypw1"}, 200)

	res := e.mustDo("GET", "/api/vault", nil, 200)
	creds := res["credentials"].([]any)
	if len(creds) != 2 {
		t.Fatalf("creds=%d want 2", len(creds))
	}
	first := creds[0].(map[string]any)
	second := creds[1].(map[string]any)
	// Old display order was by index: first(index1) then second(index2),
	// even though file order was [second, first].
	if first["account"] != "first" || second["account"] != "second" {
		t.Errorf("legacy order wrong: %v then %v", first["account"], second["account"])
	}
	id1 := first["id"].(string)
	id2 := second["id"].(string)
	if id1 == "" || id2 == "" || id1 == id2 {
		t.Errorf("IDs not assigned: %q %q", id1, id2)
	}

	// IDs persist: lock, unlock again → same IDs.
	e.mustDo("POST", "/api/lock", nil, 200)
	e.mustDo("POST", "/api/unlock", map[string]string{"password": "legacypw1"}, 200)
	res = e.mustDo("GET", "/api/vault", nil, 200)
	creds = res["credentials"].([]any)
	if creds[0].(map[string]any)["id"].(string) != id1 {
		t.Errorf("IDs not persisted across sessions")
	}
}

func TestIndexPageInjectsToken(t *testing.T) {
	e := newTestEnv(t, time.Minute)
	rec := e.do("GET", "/", nil, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status=%d", rec.Code)
	}
	body := rec.Body.String()
	if bytes.Contains([]byte(body), []byte("UI not built")) {
		t.Skip("frontend assets not built (run: make ui)")
	}
	if !bytes.Contains([]byte(body), []byte(e.token)) {
		t.Errorf("token not injected into index page")
	}
	if bytes.Contains([]byte(body), []byte(`"`+tokenPlaceholder+`"`)) {
		t.Errorf("placeholder value still present in index page")
	}
}

func TestBootstrapTokenEndpoint(t *testing.T) {
	e := newTestEnv(t, time.Minute)
	rec := e.do("GET", "/api/token", nil, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	var out map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["token"] != e.token {
		t.Errorf("token mismatch")
	}
}
