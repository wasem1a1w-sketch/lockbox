package storage

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"lockbox/internal/crypto"
	"lockbox/internal/vault"
)

type SecureStorage struct {
	FileName string
}

func NewSecureStorage(fileName string) *SecureStorage {
	return &SecureStorage{FileName: fileName}
}

func (s *SecureStorage) Save(v vault.Vault, key, salt []byte, iterations int) error {
	plaintext, err := json.Marshal(v)
	if err != nil {
		return err
	}

	ciphertext, err := crypto.Encrypt(plaintext, key)
	if err != nil {
		return err
	}

	iterBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(iterBytes, uint32(iterations))

	fileBytes := append(salt, iterBytes...)
	fileBytes = append(fileBytes, ciphertext...)

	if err := os.MkdirAll(filepath.Dir(s.FileName), 0700); err != nil {
		return err
	}

	return os.WriteFile(s.FileName, fileBytes, 0600)
}

func (s *SecureStorage) Load(key []byte) (vault.Vault, int, error) {
	fileBytes, err := os.ReadFile(s.FileName)
	if err != nil {
		if os.IsNotExist(err) {
			return vault.Vault{}, 0, nil
		}
		return nil, 0, err
	}

	if len(fileBytes) < crypto.SaltSize {
		return nil, 0, errors.New("malformed or corrupted vault file")
	}

	var iterations int
	var ciphertext []byte

	if len(fileBytes) >= crypto.SaltSize+4 {
		iterations = int(binary.BigEndian.Uint32(fileBytes[crypto.SaltSize : crypto.SaltSize+4]))
		ciphertext = fileBytes[crypto.SaltSize+4:]
	} else {
		iterations = 100000
		ciphertext = fileBytes[crypto.SaltSize:]
	}

	if iterations < 1000 || iterations > 10000000 {
		iterations = 100000
		ciphertext = fileBytes[crypto.SaltSize:]
	}

	plaintext, err := crypto.Decrypt(ciphertext, key)
	if err != nil {
		return nil, 0, errors.New("access denied: incorrect master password")
	}

	var v vault.Vault
	err = json.Unmarshal(plaintext, &v)
	return v, iterations, err
}

func (s *SecureStorage) LoadWithIterations(password string, salt []byte, iterations int) (vault.Vault, int, error) {
	key := crypto.DeriveKey(password, salt, iterations)
	return s.Load(key)
}

func (s *SecureStorage) GetOrGenerateSalt() ([]byte, error) {
	file, err := os.Open(s.FileName)
	if err == nil {
		defer file.Close()
		salt := make([]byte, crypto.SaltSize)
		_, err := io.ReadAtLeast(file, salt, crypto.SaltSize)
		return salt, err
	}

	return crypto.GenerateSalt()
}

func (s *SecureStorage) GetIterations() (int, error) {
	file, err := os.Open(s.FileName)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	header := make([]byte, crypto.SaltSize+4)
	_, err = io.ReadAtLeast(file, header, crypto.SaltSize+4)
	if err != nil {
		return 0, err
	}

	iterations := int(binary.BigEndian.Uint32(header[crypto.SaltSize:]))
	if iterations < 1000 || iterations > 10000000 {
		return 100000, nil
	}
	return iterations, nil
}

// layout describes one known vault file format.
// legacy: [16B salt][ciphertext]                     (pre-refactor CLI, fixed 100000 iterations)
// current: [16B salt][4B iterations][ciphertext]
type layout struct {
	iterations int
	ctOffset   int
}

// LoadWithPassword decrypts the vault, trying every known file layout until
// one authenticates. GCM authentication guarantees no false positives.
// Returns the vault and the iterations actually used, so a save rewrites
// the file in the current format with a correct header.
func (s *SecureStorage) LoadWithPassword(password string, explicitIterations int) (vault.Vault, int, error) {
	fileBytes, err := os.ReadFile(s.FileName)
	if err != nil {
		if os.IsNotExist(err) {
			return vault.Vault{}, 0, nil
		}
		return nil, 0, err
	}
	if len(fileBytes) < crypto.SaltSize {
		return nil, 0, errors.New("malformed or corrupted vault file")
	}
	salt := fileBytes[:crypto.SaltSize]

	for _, l := range detectLayouts(fileBytes, explicitIterations) {
		if l.ctOffset >= len(fileBytes) {
			continue
		}
		key := crypto.DeriveKey(password, salt, l.iterations)
		plaintext, err := crypto.Decrypt(fileBytes[l.ctOffset:], key)
		if err != nil {
			continue
		}
		v, err := decodeVault(plaintext)
		if err != nil {
			continue
		}
		return v, l.iterations, nil
	}
	return nil, 0, errors.New("access denied: incorrect master password")
}

// legacyCredential matches both current and pre-refactor JSON shapes.
// The old CLI stored a positional `index` field that the current model
// lacks; decodeVault maps it to SortOrder so display order survives.
type legacyCredential struct {
	ID        string    `json:"id"`
	Index     int       `json:"index"`
	Account   string    `json:"account"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	SavedAt   time.Time `json:"saved_at"`
	SortOrder int       `json:"sort_order"`
}

func decodeVault(plaintext []byte) (vault.Vault, error) {
	var raw []legacyCredential
	if err := json.Unmarshal(plaintext, &raw); err != nil {
		return nil, err
	}
	v := make(vault.Vault, len(raw))
	for i, r := range raw {
		sortOrder := r.SortOrder
		if sortOrder == 0 && r.Index > 0 {
			sortOrder = r.Index
		}
		v[i] = vault.Credential{
			ID:        r.ID,
			Account:   r.Account,
			Username:  r.Username,
			Password:  r.Password,
			SavedAt:   r.SavedAt,
			SortOrder: sortOrder,
		}
	}
	return v, nil
}

func detectLayouts(fileBytes []byte, explicitIterations int) []layout {
	if explicitIterations > 0 {
		return []layout{
			{explicitIterations, crypto.SaltSize + 4},
			{explicitIterations, crypto.SaltSize},
		}
	}
	var out []layout
	if len(fileBytes) >= crypto.SaltSize+4 {
		headerIterations := int(binary.BigEndian.Uint32(fileBytes[crypto.SaltSize : crypto.SaltSize+4]))
		if headerIterations >= 1000 && headerIterations <= 10000000 {
			out = append(out, layout{headerIterations, crypto.SaltSize + 4})
		}
	}
	// Legacy CLI always derived with 100000 iterations.
	out = append(out, layout{crypto.DefaultIterations, crypto.SaltSize})
	return out
}
