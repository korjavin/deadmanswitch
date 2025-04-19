package auth

import (
	"context"
	"testing"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
)

// MockRepository is a mock implementation of the storage.Repository interface
type MockRepository struct {
	passkeys []*models.Passkey
}

func (m *MockRepository) ListPasskeysByUserID(ctx context.Context, userID string) ([]*models.Passkey, error) {
	var result []*models.Passkey
	for _, p := range m.passkeys {
		if p.UserID == userID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (m *MockRepository) ListPasskeys(ctx context.Context) ([]*models.Passkey, error) {
	return m.passkeys, nil
}

func (m *MockRepository) GetPasskeyByCredentialID(ctx context.Context, credentialID []byte) (*models.Passkey, error) {
	for _, p := range m.passkeys {
		if string(p.CredentialID) == string(credentialID) {
			return p, nil
		}
	}
	return nil, storage.ErrNotFound
}

func (m *MockRepository) CreatePasskey(ctx context.Context, passkey *models.Passkey) error {
	m.passkeys = append(m.passkeys, passkey)
	return nil
}

func (m *MockRepository) UpdatePasskey(ctx context.Context, passkey *models.Passkey) error {
	for i, p := range m.passkeys {
		if p.ID == passkey.ID {
			m.passkeys[i] = passkey
			return nil
		}
	}
	return storage.ErrNotFound
}

// Implement other methods of the Repository interface with empty implementations
func (m *MockRepository) CreateUser(ctx context.Context, user *models.User) error { return nil }
func (m *MockRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return nil, nil
}
func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}
func (m *MockRepository) GetUserByTelegramID(ctx context.Context, telegramID string) (*models.User, error) {
	return nil, nil
}
func (m *MockRepository) UpdateUser(ctx context.Context, user *models.User) error       { return nil }
func (m *MockRepository) DeleteUser(ctx context.Context, id string) error               { return nil }
func (m *MockRepository) ListUsers(ctx context.Context) ([]*models.User, error)         { return nil, nil }
func (m *MockRepository) CreateSecret(ctx context.Context, secret *models.Secret) error { return nil }
func (m *MockRepository) GetSecretByID(ctx context.Context, id string) (*models.Secret, error) {
	return nil, nil
}
func (m *MockRepository) ListSecretsByUserID(ctx context.Context, userID string) ([]*models.Secret, error) {
	return nil, nil
}
func (m *MockRepository) UpdateSecret(ctx context.Context, secret *models.Secret) error { return nil }
func (m *MockRepository) DeleteSecret(ctx context.Context, id string) error             { return nil }
func (m *MockRepository) CreateRecipient(ctx context.Context, recipient *models.Recipient) error {
	return nil
}
func (m *MockRepository) GetRecipientByID(ctx context.Context, id string) (*models.Recipient, error) {
	return nil, nil
}
func (m *MockRepository) ListRecipientsByUserID(ctx context.Context, userID string) ([]*models.Recipient, error) {
	return nil, nil
}
func (m *MockRepository) UpdateRecipient(ctx context.Context, recipient *models.Recipient) error {
	return nil
}
func (m *MockRepository) DeleteRecipient(ctx context.Context, id string) error { return nil }
func (m *MockRepository) CreateSecretAssignment(ctx context.Context, assignment *models.SecretAssignment) error {
	return nil
}
func (m *MockRepository) GetSecretAssignmentByID(ctx context.Context, id string) (*models.SecretAssignment, error) {
	return nil, nil
}
func (m *MockRepository) ListSecretAssignmentsBySecretID(ctx context.Context, secretID string) ([]*models.SecretAssignment, error) {
	return nil, nil
}
func (m *MockRepository) ListSecretAssignmentsByRecipientID(ctx context.Context, recipientID string) ([]*models.SecretAssignment, error) {
	return nil, nil
}
func (m *MockRepository) ListSecretAssignmentsByUserID(ctx context.Context, userID string) ([]*models.SecretAssignment, error) {
	return nil, nil
}
func (m *MockRepository) DeleteSecretAssignment(ctx context.Context, id string) error { return nil }
func (m *MockRepository) CreatePingHistory(ctx context.Context, ping *models.PingHistory) error {
	return nil
}
func (m *MockRepository) UpdatePingHistory(ctx context.Context, ping *models.PingHistory) error {
	return nil
}
func (m *MockRepository) GetLatestPingByUserID(ctx context.Context, userID string) (*models.PingHistory, error) {
	return nil, nil
}
func (m *MockRepository) ListPingHistoryByUserID(ctx context.Context, userID string) ([]*models.PingHistory, error) {
	return nil, nil
}
func (m *MockRepository) CreatePingVerification(ctx context.Context, verification *models.PingVerification) error {
	return nil
}
func (m *MockRepository) GetPingVerificationByCode(ctx context.Context, code string) (*models.PingVerification, error) {
	return nil, nil
}
func (m *MockRepository) UpdatePingVerification(ctx context.Context, verification *models.PingVerification) error {
	return nil
}
func (m *MockRepository) CreateDeliveryEvent(ctx context.Context, event *models.DeliveryEvent) error {
	return nil
}
func (m *MockRepository) ListDeliveryEventsByUserID(ctx context.Context, userID string) ([]*models.DeliveryEvent, error) {
	return nil, nil
}
func (m *MockRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error { return nil }
func (m *MockRepository) ListAuditLogsByUserID(ctx context.Context, userID string) ([]*models.AuditLog, error) {
	return nil, nil
}
func (m *MockRepository) CreateSession(ctx context.Context, session *models.Session) error {
	return nil
}
func (m *MockRepository) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	return nil, nil
}
func (m *MockRepository) DeleteSession(ctx context.Context, id string) error         { return nil }
func (m *MockRepository) DeleteExpiredSessions(ctx context.Context) error            { return nil }
func (m *MockRepository) UpdateSessionActivity(ctx context.Context, id string) error { return nil }
func (m *MockRepository) GetUsersForPinging(ctx context.Context) ([]*models.User, error) {
	return nil, nil
}
func (m *MockRepository) GetUsersWithExpiredPings(ctx context.Context) ([]*models.User, error) {
	return nil, nil
}
func (m *MockRepository) BeginTx(ctx context.Context) (storage.Transaction, error) { return nil, nil }
func (m *MockRepository) GetPasskeyByID(ctx context.Context, id string) (*models.Passkey, error) {
	return nil, nil
}
func (m *MockRepository) DeletePasskey(ctx context.Context, id string) error              { return nil }
func (m *MockRepository) DeletePasskeysByUserID(ctx context.Context, userID string) error { return nil }

func TestNewWebAuthnService(t *testing.T) {
	repo := &MockRepository{}
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
	repo := &MockRepository{
		passkeys: []*models.Passkey{
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
	if string(credentials[0].ID) != string(repo.passkeys[0].CredentialID) {
		t.Errorf("Expected credential ID %v, got %v", repo.passkeys[0].CredentialID, credentials[0].ID)
	}
	if string(credentials[0].PublicKey) != string(repo.passkeys[0].PublicKey) {
		t.Errorf("Expected public key %v, got %v", repo.passkeys[0].PublicKey, credentials[0].PublicKey)
	}
	if string(credentials[0].Authenticator.AAGUID) != string(repo.passkeys[0].AAGUID) {
		t.Errorf("Expected AAGUID %v, got %v", repo.passkeys[0].AAGUID, credentials[0].Authenticator.AAGUID)
	}
	if credentials[0].Authenticator.SignCount != repo.passkeys[0].SignCount {
		t.Errorf("Expected sign count %d, got %d", repo.passkeys[0].SignCount, credentials[0].Authenticator.SignCount)
	}
	if credentials[0].AttestationType != repo.passkeys[0].AttestationType {
		t.Errorf("Expected attestation type %s, got %s", repo.passkeys[0].AttestationType, credentials[0].AttestationType)
	}
	if len(credentials[0].Transport) != len(repo.passkeys[0].Transports) {
		t.Errorf("Expected %d transports, got %d", len(repo.passkeys[0].Transports), len(credentials[0].Transport))
	} else if string(credentials[0].Transport[0]) != repo.passkeys[0].Transports[0] {
		t.Errorf("Expected transport %s, got %s", repo.passkeys[0].Transports[0], credentials[0].Transport[0])
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
