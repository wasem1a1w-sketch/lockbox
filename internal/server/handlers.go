package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	"lockbox/internal/config"
	"lockbox/internal/crypto"
	"lockbox/internal/storage"
	"lockbox/internal/vault"
)

// --- request / response shapes ---

type unlockRequest struct {
	Password string `json:"password"`
	Confirm  string `json:"confirm"`
}

type credentialRequest struct {
	Account  string `json:"account"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type reorderRequest struct {
	ID string `json:"id"`
	To int    `json:"to"`
}

type generateRequest struct {
	Length int `json:"length"`
}

type changeMasterRequest struct {
	NewPassword string `json:"newPassword"`
}

type configRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type credentialResponse struct {
	ID        string `json:"id"`
	Account   string `json:"account"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	SavedAt   string `json:"savedAt"`
	SortOrder int    `json:"sortOrder"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeBody(r *http.Request, dst any) error {
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		return errors.New("content-type must be application/json")
	}
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func loadConfig() (*configView, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return &configView{
		VaultPath:        cfg.VaultPath,
		KDFIterations:    cfg.KDFIterations,
		DefaultGenLength: cfg.DefaultGenLen,
		ConfigPath:       config.DefaultConfigPath(),
	}, nil
}

func toResponses(v vault.Vault) []credentialResponse {
	out := make([]credentialResponse, 0, len(v))
	for _, c := range v {
		out = append(out, credentialResponse{
			ID:        c.ID,
			Account:   c.Account,
			Username:  c.Username,
			Password:  c.Password,
			SavedAt:   c.SavedAt.Format("2006-01-02 15:04"),
			SortOrder: c.SortOrder,
		})
	}
	return out
}

// requireSession resolves config + live session or writes the proper error.
func (s *Server) requireSession(w http.ResponseWriter, r *http.Request) (*configView, *session) {
	cfg, err := s.currentConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config load failed: "+err.Error())
		return nil, nil
	}
	sess := s.sessions.get(cfg.VaultPath)
	if sess == nil {
		writeError(w, http.StatusUnauthorized, "vault locked")
		return nil, nil
	}
	sess.touch()
	return cfg, sess
}

// loadVault decrypts the current vault for the session key.
func loadVault(cfg *configView, sess *session) (vault.Vault, error) {
	st := storage.NewSecureStorage(cfg.VaultPath)
	v, _, err := st.Load(sess.key)
	return v, err
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func saveVault(cfg *configView, sess *session, v vault.Vault) error {
	st := storage.NewSecureStorage(cfg.VaultPath)
	return st.Save(v, sess.key, sess.salt, sess.iterations)
}

// --- handlers ---

func (s *Server) handleToken(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"token": s.token})
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.currentConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config load failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{
		"vaultExists": fileExists(cfg.VaultPath),
		"unlocked":    s.sessions.active(),
	})
}

func (s *Server) handleLock(w http.ResponseWriter, r *http.Request) {
	s.sessions.clear()
	writeJSON(w, http.StatusOK, map[string]bool{"locked": true})
}

func (s *Server) handleUnlock(w http.ResponseWriter, r *http.Request) {
	var req unlockRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	if req.Password == "" {
		writeError(w, http.StatusBadRequest, "password cannot be blank")
		return
	}
	cfg, err := s.currentConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config load failed")
		return
	}

	st := storage.NewSecureStorage(cfg.VaultPath)
	_, statErr := os.Stat(cfg.VaultPath)
	vaultExists := statErr == nil

	if !vaultExists {
		// New vault: create flow (confirm required).
		if req.Confirm != req.Password {
			writeError(w, http.StatusBadRequest, "passwords do not match")
			return
		}
		salt, err := crypto.GenerateSalt()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "initialization error: "+err.Error())
			return
		}
		key := crypto.DeriveKey(req.Password, salt, cfg.KDFIterations)
		if err := st.Save(vault.Vault{}, key, salt, cfg.KDFIterations); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create vault: "+err.Error())
			return
		}
		s.sessions.set(key, salt, cfg.KDFIterations, cfg.VaultPath)
		writeJSON(w, http.StatusOK, map[string]bool{"created": true, "unlocked": true})
		return
	}

	salt, err := st.GetOrGenerateSalt()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "initialization error: "+err.Error())
		return
	}
	v, iterations, err := st.LoadWithPassword(req.Password, 0)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "incorrect master password")
		return
	}
	key := crypto.DeriveKey(req.Password, salt, iterations)

	// One-time legacy upgrade: assign UUIDs / normalize order, persist.
	v.SortByOrder()
	if changed := v.EnsureIDs(); changed > 0 {
		if err := st.Save(v, key, salt, iterations); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to upgrade vault: "+err.Error())
			return
		}
	}

	s.sessions.set(key, salt, iterations, cfg.VaultPath)
	writeJSON(w, http.StatusOK, map[string]bool{"created": false, "unlocked": true})
}

func (s *Server) handleVaultList(w http.ResponseWriter, r *http.Request) {
	cfg, sess := s.requireSession(w, r)
	if sess == nil {
		return
	}
	v, err := loadVault(cfg, sess)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "vault locked")
		return
	}
	v.SortByOrder()
	search := r.URL.Query().Get("search")
	user := r.URL.Query().Get("user")
	filtered := v.Filter(search, user)
	writeJSON(w, http.StatusOK, map[string]any{"credentials": toResponses(filtered)})
}

func (s *Server) handleVaultAdd(w http.ResponseWriter, r *http.Request) {
	cfg, sess := s.requireSession(w, r)
	if sess == nil {
		return
	}
	var req credentialRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	if req.Account == "" || req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "account, username and password are required")
		return
	}
	v, err := loadVault(cfg, sess)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "vault locked")
		return
	}
	v.Add(req.Account, req.Username, req.Password)
	if err := saveVault(cfg, sess, v); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save vault: "+err.Error())
		return
	}
	created := v[len(v)-1]
	writeJSON(w, http.StatusCreated, map[string]any{
		"credential": credentialResponse{
			ID: created.ID, Account: created.Account, Username: created.Username,
			Password: created.Password, SavedAt: created.SavedAt.Format("2006-01-02 15:04"),
			SortOrder: created.SortOrder,
		},
	})
}

func (s *Server) handleVaultEdit(w http.ResponseWriter, r *http.Request) {
	cfg, sess := s.requireSession(w, r)
	if sess == nil {
		return
	}
	var req credentialRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	if req.Account == "" || req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "account, username and password are required")
		return
	}
	v, err := loadVault(cfg, sess)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "vault locked")
		return
	}
	id := r.PathValue("id")
	if err := v.EditFields(id, req.Account, req.Username, req.Password); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err := saveVault(cfg, sess, v); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save vault: "+err.Error())
		return
	}
	cred := v.FindByID(id)
	writeJSON(w, http.StatusOK, map[string]any{
		"credential": credentialResponse{
			ID: cred.ID, Account: cred.Account, Username: cred.Username,
			Password: cred.Password, SavedAt: cred.SavedAt.Format("2006-01-02 15:04"),
			SortOrder: cred.SortOrder,
		},
	})
}

func (s *Server) handleVaultDelete(w http.ResponseWriter, r *http.Request) {
	cfg, sess := s.requireSession(w, r)
	if sess == nil {
		return
	}
	v, err := loadVault(cfg, sess)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "vault locked")
		return
	}
	id := r.PathValue("id")
	if err := v.DeleteByID(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if err := saveVault(cfg, sess, v); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save vault: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (s *Server) handleVaultReorder(w http.ResponseWriter, r *http.Request) {
	cfg, sess := s.requireSession(w, r)
	if sess == nil {
		return
	}
	var req reorderRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	v, err := loadVault(cfg, sess)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "vault locked")
		return
	}
	v.SortByOrder()
	oldIndex := -1
	for i := range v {
		if v[i].ID == req.ID {
			oldIndex = i + 1
			break
		}
	}
	if oldIndex == -1 {
		writeError(w, http.StatusNotFound, "credential not found")
		return
	}
	if err := v.ReOrder(oldIndex, req.To); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := saveVault(cfg, sess, v); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save vault: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"reordered": true})
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.currentConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config load failed")
		return
	}
	var req generateRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	if req.Length == 0 {
		req.Length = cfg.DefaultGenLength
	}
	if req.Length < 4 || req.Length > 256 {
		writeError(w, http.StatusBadRequest, "length must be between 4 and 256")
		return
	}
	password, err := vault.GeneratePassword(req.Length)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"password": password})
}

func (s *Server) handleChangeMaster(w http.ResponseWriter, r *http.Request) {
	cfg, sess := s.requireSession(w, r)
	if sess == nil {
		return
	}
	var req changeMasterRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	if req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "new password cannot be blank")
		return
	}
	v, err := loadVault(cfg, sess)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "vault locked")
		return
	}
	newSalt, err := crypto.GenerateSalt()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate new salt: "+err.Error())
		return
	}
	newKey := crypto.DeriveKey(req.NewPassword, newSalt, sess.iterations)
	st := storage.NewSecureStorage(cfg.VaultPath)
	if err := st.Save(v, newKey, newSalt, sess.iterations); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save vault: "+err.Error())
		return
	}
	s.sessions.set(newKey, newSalt, sess.iterations, cfg.VaultPath)
	writeJSON(w, http.StatusOK, map[string]bool{"changed": true})
}

func (s *Server) handleConfigGet(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.currentConfig()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config load failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"vaultPath":        cfg.VaultPath,
		"kdfIterations":    cfg.KDFIterations,
		"defaultGenLength": cfg.DefaultGenLength,
		"configPath":       cfg.ConfigPath,
		"vaultExists":      fileExists(cfg.VaultPath),
	})
}

func (s *Server) handleConfigSet(w http.ResponseWriter, r *http.Request) {
	var req configRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	normalized, err := config.Set(req.Key, req.Value)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	sessionInvalidated := false
	if req.Key == "vault_path" {
		// Session is bound to the old path — force re-unlock.
		s.sessions.clear()
		sessionInvalidated = true
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"key": req.Key, "value": normalized, "sessionInvalidated": sessionInvalidated,
	})
}
