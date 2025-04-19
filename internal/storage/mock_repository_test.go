package storage

import (
	"context"
	"testing"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
)

func TestNewMockRepository(t *testing.T) {
	repo := NewMockRepository()
	if repo == nil {
		t.Fatal("Expected non-nil repository")
	}

	// Check that all slices are initialized
	if repo.Users == nil {
		t.Error("Expected Users slice to be initialized")
	}
	if repo.Recipients == nil {
		t.Error("Expected Recipients slice to be initialized")
	}
	if repo.Secrets == nil {
		t.Error("Expected Secrets slice to be initialized")
	}
	if repo.SecretAssignments == nil {
		t.Error("Expected SecretAssignments slice to be initialized")
	}
	if repo.PingHistories == nil {
		t.Error("Expected PingHistories slice to be initialized")
	}
	if repo.PingVerifications == nil {
		t.Error("Expected PingVerifications slice to be initialized")
	}
	if repo.DeliveryEvents == nil {
		t.Error("Expected DeliveryEvents slice to be initialized")
	}
	if repo.Sessions == nil {
		t.Error("Expected Sessions slice to be initialized")
	}
	if repo.UsersForPinging == nil {
		t.Error("Expected UsersForPinging slice to be initialized")
	}
	if repo.UsersWithExpiredPings == nil {
		t.Error("Expected UsersWithExpiredPings slice to be initialized")
	}
}

func TestMockRepositoryUserMethods(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Test CreateUser
	user := &models.User{
		ID:    "user1",
		Email: "user1@example.com",
	}
	err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// Test GetUserByID
	retrievedUser, err := repo.GetUserByID(ctx, "user1")
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if retrievedUser == nil {
		t.Fatal("Expected non-nil user")
	}
	if retrievedUser.ID != "user1" {
		t.Errorf("Expected user ID to be 'user1', got '%s'", retrievedUser.ID)
	}

	// Test GetUserByEmail
	retrievedUser, err = repo.GetUserByEmail(ctx, "user1@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}
	if retrievedUser == nil {
		t.Fatal("Expected non-nil user")
	}
	if retrievedUser.Email != "user1@example.com" {
		t.Errorf("Expected user email to be 'user1@example.com', got '%s'", retrievedUser.Email)
	}

	// Test UpdateUser
	user.Email = "updated@example.com"
	err = repo.UpdateUser(ctx, user)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	// Verify the update
	retrievedUser, err = repo.GetUserByID(ctx, "user1")
	if err != nil {
		t.Fatalf("GetUserByID failed after update: %v", err)
	}
	if retrievedUser.Email != "updated@example.com" {
		t.Errorf("Expected updated email to be 'updated@example.com', got '%s'", retrievedUser.Email)
	}

	// Test ListUsers
	users, err := repo.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	// Test DeleteUser
	err = repo.DeleteUser(ctx, "user1")
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	// Verify the deletion
	users, err = repo.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers failed after deletion: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("Expected 0 users after deletion, got %d", len(users))
	}
}

func TestMockRepositorySecretMethods(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Test CreateSecret
	secret := &models.Secret{
		ID:            "secret1",
		UserID:        "user1",
		Name:          "Test Secret",
		EncryptedData: "encrypted-data",
	}
	err := repo.CreateSecret(ctx, secret)
	if err != nil {
		t.Fatalf("CreateSecret failed: %v", err)
	}

	// Test GetSecretByID
	retrievedSecret, err := repo.GetSecretByID(ctx, "secret1")
	if err != nil {
		t.Fatalf("GetSecretByID failed: %v", err)
	}
	if retrievedSecret == nil {
		t.Fatal("Expected non-nil secret")
	}
	if retrievedSecret.ID != "secret1" {
		t.Errorf("Expected secret ID to be 'secret1', got '%s'", retrievedSecret.ID)
	}

	// Test ListSecretsByUserID
	secrets, err := repo.ListSecretsByUserID(ctx, "user1")
	if err != nil {
		t.Fatalf("ListSecretsByUserID failed: %v", err)
	}
	if len(secrets) != 1 {
		t.Errorf("Expected 1 secret, got %d", len(secrets))
	}

	// Test UpdateSecret
	secret.Name = "Updated Secret"
	err = repo.UpdateSecret(ctx, secret)
	if err != nil {
		t.Fatalf("UpdateSecret failed: %v", err)
	}

	// Verify the update
	retrievedSecret, err = repo.GetSecretByID(ctx, "secret1")
	if err != nil {
		t.Fatalf("GetSecretByID failed after update: %v", err)
	}
	if retrievedSecret.Name != "Updated Secret" {
		t.Errorf("Expected updated name to be 'Updated Secret', got '%s'", retrievedSecret.Name)
	}

	// Test DeleteSecret
	err = repo.DeleteSecret(ctx, "secret1")
	if err != nil {
		t.Fatalf("DeleteSecret failed: %v", err)
	}

	// Verify the deletion
	secrets, err = repo.ListSecretsByUserID(ctx, "user1")
	if err != nil {
		t.Fatalf("ListSecretsByUserID failed after deletion: %v", err)
	}
	if len(secrets) != 0 {
		t.Errorf("Expected 0 secrets after deletion, got %d", len(secrets))
	}
}

func TestMockRepositoryRecipientMethods(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Test CreateRecipient
	recipient := &models.Recipient{
		ID:     "recipient1",
		UserID: "user1",
		Name:   "Test Recipient",
		Email:  "recipient@example.com",
	}
	err := repo.CreateRecipient(ctx, recipient)
	if err != nil {
		t.Fatalf("CreateRecipient failed: %v", err)
	}

	// Test GetRecipientByID
	retrievedRecipient, err := repo.GetRecipientByID(ctx, "recipient1")
	if err != nil {
		t.Fatalf("GetRecipientByID failed: %v", err)
	}
	if retrievedRecipient == nil {
		t.Fatal("Expected non-nil recipient")
	}
	if retrievedRecipient.ID != "recipient1" {
		t.Errorf("Expected recipient ID to be 'recipient1', got '%s'", retrievedRecipient.ID)
	}

	// Test ListRecipientsByUserID
	recipients, err := repo.ListRecipientsByUserID(ctx, "user1")
	if err != nil {
		t.Fatalf("ListRecipientsByUserID failed: %v", err)
	}
	if len(recipients) != 1 {
		t.Errorf("Expected 1 recipient, got %d", len(recipients))
	}

	// Test UpdateRecipient
	recipient.Name = "Updated Recipient"
	err = repo.UpdateRecipient(ctx, recipient)
	if err != nil {
		t.Fatalf("UpdateRecipient failed: %v", err)
	}

	// Verify the update
	retrievedRecipient, err = repo.GetRecipientByID(ctx, "recipient1")
	if err != nil {
		t.Fatalf("GetRecipientByID failed after update: %v", err)
	}
	if retrievedRecipient.Name != "Updated Recipient" {
		t.Errorf("Expected updated name to be 'Updated Recipient', got '%s'", retrievedRecipient.Name)
	}

	// Test DeleteRecipient
	err = repo.DeleteRecipient(ctx, "recipient1")
	if err != nil {
		t.Fatalf("DeleteRecipient failed: %v", err)
	}

	// Verify the deletion
	recipients, err = repo.ListRecipientsByUserID(ctx, "user1")
	if err != nil {
		t.Fatalf("ListRecipientsByUserID failed after deletion: %v", err)
	}
	if len(recipients) != 0 {
		t.Errorf("Expected 0 recipients after deletion, got %d", len(recipients))
	}
}

func TestMockRepositoryPasskeyMethods(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Test CreatePasskey
	passkey := &models.Passkey{
		ID:              "passkey1",
		UserID:          "user1",
		CredentialID:    []byte{1, 2, 3},
		PublicKey:       []byte{4, 5, 6},
		AAGUID:          []byte{7, 8, 9},
		SignCount:       1,
		Name:            "Test Passkey",
		CreatedAt:       time.Now(),
		LastUsedAt:      time.Now(),
		Transports:      []string{"internal"},
		AttestationType: "none",
	}
	err := repo.CreatePasskey(ctx, passkey)
	if err != nil {
		t.Fatalf("CreatePasskey failed: %v", err)
	}

	// Test GetPasskeyByID
	retrievedPasskey, err := repo.GetPasskeyByID(ctx, "passkey1")
	if err != nil {
		t.Fatalf("GetPasskeyByID failed: %v", err)
	}
	if retrievedPasskey == nil {
		t.Fatal("Expected non-nil passkey")
	}
	if retrievedPasskey.ID != "passkey1" {
		t.Errorf("Expected passkey ID to be 'passkey1', got '%s'", retrievedPasskey.ID)
	}

	// Test GetPasskeyByCredentialID
	retrievedPasskey, err = repo.GetPasskeyByCredentialID(ctx, []byte{1, 2, 3})
	if err != nil {
		t.Fatalf("GetPasskeyByCredentialID failed: %v", err)
	}
	if retrievedPasskey == nil {
		t.Fatal("Expected non-nil passkey")
	}
	if retrievedPasskey.ID != "passkey1" {
		t.Errorf("Expected passkey ID to be 'passkey1', got '%s'", retrievedPasskey.ID)
	}

	// Test ListPasskeysByUserID
	passkeys, err := repo.ListPasskeysByUserID(ctx, "user1")
	if err != nil {
		t.Fatalf("ListPasskeysByUserID failed: %v", err)
	}
	if len(passkeys) != 1 {
		t.Errorf("Expected 1 passkey, got %d", len(passkeys))
	}

	// Test ListPasskeys
	allPasskeys, err := repo.ListPasskeys(ctx)
	if err != nil {
		t.Fatalf("ListPasskeys failed: %v", err)
	}
	if len(allPasskeys) != 1 {
		t.Errorf("Expected 1 passkey, got %d", len(allPasskeys))
	}

	// Test UpdatePasskey
	passkey.Name = "Updated Passkey"
	err = repo.UpdatePasskey(ctx, passkey)
	if err != nil {
		t.Fatalf("UpdatePasskey failed: %v", err)
	}

	// Verify the update
	retrievedPasskey, err = repo.GetPasskeyByID(ctx, "passkey1")
	if err != nil {
		t.Fatalf("GetPasskeyByID failed after update: %v", err)
	}
	if retrievedPasskey.Name != "Updated Passkey" {
		t.Errorf("Expected updated name to be 'Updated Passkey', got '%s'", retrievedPasskey.Name)
	}

	// Test DeletePasskey
	err = repo.DeletePasskey(ctx, "passkey1")
	if err != nil {
		t.Fatalf("DeletePasskey failed: %v", err)
	}

	// Verify the deletion
	passkeys, err = repo.ListPasskeysByUserID(ctx, "user1")
	if err != nil {
		t.Fatalf("ListPasskeysByUserID failed after deletion: %v", err)
	}
	if len(passkeys) != 0 {
		t.Errorf("Expected 0 passkeys after deletion, got %d", len(passkeys))
	}
}

func TestMockRepositorySecretAssignmentMethods(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Test CreateSecretAssignment
	assignment := &models.SecretAssignment{
		ID:          "assignment1",
		UserID:      "user1",
		SecretID:    "secret1",
		RecipientID: "recipient1",
	}
	err := repo.CreateSecretAssignment(ctx, assignment)
	if err != nil {
		t.Fatalf("CreateSecretAssignment failed: %v", err)
	}

	// Test GetSecretAssignmentByID
	retrievedAssignment, err := repo.GetSecretAssignmentByID(ctx, "assignment1")
	if err != nil {
		t.Fatalf("GetSecretAssignmentByID failed: %v", err)
	}
	if retrievedAssignment == nil {
		t.Fatal("Expected non-nil assignment")
	}
	if retrievedAssignment.ID != "assignment1" {
		t.Errorf("Expected assignment ID to be 'assignment1', got '%s'", retrievedAssignment.ID)
	}

	// Test ListSecretAssignmentsBySecretID
	assignmentsBySecret, err := repo.ListSecretAssignmentsBySecretID(ctx, "secret1")
	if err != nil {
		t.Fatalf("ListSecretAssignmentsBySecretID failed: %v", err)
	}
	if len(assignmentsBySecret) != 1 {
		t.Errorf("Expected 1 assignment, got %d", len(assignmentsBySecret))
	}

	// Test ListSecretAssignmentsByRecipientID
	assignmentsByRecipient, err := repo.ListSecretAssignmentsByRecipientID(ctx, "recipient1")
	if err != nil {
		t.Fatalf("ListSecretAssignmentsByRecipientID failed: %v", err)
	}
	if len(assignmentsByRecipient) != 1 {
		t.Errorf("Expected 1 assignment, got %d", len(assignmentsByRecipient))
	}

	// Test ListSecretAssignmentsByUserID
	assignmentsByUser, err := repo.ListSecretAssignmentsByUserID(ctx, "user1")
	if err != nil {
		t.Fatalf("ListSecretAssignmentsByUserID failed: %v", err)
	}
	if len(assignmentsByUser) != 1 {
		t.Errorf("Expected 1 assignment, got %d", len(assignmentsByUser))
	}

	// Test DeleteSecretAssignment
	err = repo.DeleteSecretAssignment(ctx, "assignment1")
	if err != nil {
		t.Fatalf("DeleteSecretAssignment failed: %v", err)
	}

	// Verify the deletion
	assignmentsByUser, err = repo.ListSecretAssignmentsByUserID(ctx, "user1")
	if err != nil {
		t.Fatalf("ListSecretAssignmentsByUserID failed after deletion: %v", err)
	}
	if len(assignmentsByUser) != 0 {
		t.Errorf("Expected 0 assignments after deletion, got %d", len(assignmentsByUser))
	}
}

func TestMockRepositorySessionMethods(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Test CreateSession
	session := &models.Session{
		ID:           "session1",
		UserID:       "user1",
		Token:        "token1",
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		LastActivity: time.Now(),
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
	}
	err := repo.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Test GetSessionByToken
	retrievedSession, err := repo.GetSessionByToken(ctx, "token1")
	if err != nil {
		t.Fatalf("GetSessionByToken failed: %v", err)
	}
	if retrievedSession == nil {
		t.Fatal("Expected non-nil session")
	}
	if retrievedSession.ID != "session1" {
		t.Errorf("Expected session ID to be 'session1', got '%s'", retrievedSession.ID)
	}

	// Test UpdateSessionActivity
	err = repo.UpdateSessionActivity(ctx, "session1")
	if err != nil {
		t.Fatalf("UpdateSessionActivity failed: %v", err)
	}

	// Test DeleteSession
	err = repo.DeleteSession(ctx, "session1")
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	// Verify the deletion
	retrievedSession, err = repo.GetSessionByToken(ctx, "token1")
	if err == nil {
		t.Fatal("Expected error after session deletion, got nil")
	}
	if retrievedSession != nil {
		t.Errorf("Expected nil session after deletion, got %v", retrievedSession)
	}

	// Test DeleteExpiredSessions
	err = repo.DeleteExpiredSessions(ctx)
	if err != nil {
		t.Fatalf("DeleteExpiredSessions failed: %v", err)
	}
}

func TestMockRepositoryPingMethods(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Test CreatePingHistory
	pingHistory := &models.PingHistory{
		ID:     "ping1",
		UserID: "user1",
		SentAt: time.Now(),
		Method: "email",
		Status: "sent",
	}
	err := repo.CreatePingHistory(ctx, pingHistory)
	if err != nil {
		t.Fatalf("CreatePingHistory failed: %v", err)
	}

	// Test UpdatePingHistory
	pingHistory.Status = "responded"
	respondedAt := time.Now()
	pingHistory.RespondedAt = &respondedAt
	err = repo.UpdatePingHistory(ctx, pingHistory)
	if err != nil {
		t.Fatalf("UpdatePingHistory failed: %v", err)
	}

	// Test GetLatestPingByUserID
	latestPing, err := repo.GetLatestPingByUserID(ctx, "user1")
	if err != nil {
		t.Fatalf("GetLatestPingByUserID failed: %v", err)
	}
	if latestPing == nil {
		t.Fatal("Expected non-nil ping history")
	}
	if latestPing.ID != "ping1" {
		t.Errorf("Expected ping ID to be 'ping1', got '%s'", latestPing.ID)
	}

	// Test ListPingHistoryByUserID
	pingHistories, err := repo.ListPingHistoryByUserID(ctx, "user1")
	if err != nil {
		t.Fatalf("ListPingHistoryByUserID failed: %v", err)
	}
	if len(pingHistories) != 1 {
		t.Errorf("Expected 1 ping history, got %d", len(pingHistories))
	}

	// Test CreatePingVerification
	verification := &models.PingVerification{
		ID:        "verification1",
		UserID:    "user1",
		Code:      "code1",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Used:      false,
		CreatedAt: time.Now(),
	}
	err = repo.CreatePingVerification(ctx, verification)
	if err != nil {
		t.Fatalf("CreatePingVerification failed: %v", err)
	}

	// Test GetPingVerificationByCode
	retrievedVerification, err := repo.GetPingVerificationByCode(ctx, "code1")
	if err != nil {
		t.Fatalf("GetPingVerificationByCode failed: %v", err)
	}
	if retrievedVerification == nil {
		t.Fatal("Expected non-nil verification")
	}
	if retrievedVerification.ID != "verification1" {
		t.Errorf("Expected verification ID to be 'verification1', got '%s'", retrievedVerification.ID)
	}

	// Test UpdatePingVerification
	verification.Used = true
	err = repo.UpdatePingVerification(ctx, verification)
	if err != nil {
		t.Fatalf("UpdatePingVerification failed: %v", err)
	}

	// Verify the update
	retrievedVerification, err = repo.GetPingVerificationByCode(ctx, "code1")
	if err != nil {
		t.Fatalf("GetPingVerificationByCode failed after update: %v", err)
	}
	if !retrievedVerification.Used {
		t.Error("Expected verification to be marked as used")
	}
}

func TestMockRepositoryDeliveryMethods(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Test CreateDeliveryEvent
	event := &models.DeliveryEvent{
		ID:          "event1",
		UserID:      "user1",
		RecipientID: "recipient1",
		SentAt:      time.Now(),
		Status:      "sent",
	}
	err := repo.CreateDeliveryEvent(ctx, event)
	if err != nil {
		t.Fatalf("CreateDeliveryEvent failed: %v", err)
	}

	// Test ListDeliveryEventsByUserID
	events, err := repo.ListDeliveryEventsByUserID(ctx, "user1")
	if err != nil {
		t.Fatalf("ListDeliveryEventsByUserID failed: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 delivery event, got %d", len(events))
	}
}

func TestMockRepositoryAuditLogMethods(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Test CreateAuditLog
	log := &models.AuditLog{
		ID:        "log1",
		UserID:    "user1",
		Action:    "login",
		Timestamp: time.Now(),
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
		Details:   "test login",
	}
	err := repo.CreateAuditLog(ctx, log)
	if err != nil {
		t.Fatalf("CreateAuditLog failed: %v", err)
	}

	// Test ListAuditLogsByUserID
	logs, err := repo.ListAuditLogsByUserID(ctx, "user1")
	if err != nil {
		t.Fatalf("ListAuditLogsByUserID failed: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(logs))
	}
}

func TestMockRepositoryTransactionMethods(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Test BeginTx
	tx, err := repo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}
	if tx == nil {
		t.Fatal("Expected non-nil transaction")
	}

	// Test transaction methods
	user := &models.User{
		ID:    "user1",
		Email: "user1@example.com",
	}
	err = tx.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("Transaction CreateUser failed: %v", err)
	}

	// Test Commit
	err = tx.Commit()
	if err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Start another transaction for rollback
	tx, err = repo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}

	// Test Rollback
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
}

func TestMockRepositoryGetUsersForPinging(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Add users for pinging
	user1 := &models.User{
		ID:             "user1",
		Email:          "user1@example.com",
		PingingEnabled: true,
	}
	user2 := &models.User{
		ID:             "user2",
		Email:          "user2@example.com",
		PingingEnabled: true,
	}
	repo.UsersForPinging = []*models.User{user1, user2}

	// Test GetUsersForPinging
	users, err := repo.GetUsersForPinging(ctx)
	if err != nil {
		t.Fatalf("GetUsersForPinging failed: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users for pinging, got %d", len(users))
	}
}

func TestMockRepositoryGetUsersWithExpiredPings(t *testing.T) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Add users with expired pings
	user1 := &models.User{
		ID:             "user1",
		Email:          "user1@example.com",
		PingingEnabled: true,
	}
	user2 := &models.User{
		ID:             "user2",
		Email:          "user2@example.com",
		PingingEnabled: true,
	}
	repo.UsersWithExpiredPings = []*models.User{user1, user2}

	// Test GetUsersWithExpiredPings
	users, err := repo.GetUsersWithExpiredPings(ctx)
	if err != nil {
		t.Fatalf("GetUsersWithExpiredPings failed: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users with expired pings, got %d", len(users))
	}
}
