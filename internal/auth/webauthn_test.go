package auth

import (
	"context"
	"testing"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
)

func TestNewWebAuthnService(t *testing.T) {
	repo := storage.NewMockRepository()
	config := WebAuthnConfig{
		RPDisplayName: "Test RP",
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
		t.Fatal("Expected non-nil webAuthn")
	}

	if service.repo == nil {
		t.Fatal("Expected non-nil repo")
	}

	if service.sessions == nil {
		t.Fatal("Expected non-nil sessions map")
	}
}

func TestCredentialIDConversion(t *testing.T) {
	// Test CredentialIDToString
	credentialID := []byte{1, 2, 3, 4, 5}
	str := CredentialIDToString(credentialID)
	if str == "" {
		t.Fatal("Expected non-empty string")
	}

	// Test StringToCredentialID
	decodedID, err := StringToCredentialID(str)
	if err != nil {
		t.Fatalf("Failed to decode credential ID: %v", err)
	}

	if string(decodedID) != string(credentialID) {
		t.Fatalf("Expected decoded ID to be %v, got %v", credentialID, decodedID)
	}

	// Test with invalid base64
	_, err = StringToCredentialID("invalid-base64!@#$")
	if err == nil {
		t.Fatal("Expected error for invalid base64")
	}
}

func TestWebAuthnConfig(t *testing.T) {
	// Test with valid config
	config := WebAuthnConfig{
		RPDisplayName: "Test RP",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}

	// Validate the config
	if config.RPDisplayName != "Test RP" {
		t.Errorf("Expected RPDisplayName to be 'Test RP', got '%s'", config.RPDisplayName)
	}
	if config.RPID != "localhost" {
		t.Errorf("Expected RPID to be 'localhost', got '%s'", config.RPID)
	}
	if config.RPOrigin != "http://localhost:8080" {
		t.Errorf("Expected RPOrigin to be 'http://localhost:8080', got '%s'", config.RPOrigin)
	}
}

func TestGetUserCredentials(t *testing.T) {
	// Create a mock repository with some passkeys
	repo := storage.NewMockRepository()

	// Add passkeys to the mock repository
	repo.Passkeys = []*models.Passkey{
		{
			ID:              "passkey1",
			UserID:          "user1",
			CredentialID:    []byte{1, 2, 3},
			PublicKey:       []byte{4, 5, 6},
			AAGUID:          []byte{7, 8, 9},
			SignCount:       1,
			Name:            "Passkey 1",
			CreatedAt:       time.Now(),
			LastUsedAt:      time.Now(),
			Transports:      []string{"internal"},
			AttestationType: "none",
		},
		{
			ID:              "passkey2",
			UserID:          "user1",
			CredentialID:    []byte{10, 11, 12},
			PublicKey:       []byte{13, 14, 15},
			AAGUID:          []byte{16, 17, 18},
			SignCount:       2,
			Name:            "Passkey 2",
			CreatedAt:       time.Now(),
			LastUsedAt:      time.Now(),
			Transports:      []string{"usb"},
			AttestationType: "none",
		},
		{
			ID:              "passkey3",
			UserID:          "user2",
			CredentialID:    []byte{19, 20, 21},
			PublicKey:       []byte{22, 23, 24},
			AAGUID:          []byte{25, 26, 27},
			SignCount:       3,
			Name:            "Passkey 3",
			CreatedAt:       time.Now(),
			LastUsedAt:      time.Now(),
			Transports:      []string{"nfc"},
			AttestationType: "none",
		},
	}

	// Create a WebAuthnService
	config := WebAuthnConfig{
		RPDisplayName: "Test RP",
		RPID:          "localhost",
		RPOrigin:      "http://localhost:8080",
	}

	service, err := NewWebAuthnService(config, repo)
	if err != nil {
		t.Fatalf("Failed to create WebAuthnService: %v", err)
	}

	// Test getUserCredentials for user1
	user1 := &models.User{ID: "user1"}
	credentials, err := service.getUserCredentials(context.Background(), user1)
	if err != nil {
		t.Fatalf("Failed to get user credentials: %v", err)
	}

	if len(credentials) != 2 {
		t.Fatalf("Expected 2 credentials, got %d", len(credentials))
	}

	// Check the first credential
	if string(credentials[0].ID) != string(repo.Passkeys[0].CredentialID) {
		t.Errorf("Expected credential ID %v, got %v", repo.Passkeys[0].CredentialID, credentials[0].ID)
	}
	if string(credentials[0].PublicKey) != string(repo.Passkeys[0].PublicKey) {
		t.Errorf("Expected public key %v, got %v", repo.Passkeys[0].PublicKey, credentials[0].PublicKey)
	}
	if string(credentials[0].Authenticator.AAGUID) != string(repo.Passkeys[0].AAGUID) {
		t.Errorf("Expected AAGUID %v, got %v", repo.Passkeys[0].AAGUID, credentials[0].Authenticator.AAGUID)
	}
	if credentials[0].Authenticator.SignCount != repo.Passkeys[0].SignCount {
		t.Errorf("Expected sign count %d, got %d", repo.Passkeys[0].SignCount, credentials[0].Authenticator.SignCount)
	}
	if credentials[0].AttestationType != repo.Passkeys[0].AttestationType {
		t.Errorf("Expected attestation type %s, got %s", repo.Passkeys[0].AttestationType, credentials[0].AttestationType)
	}
	if len(credentials[0].Transport) != len(repo.Passkeys[0].Transports) {
		t.Errorf("Expected %d transports, got %d", len(repo.Passkeys[0].Transports), len(credentials[0].Transport))
	} else if string(credentials[0].Transport[0]) != repo.Passkeys[0].Transports[0] {
		t.Errorf("Expected transport %s, got %s", repo.Passkeys[0].Transports[0], credentials[0].Transport[0])
	}

	// Test getUserCredentials for user2
	user2 := &models.User{ID: "user2"}
	credentials, err = service.getUserCredentials(context.Background(), user2)
	if err != nil {
		t.Fatalf("Failed to get user credentials: %v", err)
	}

	if len(credentials) != 1 {
		t.Fatalf("Expected 1 credential, got %d", len(credentials))
	}

	// Test getUserCredentials for non-existent user
	user3 := &models.User{ID: "user3"}
	credentials, err = service.getUserCredentials(context.Background(), user3)
	if err != nil {
		t.Fatalf("Failed to get user credentials: %v", err)
	}

	if len(credentials) != 0 {
		t.Fatalf("Expected 0 credentials, got %d", len(credentials))
	}
}
