package utils

import (
	"encoding/base64"
	"encoding/hex"
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	// Generate an ID
	id := GenerateID()

	// Check that the ID is not empty
	if id == "" {
		t.Fatal("Expected non-empty ID")
	}

	// Check that the ID is a valid hex string
	if len(id) != 32 {
		t.Errorf("Expected ID length to be 32, got %d", len(id))
	}

	// Verify it's a valid hex string
	_, err := hex.DecodeString(id)
	if err != nil {
		t.Errorf("ID is not a valid hex string: %v", err)
	}

	// Generate another ID and check that they're different
	id2 := GenerateID()
	if id == id2 {
		t.Error("Expected different IDs, got the same ID twice")
	}
}

func TestGenerateSecureToken(t *testing.T) {
	// Generate a token
	token := GenerateSecureToken()

	// Check that the token is not empty
	if token == "" {
		t.Fatal("Expected non-empty token")
	}

	// Check that the token is a valid base64 URL-encoded string
	// Base64 URL-encoded strings should only contain alphanumeric characters, '-', '_', and possibly '='
	for _, c := range token {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '=') {
			t.Errorf("Token contains invalid character: %c", c)
			break
		}
	}

	// Try to decode the token
	_, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		t.Errorf("Token is not a valid base64 URL-encoded string: %v", err)
	}

	// Generate another token and check that they're different
	token2 := GenerateSecureToken()
	if token == token2 {
		t.Error("Expected different tokens, got the same token twice")
	}
}

func TestVerifyPassword(t *testing.T) {
	// Hash a password
	password := "test-password"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Verify the correct password
	if !VerifyPassword(hashedPassword, password) {
		t.Error("Expected password verification to succeed with correct password")
	}

	// Verify an incorrect password
	if VerifyPassword(hashedPassword, "wrong-password") {
		t.Error("Expected password verification to fail with incorrect password")
	}

	// Verify with invalid hash
	if VerifyPassword([]byte("invalid-hash"), password) {
		t.Error("Expected password verification to fail with invalid hash")
	}
}

func TestHashPassword(t *testing.T) {
	// Hash a password
	password := "test-password"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Check that the hash is not empty
	if len(hashedPassword) == 0 {
		t.Fatal("Expected non-empty hash")
	}

	// Check that the hash starts with the bcrypt identifier
	if !strings.HasPrefix(string(hashedPassword), "$2a$") {
		t.Errorf("Expected hash to start with bcrypt identifier, got %s", string(hashedPassword))
	}

	// Hash the same password again and check that the hashes are different (due to salt)
	hashedPassword2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	if string(hashedPassword) == string(hashedPassword2) {
		t.Error("Expected different hashes for the same password, got the same hash twice")
	}
}

func TestDetermineContactMethod(t *testing.T) {
	// Test with telegram username
	method := DetermineContactMethod("telegram_user", "user@example.com")
	if method != "telegram" {
		t.Errorf("Expected method to be 'telegram', got '%s'", method)
	}

	// Test with empty telegram username
	method = DetermineContactMethod("", "user@example.com")
	if method != "email" {
		t.Errorf("Expected method to be 'email', got '%s'", method)
	}

	// Test with both empty
	method = DetermineContactMethod("", "")
	if method != "email" {
		t.Errorf("Expected method to be 'email', got '%s'", method)
	}
}
