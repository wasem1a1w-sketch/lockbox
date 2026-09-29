package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"lockbox/internal/crypto"
	"lockbox/internal/storage"
	"lockbox/internal/vault"
)

type session struct {
	storage    *storage.SecureStorage
	key        []byte
	salt       []byte
	vault      vault.Vault
	iterations int
}

func (s *session) save() error {
	return s.storage.Save(s.vault, s.key, s.salt, s.iterations)
}

// unlockVault decrypts the vault, prompting for the master password.
// allowCreate=true starts the create-new-vault flow when the file is absent.
// allowCreate=false returns an empty session without prompting (read-only cmds).
func unlockVault(allowCreate bool) (*session, error) {
	s := storage.NewSecureStorage(cfg.VaultPath)
	_, statErr := os.Stat(cfg.VaultPath)
	vaultExists := statErr == nil

	if !vaultExists && !allowCreate {
		return &session{storage: s, iterations: cfg.KDFIterations}, nil
	}

	salt, err := s.GetOrGenerateSalt()
	if err != nil {
		return nil, fmt.Errorf("🚨 Initialization error: %w", err)
	}

	if !vaultExists {
		pw, err := readNewMasterPassword()
		if err != nil {
			return nil, err
		}
		key := crypto.DeriveKey(pw, salt, cfg.KDFIterations)
		fmt.Println("🔓 New vault created.")
		return &session{storage: s, key: key, salt: salt, iterations: cfg.KDFIterations}, nil
	}

	pw, err := readMasterPassword("🔑 Enter Vault Master Password: ")
	if err != nil {
		return nil, err
	}
	if pw == "" {
		return nil, fmt.Errorf("🚨 Master password cannot be blank.")
	}
	explicit := 0
	if iterationsExplicit {
		explicit = cfg.KDFIterations
	}
	v, iterations, err := s.LoadWithPassword(pw, explicit)
	if err != nil {
		return nil, fmt.Errorf("❌ Authentication Failed: %w", err)
	}
	sortByOrder(v)
	key := crypto.DeriveKey(pw, salt, iterations)
	fmt.Println("🔓 Vault unlocked successfully.")
	return &session{storage: s, key: key, salt: salt, vault: v, iterations: iterations}, nil
}

func sortByOrder(v vault.Vault) {
	sort.SliceStable(v, func(i, j int) bool {
		return v[i].SortOrder < v[j].SortOrder
	})
}

// resolveID accepts a display index (1-based), a full UUID, or a UUID prefix.
// Prefix match must be unambiguous.
func resolveID(v vault.Vault, ref string) *vault.Credential {
	if n, err := strconv.Atoi(ref); err == nil {
		if c := v.FindByIndex(n); c != nil {
			return c
		}
	}
	if c := v.FindByID(ref); c != nil {
		return c
	}
	var match *vault.Credential
	for i := range v {
		if strings.HasPrefix(v[i].ID, ref) {
			if match != nil {
				return nil // ambiguous prefix
			}
			match = &v[i]
		}
	}
	return match
}

// displayPosition returns the 1-based display index of a credential ID.
func displayPosition(v vault.Vault, id string) int {
	for i := range v {
		if v[i].ID == id {
			return i + 1
		}
	}
	return 0
}
