package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID generates a random ID for database entities
func GenerateID() string {
	// Generate a random 16-byte ID
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(idBytes)
}

// GenerateSalt generates a random salt for cryptographic operations
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}
