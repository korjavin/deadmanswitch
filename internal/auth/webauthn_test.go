package auth

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
)

// TestNewWebAuthnService tests creating a new WebAuthnService
func TestNewWebAuthnService(t *testing.T) {
	repo := storage.NewMockRepository()

	config := WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}

	service, err := NewWebAuthnService(config, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	if service == nil {
		t.Fatal("Expected non-nil WebAuthnService")
	}

	if service.webAuthn == nil {
		t.Error("Expected non-nil webAuthn in service")
	}

	if service.repo != repo {
		t.Error("Expected repo in service to match input repo")
	}

	if service.sessions == nil {
		t.Error("Expected non-nil sessions map in service")
	}
}

// TestCredentialIDConversions tests the credential ID conversion functions
func TestCredentialIDConversions(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"Empty", []byte{}},
		{"Simple", []byte("test-credential")},
		{"Complex", []byte{0, 1, 2, 3, 4, 5, 255, 254, 253}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Convert to string
			str := CredentialIDToString(test.input)

			// Convert back to byte slice
			result, err := StringToCredentialID(str)
			if err != nil {
				t.Fatalf("StringToCredentialID failed: %v", err)
			}

			// Check that the result matches the input
			if len(result) != len(test.input) {
				t.Errorf("Expected result length %d, got %d", len(test.input), len(result))
			}

			match := true
			for i, b := range test.input {
				if i >= len(result) || result[i] != b {
					match = false
					break
				}
			}

			if !match {
				t.Errorf("Result does not match input. Input: %v, Result: %v", test.input, result)
			}
		})
	}
}

// TestStringToCredentialIDError tests error cases for StringToCredentialID
func TestStringToCredentialIDError(t *testing.T) {
	// Invalid base64 strings
	invalidInputs := []string{
		"!@#$%", // Not base64
		"a===",  // Invalid padding
	}

	for _, input := range invalidInputs {
		_, err := StringToCredentialID(input)
		if err == nil {
			t.Errorf("Expected error for invalid input: %s", input)
		}
	}
}

// TestGetUserCredentials tests the getUserCredentials method
func TestGetUserCredentials(t *testing.T) {
	// Create mock repository
	repo := storage.NewMockRepository()

	// Create a WebAuthnService
	config := WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}

	service, err := NewWebAuthnService(config, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	// Create a test user
	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
	}

	// Test with no passkeys
	ctx := context.Background()
	credentials, err := service.getUserCredentials(ctx, user)
	if err != nil {
		t.Fatalf("getUserCredentials failed with no passkeys: %v", err)
	}
	if len(credentials) != 0 {
		t.Errorf("Expected 0 credentials with no passkeys, got %d", len(credentials))
	}

	// Add some passkeys to the repository
	passkey1 := &models.Passkey{
		ID:              "passkey1",
		UserID:          user.ID,
		CredentialID:    []byte("credential1"),
		PublicKey:       []byte("publickey1"),
		AAGUID:          []byte("aaguid1"),
		SignCount:       1,
		AttestationType: "none",
		Transports:      []string{"internal", "usb"},
		CreatedAt:       time.Now(),
		LastUsedAt:      time.Now(),
	}

	passkey2 := &models.Passkey{
		ID:              "passkey2",
		UserID:          user.ID,
		CredentialID:    []byte("credential2"),
		PublicKey:       []byte("publickey2"),
		AAGUID:          []byte("aaguid2"),
		SignCount:       2,
		AttestationType: "direct",
		Transports:      []string{"ble"},
		CreatedAt:       time.Now(),
		LastUsedAt:      time.Now(),
	}

	repo.Passkeys = append(repo.Passkeys, passkey1, passkey2)

	// Test with passkeys
	credentials, err = service.getUserCredentials(ctx, user)
	if err != nil {
		t.Fatalf("getUserCredentials failed with passkeys: %v", err)
	}
	if len(credentials) != 2 {
		t.Errorf("Expected 2 credentials with passkeys, got %d", len(credentials))
	}

	// Check the first credential
	if !byteSliceEqual(credentials[0].ID, passkey1.CredentialID) {
		t.Errorf("Expected credential ID to match passkey1")
	}
	if !byteSliceEqual(credentials[0].PublicKey, passkey1.PublicKey) {
		t.Errorf("Expected public key to match passkey1")
	}
	if credentials[0].AttestationType != passkey1.AttestationType {
		t.Errorf("Expected attestation type to match passkey1")
	}
	if len(credentials[0].Transport) != len(passkey1.Transports) {
		t.Errorf("Expected transport count to match passkey1")
	}
	if !byteSliceEqual(credentials[0].Authenticator.AAGUID, passkey1.AAGUID) {
		t.Errorf("Expected AAGUID to match passkey1")
	}
	if credentials[0].Authenticator.SignCount != passkey1.SignCount {
		t.Errorf("Expected sign count to match passkey1")
	}
}

// Helper function to compare byte slices
func byteSliceEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// TestWebAuthnSessionHandling tests session creation and cookie setting
func TestWebAuthnSessionHandling(t *testing.T) {
	// Create mock repository
	repo := storage.NewMockRepository()

	// Create a WebAuthnService
	config := WebAuthnConfig{
		RPDisplayName: "Test Service",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}

	service, err := NewWebAuthnService(config, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	// Create a test response writer
	rw := httptest.NewRecorder()

	// Create a session ID and store a mock session
	sessionID := "test-session"
	sessionData := &webauthn.SessionData{
		Challenge:            base64.RawURLEncoding.EncodeToString([]byte("test-challenge")),
		UserID:               []byte("user123"),
		AllowedCredentialIDs: [][]byte{},
	}

	// Store the session
	service.mutex.Lock()
	service.sessions[sessionID] = sessionData
	service.mutex.Unlock()

	// Set a cookie with the session ID
	http.SetCookie(rw, &http.Cookie{
		Name:     "webauthn_session_id",
		Value:    sessionID,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   false,
	})

	// Check that the cookie was set
	cookies := rw.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == "webauthn_session_id" {
			found = true
			if cookie.Value != sessionID {
				t.Errorf("Expected cookie value %s, got %s", sessionID, cookie.Value)
			}
		}
	}

	if !found {
		t.Error("webauthn_session_id cookie not found")
	}
}
