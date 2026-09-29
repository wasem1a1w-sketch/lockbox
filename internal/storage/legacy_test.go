package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"lockbox/internal/crypto"
	"lockbox/internal/vault"
)

// writeLegacyVault writes a pre-refactor vault file: [16B salt][ciphertext],
// no iterations header, key derived with 100000 iterations.
func writeLegacyVault(t *testing.T, path, password string, v vault.Vault) {
	t.Helper()
	plaintext, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	salt, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	key := crypto.DeriveKey(password, salt, 100000)
	ciphertext, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(salt, ciphertext...), 0600); err != nil {
		t.Fatal(err)
	}
}

func sampleVault() vault.Vault {
	return vault.Vault{
		{ID: "a1b2c3d4-0000-4000-8000-000000000001", Account: "example.com", Username: "alice", Password: "old-pass-1"},
		{ID: "a1b2c3d4-0000-4000-8000-000000000002", Account: "github.com", Username: "bob", Password: "old-pass-2"},
	}
}

func TestLoadLegacyFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.lock")
	writeLegacyVault(t, path, "MasterPw1", sampleVault())

	s := NewSecureStorage(path)
	got, iterations, err := s.LoadWithPassword("MasterPw1", 0)
	if err != nil {
		t.Fatalf("legacy load failed: %v", err)
	}
	if iterations != 100000 {
		t.Errorf("iterations = %d, want 100000", iterations)
	}
	if len(got) != 2 || got[0].Password != "old-pass-1" || got[1].Account != "github.com" {
		t.Errorf("unexpected vault content: %+v", got)
	}
}

func TestLoadLegacyFormatWrongPassword(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.lock")
	writeLegacyVault(t, path, "MasterPw1", sampleVault())

	s := NewSecureStorage(path)
	if _, _, err := s.LoadWithPassword("wrong", 0); err == nil {
		t.Fatal("want auth error, got nil")
	}
}

func TestLoadCurrentFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.lock")
	s := NewSecureStorage(path)

	salt, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	key := crypto.DeriveKey("MasterPw1", salt, 150000)
	if err := s.Save(sampleVault(), key, salt, 150000); err != nil {
		t.Fatal(err)
	}

	got, iterations, err := s.LoadWithPassword("MasterPw1", 0)
	if err != nil {
		t.Fatalf("current load failed: %v", err)
	}
	if iterations != 150000 {
		t.Errorf("iterations = %d, want 150000 (header wins)", iterations)
	}
	if len(got) != 2 {
		t.Errorf("want 2 creds, got %d", len(got))
	}
}

func TestLegacyMigratesOnSave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.lock")
	writeLegacyVault(t, path, "MasterPw1", sampleVault())

	s := NewSecureStorage(path)
	got, iterations, err := s.LoadWithPassword("MasterPw1", 0)
	if err != nil {
		t.Fatal(err)
	}
	got.Add("new.com", "carol", "np")
	salt := make([]byte, crypto.SaltSize)
	copy(salt, "0123456789abcdef")
	key := crypto.DeriveKey("MasterPw1", salt, iterations)
	if err := s.Save(got, key, salt, iterations); err != nil {
		t.Fatal(err)
	}

	// Re-open: must parse as current format with iterations header.
	got2, iterations2, err := s.LoadWithPassword("MasterPw1", 0)
	if err != nil {
		t.Fatalf("reload after migration failed: %v", err)
	}
	if iterations2 != 100000 {
		t.Errorf("iterations = %d, want 100000", iterations2)
	}
	if len(got2) != 3 {
		t.Errorf("want 3 creds after migration, got %d", len(got2))
	}
}

func TestLegacyIndexMapsToSortOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.lock")
	// Old CLI JSON: positional index, no id / sort_order. File order != index order.
	plaintext := []byte(`[
		{"index":2,"account":"second","username":"u2","password":"p2","saved_at":"2026-06-01T10:00:00Z"},
		{"index":1,"account":"first","username":"u1","password":"p1","saved_at":"2026-06-01T09:00:00Z"}
	]`)
	salt, err := crypto.GenerateSalt()
	if err != nil {
		t.Fatal(err)
	}
	key := crypto.DeriveKey("pw", salt, 100000)
	ct, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(salt, ct...), 0600); err != nil {
		t.Fatal(err)
	}

	s := NewSecureStorage(path)
	v, iterations, err := s.LoadWithPassword("pw", 0)
	if err != nil {
		t.Fatal(err)
	}
	if iterations != 100000 {
		t.Errorf("iterations=%d want 100000", iterations)
	}
	if v[0].SortOrder != 2 || v[1].SortOrder != 1 {
		t.Errorf("index not mapped: sortOrders %d %d want 2 1", v[0].SortOrder, v[1].SortOrder)
	}
	v.SortByOrder()
	if v[0].Account != "first" || v[1].Account != "second" {
		t.Errorf("display order wrong: %s, %s", v[0].Account, v[1].Account)
	}
	if v[0].ID != "" {
		t.Errorf("legacy ID should be empty until EnsureIDs, got %q", v[0].ID)
	}
}
