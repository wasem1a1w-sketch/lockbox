package storage

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

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