package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"os"
)

type SecureStorage struct {
	FileName string
}

func NewSecureStorage(fileName string) *SecureStorage {
	return &SecureStorage{FileName: fileName}
}

// Save encrypts the vault payload and commits it to disk with a salt header prefix
func (s *SecureStorage) Save(vault Vault, key []byte, salt []byte) error {
	plaintext, err := json.Marshal(vault)
	if err != nil {
		return err
	}

	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		return err
	}

	// Layout layout: [16-byte Salt Header] + [Encrypted Data Body]
	fileBytes := append(salt, ciphertext...)

	// Write with 0600 permissions so ONLY your user account can read/write the file
	return os.WriteFile(s.FileName, fileBytes, 0600)
}

// Load reads the vault file, separates the salt header, and attempts to decrypt the data
func (s *SecureStorage) Load(key []byte) (Vault, error) {
	fileBytes, err := os.ReadFile(s.FileName)
	if err != nil {
		if os.IsNotExist(err) {
			return Vault{}, nil // Return an empty vault if the file hasn't been created yet
		}
		return nil, err
	}

	// If the file is smaller than 16 bytes, it doesn't even contain a valid salt header
	if len(fileBytes) < 16 {
		return nil, errors.New("malformed or corrupted vault file")
	}

	// Extract the ciphertext body by slicing past the first 16 bytes of salt
	ciphertext := fileBytes[16:]

	plaintext, err := Decrypt(ciphertext, key)
	if err != nil {
		return nil, errors.New("access denied: incorrect master password")
	}

	var vault Vault
	err = json.Unmarshal(plaintext, &vault)
	return vault, err
}

// GetOrGenerateSalt reads the existing 16-byte salt header or generates a brand new one
func (s *SecureStorage) GetOrGenerateSalt() ([]byte, error) {
	file, err := os.Open(s.FileName)
	if err == nil {
		defer file.Close()
		salt := make([]byte, 16)
		_, err := io.ReadAtLeast(file, salt, 16)
		return salt, err
	}

	// FIXED: Use io.ReadFull to populate the salt byte slice using the rand.Reader stream
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	return salt, nil
}
