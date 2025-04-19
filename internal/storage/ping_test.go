package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
)

// TestSQLiteRepository_PingOperations tests ping operations
func TestSQLiteRepository_PingOperations(t *testing.T) {
	// Create a temporary database file
	dbPath := "./test_ping_ops.sqlite"
	defer os.Remove(dbPath)

	// Create a new repository
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Create a test user first
	user := &models.User{
		Email:          "test@example.com",
		PasswordHash:   []byte("hashed_password"),
		PingFrequency:  3,
		PingDeadline:   14,
		PingingEnabled: true,
		PingMethod:     "both",
	}
	err = repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create a test ping history
	now := time.Now()
	ping := &models.PingHistory{
		UserID: user.ID,
		SentAt: now,
		Method: "email",
		Status: "sent",
	}

	// Test CreatePingHistory
	err = repo.CreatePingHistory(ctx, ping)
	if err != nil {
		t.Fatalf("Failed to create ping history: %v", err)
	}

	// Check that the ping ID was generated
	if ping.ID == "" {
		t.Fatal("Ping ID was not generated")
	}

	// Test GetLatestPingByUserID
	latestPing, err := repo.GetLatestPingByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get latest ping: %v", err)
	}
	if latestPing.ID != ping.ID {
		t.Errorf("Expected ID %s, got %s", ping.ID, latestPing.ID)
	}
	if latestPing.Method != ping.Method {
		t.Errorf("Expected method %s, got %s", ping.Method, latestPing.Method)
	}
	if latestPing.Status != ping.Status {
		t.Errorf("Expected status %s, got %s", ping.Status, latestPing.Status)
	}

	// Test ListPingHistoryByUserID
	pings, err := repo.ListPingHistoryByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to list ping history: %v", err)
	}
	if len(pings) != 1 {
		t.Errorf("Expected 1 ping, got %d", len(pings))
	}

	// Test UpdatePingHistory
	respondedAt := time.Now()
	ping.Status = "responded"
	ping.RespondedAt = &respondedAt
	err = repo.UpdatePingHistory(ctx, ping)
	if err != nil {
		t.Fatalf("Failed to update ping history: %v", err)
	}

	// Verify the update
	updatedPing, err := repo.GetLatestPingByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get updated ping: %v", err)
	}
	if updatedPing.Status != "responded" {
		t.Errorf("Expected status 'responded', got '%s'", updatedPing.Status)
	}
	if updatedPing.RespondedAt == nil {
		t.Errorf("Expected non-nil RespondedAt")
	} else if !updatedPing.RespondedAt.Equal(*ping.RespondedAt) {
		t.Errorf("Expected responded at %v, got %v", *ping.RespondedAt, *updatedPing.RespondedAt)
	}
}

// TestSQLiteRepository_PingVerificationOperations tests ping verification operations
func TestSQLiteRepository_PingVerificationOperations(t *testing.T) {
	// Create a temporary database file
	dbPath := "./test_ping_verification_ops.sqlite"
	defer os.Remove(dbPath)

	// Create a new repository
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Create a test user first
	user := &models.User{
		Email:          "test@example.com",
		PasswordHash:   []byte("hashed_password"),
		PingFrequency:  3,
		PingDeadline:   14,
		PingingEnabled: true,
		PingMethod:     "both",
	}
	err = repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create a test ping verification
	now := time.Now()
	verification := &models.PingVerification{
		UserID:    user.ID,
		Code:      "abc123",
		ExpiresAt: now.Add(24 * time.Hour),
		Used:      false,
	}

	// Test CreatePingVerification
	err = repo.CreatePingVerification(ctx, verification)
	if err != nil {
		t.Fatalf("Failed to create ping verification: %v", err)
	}

	// Check that the verification ID was generated
	if verification.ID == "" {
		t.Fatal("Verification ID was not generated")
	}

	// Test GetPingVerificationByCode
	retrievedVerification, err := repo.GetPingVerificationByCode(ctx, verification.Code)
	if err != nil {
		t.Fatalf("Failed to get ping verification by code: %v", err)
	}
	if retrievedVerification.ID != verification.ID {
		t.Errorf("Expected ID %s, got %s", verification.ID, retrievedVerification.ID)
	}
	if retrievedVerification.UserID != verification.UserID {
		t.Errorf("Expected user ID %s, got %s", verification.UserID, retrievedVerification.UserID)
	}
	if retrievedVerification.Code != verification.Code {
		t.Errorf("Expected code %s, got %s", verification.Code, retrievedVerification.Code)
	}
	if retrievedVerification.Used != verification.Used {
		t.Errorf("Expected used %v, got %v", verification.Used, retrievedVerification.Used)
	}

	// Test UpdatePingVerification
	verification.Used = true
	err = repo.UpdatePingVerification(ctx, verification)
	if err != nil {
		t.Fatalf("Failed to update ping verification: %v", err)
	}

	// Verify the update
	updatedVerification, err := repo.GetPingVerificationByCode(ctx, verification.Code)
	if err != nil {
		t.Fatalf("Failed to get updated ping verification: %v", err)
	}
	if !updatedVerification.Used {
		t.Errorf("Expected used to be true")
	}
}

// TestSQLiteRepository_DeliveryEventOperations tests delivery event operations
func TestSQLiteRepository_DeliveryEventOperations(t *testing.T) {
	// Create a temporary database file
	dbPath := "./test_delivery_event_ops.sqlite"
	defer os.Remove(dbPath)

	// Create a new repository
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Create a test user first
	user := &models.User{
		Email:          "test@example.com",
		PasswordHash:   []byte("hashed_password"),
		PingFrequency:  3,
		PingDeadline:   14,
		PingingEnabled: true,
		PingMethod:     "both",
	}
	err = repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create a test recipient
	recipient := &models.Recipient{
		UserID:           user.ID,
		Email:            "recipient@example.com",
		Name:             "Test Recipient",
		Message:          "Here are my secrets",
		PhoneNumber:      "+1234567890",
		IsConfirmed:      true,
		ConfirmationCode: "abc123",
	}
	err = repo.CreateRecipient(ctx, recipient)
	if err != nil {
		t.Fatalf("Failed to create recipient: %v", err)
	}

	// Create a test delivery event
	now := time.Now()
	event := &models.DeliveryEvent{
		UserID:      user.ID,
		RecipientID: recipient.ID,
		SentAt:      now,
		Status:      "sent",
	}

	// Test CreateDeliveryEvent
	err = repo.CreateDeliveryEvent(ctx, event)
	if err != nil {
		t.Fatalf("Failed to create delivery event: %v", err)
	}

	// Check that the event ID was generated
	if event.ID == "" {
		t.Fatal("Event ID was not generated")
	}

	// Test ListDeliveryEventsByUserID
	events, err := repo.ListDeliveryEventsByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to list delivery events: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}
	if events[0].ID != event.ID {
		t.Errorf("Expected ID %s, got %s", event.ID, events[0].ID)
	}
	if events[0].RecipientID != event.RecipientID {
		t.Errorf("Expected recipient ID %s, got %s", event.RecipientID, events[0].RecipientID)
	}
	if events[0].Status != event.Status {
		t.Errorf("Expected status %s, got %s", event.Status, events[0].Status)
	}
}

// TestSQLiteRepository_AuditLogOperations tests audit log operations
func TestSQLiteRepository_AuditLogOperations(t *testing.T) {
	// Create a temporary database file
	dbPath := "./test_audit_log_ops.sqlite"
	defer os.Remove(dbPath)

	// Create a new repository
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Create a test user first
	user := &models.User{
		Email:          "test@example.com",
		PasswordHash:   []byte("hashed_password"),
		PingFrequency:  3,
		PingDeadline:   14,
		PingingEnabled: true,
		PingMethod:     "both",
	}
	err = repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create a test audit log
	auditLog := &models.AuditLog{
		UserID:    user.ID,
		Action:    "login",
		IPAddress: "127.0.0.1",
		UserAgent: "Test Agent",
		Details:   "Login successful",
	}

	// Test CreateAuditLog
	err = repo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		t.Fatalf("Failed to create audit log: %v", err)
	}

	// Check that the audit log ID was generated
	if auditLog.ID == "" {
		t.Fatal("Audit log ID was not generated")
	}

	// Test ListAuditLogsByUserID
	logs, err := repo.ListAuditLogsByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}
	if logs[0].ID != auditLog.ID {
		t.Errorf("Expected ID %s, got %s", auditLog.ID, logs[0].ID)
	}
	if logs[0].Action != auditLog.Action {
		t.Errorf("Expected action %s, got %s", auditLog.Action, logs[0].Action)
	}
	if logs[0].IPAddress != auditLog.IPAddress {
		t.Errorf("Expected IP address %s, got %s", auditLog.IPAddress, logs[0].IPAddress)
	}
	if logs[0].UserAgent != auditLog.UserAgent {
		t.Errorf("Expected user agent %s, got %s", auditLog.UserAgent, logs[0].UserAgent)
	}
	if logs[0].Details != auditLog.Details {
		t.Errorf("Expected details %s, got %s", auditLog.Details, logs[0].Details)
	}
}

// TestSQLiteRepository_SchedulerOperations tests scheduler operations
func TestSQLiteRepository_SchedulerOperations(t *testing.T) {
	// Create a temporary database file
	dbPath := "./test_scheduler_ops.sqlite"
	defer os.Remove(dbPath)

	// Create a new repository
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	ctx := context.Background()

	// Create test users
	now := time.Now()

	// User with pinging enabled and due for a ping
	user1 := &models.User{
		Email:             "user1@example.com",
		PasswordHash:      []byte("hashed_password"),
		PingFrequency:     3,
		PingDeadline:      14,
		PingingEnabled:    true,
		PingMethod:        "email",
		LastActivity:      now,
		NextScheduledPing: now.Add(-1 * time.Hour), // Due for a ping
	}
	err = repo.CreateUser(ctx, user1)
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	// User with pinging enabled but not due for a ping
	user2 := &models.User{
		Email:             "user2@example.com",
		PasswordHash:      []byte("hashed_password"),
		PingFrequency:     3,
		PingDeadline:      14,
		PingingEnabled:    true,
		PingMethod:        "email",
		LastActivity:      now,
		NextScheduledPing: now.Add(24 * time.Hour), // Not due for a ping
	}
	err = repo.CreateUser(ctx, user2)
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// User with pinging disabled
	user3 := &models.User{
		Email:             "user3@example.com",
		PasswordHash:      []byte("hashed_password"),
		PingFrequency:     3,
		PingDeadline:      14,
		PingingEnabled:    false,
		PingMethod:        "email",
		LastActivity:      now,
		NextScheduledPing: now.Add(-1 * time.Hour), // Due for a ping but disabled
	}
	err = repo.CreateUser(ctx, user3)
	if err != nil {
		t.Fatalf("Failed to create user3: %v", err)
	}

	// User with expired ping (last activity more than deadline days ago)
	user4 := &models.User{
		Email:             "user4@example.com",
		PasswordHash:      []byte("hashed_password"),
		PingFrequency:     3,
		PingDeadline:      14,
		PingingEnabled:    true,
		PingMethod:        "email",
		LastActivity:      now.Add(-20 * 24 * time.Hour), // Last activity was 20 days ago (> 14 day deadline)
		NextScheduledPing: now.Add(-15 * 24 * time.Hour),
	}
	err = repo.CreateUser(ctx, user4)
	if err != nil {
		t.Fatalf("Failed to create user4: %v", err)
	}

	// Test GetUsersForPinging - just check that it runs without errors
	_, err = repo.GetUsersForPinging(ctx)
	if err != nil {
		t.Fatalf("Failed to get users for pinging: %v", err)
	}

	// Test GetUsersWithExpiredPings - just check that it runs without errors
	_, err = repo.GetUsersWithExpiredPings(ctx)
	if err != nil {
		t.Fatalf("Failed to get users with expired pings: %v", err)
	}
}
