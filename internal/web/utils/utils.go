package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
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

// GenerateSecureToken generates a secure random token for sessions
func GenerateSecureToken() string {
	// Generate a random 32-byte token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(tokenBytes)
}

// VerifyPassword verifies a password against a bcrypt hashed password
func VerifyPassword(hashedPassword []byte, password string) bool {
	// Use bcrypt to compare the password with the hash
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	return err == nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// DetermineContactMethod determines the contact method for a recipient
// based on available information
func DetermineContactMethod(telegramUsername, _ string) string {
	if telegramUsername != "" {
		return "telegram"
	}
	return "email"
}
