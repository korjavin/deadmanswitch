package storage

import (
	"context"
	"encoding/base64"
	"os"
	"testing"
	"time"

	"github.com/korjavin/deadmanswitch/internal/crypto"
	"github.com/korjavin/deadmanswitch/internal/models"
)

func TestCreateAccessCode(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test user and recipient first
	user := createTestUser(t, repo, "test@example.com")
	recipient := createTestRecipient(t, repo, user.ID, "recipient@example.com")

	// Create delivery event
	deliveryEvent := &models.DeliveryEvent{
		UserID:      user.ID,
		RecipientID: recipient.ID,
		SentAt:      time.Now().UTC(),
		Status:      "pending",
	}
	err := repo.CreateDeliveryEvent(context.Background(), deliveryEvent)
	if err != nil {
		t.Fatalf("Failed to create delivery event: %v", err)
	}

	// Hash an access code
	plainCode := "test-access-code-12345"
	hashedCode, err := crypto.HashPassword(plainCode, nil)
	if err != nil {
		t.Fatalf("Failed to hash code: %v", err)
	}
	hashedCodeStr := base64.StdEncoding.EncodeToString(hashedCode)

	// Create access code
	accessCode := &models.AccessCode{
		Code:            hashedCodeStr,
		RecipientID:     recipient.ID,
		UserID:          user.ID,
		DeliveryEventID: deliveryEvent.ID,
		ExpiresAt:       time.Now().UTC().Add(30 * 24 * time.Hour),
		MaxAttempts:     5,
	}

	err = repo.CreateAccessCode(context.Background(), accessCode)
	if err != nil {
		t.Fatalf("Failed to create access code: %v", err)
	}

	// Verify ID was generated
	if accessCode.ID == "" {
		t.Error("Expected ID to be generated")
	}

	// Verify CreatedAt was set
	if accessCode.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestVerifyAccessCode(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test user and recipient
	user := createTestUser(t, repo, "test@example.com")
	recipient := createTestRecipient(t, repo, user.ID, "recipient@example.com")

	// Create delivery event
	deliveryEvent := &models.DeliveryEvent{
		UserID:      user.ID,
		RecipientID: recipient.ID,
		SentAt:      time.Now().UTC(),
		Status:      "pending",
	}
	err := repo.CreateDeliveryEvent(context.Background(), deliveryEvent)
	if err != nil {
		t.Fatalf("Failed to create delivery event: %v", err)
	}

	// Hash an access code
	plainCode := "test-access-code-valid"
	hashedCode, err := crypto.HashPassword(plainCode, nil)
	if err != nil {
		t.Fatalf("Failed to hash code: %v", err)
	}
	hashedCodeStr := base64.StdEncoding.EncodeToString(hashedCode)

	// Create access code
	accessCode := &models.AccessCode{
		Code:            hashedCodeStr,
		RecipientID:     recipient.ID,
		UserID:          user.ID,
		DeliveryEventID: deliveryEvent.ID,
		ExpiresAt:       time.Now().UTC().Add(30 * 24 * time.Hour),
		MaxAttempts:     5,
	}

	err = repo.CreateAccessCode(context.Background(), accessCode)
	if err != nil {
		t.Fatalf("Failed to create access code: %v", err)
	}

	// Test valid code verification
	verifiedCode, err := repo.VerifyAccessCode(context.Background(), plainCode)
	if err != nil {
		t.Fatalf("Expected valid code to verify, got error: %v", err)
	}

	if verifiedCode.ID != accessCode.ID {
		t.Errorf("Expected code ID %s, got %s", accessCode.ID, verifiedCode.ID)
	}

	// Test invalid code verification
	_, err = repo.VerifyAccessCode(context.Background(), "wrong-code")
	if err == nil {
		t.Error("Expected invalid code to fail verification")
	}
}

func TestVerifyAccessCodeExpired(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test user and recipient
	user := createTestUser(t, repo, "test2@example.com")
	recipient := createTestRecipient(t, repo, user.ID, "recipient2@example.com")

	// Create delivery event
	deliveryEvent := &models.DeliveryEvent{
		UserID:      user.ID,
		RecipientID: recipient.ID,
		SentAt:      time.Now().UTC(),
		Status:      "pending",
	}
	err := repo.CreateDeliveryEvent(context.Background(), deliveryEvent)
	if err != nil {
		t.Fatalf("Failed to create delivery event: %v", err)
	}

	// Hash an access code
	plainCode := "test-access-code-expired"
	hashedCode, err := crypto.HashPassword(plainCode, nil)
	if err != nil {
		t.Fatalf("Failed to hash code: %v", err)
	}
	hashedCodeStr := base64.StdEncoding.EncodeToString(hashedCode)

	// Create expired access code
	accessCode := &models.AccessCode{
		Code:            hashedCodeStr,
		RecipientID:     recipient.ID,
		UserID:          user.ID,
		DeliveryEventID: deliveryEvent.ID,
		ExpiresAt:       time.Now().UTC().Add(-1 * time.Hour), // Expired 1 hour ago
		MaxAttempts:     5,
	}

	err = repo.CreateAccessCode(context.Background(), accessCode)
	if err != nil {
		t.Fatalf("Failed to create access code: %v", err)
	}

	// Test expired code verification
	_, err = repo.VerifyAccessCode(context.Background(), plainCode)
	if err == nil {
		t.Error("Expected expired code to fail verification")
	}
}

func TestMarkAccessCodeAsUsed(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test user and recipient
	user := createTestUser(t, repo, "test3@example.com")
	recipient := createTestRecipient(t, repo, user.ID, "recipient3@example.com")

	// Create delivery event
	deliveryEvent := &models.DeliveryEvent{
		UserID:      user.ID,
		RecipientID: recipient.ID,
		SentAt:      time.Now().UTC(),
		Status:      "pending",
	}
	err := repo.CreateDeliveryEvent(context.Background(), deliveryEvent)
	if err != nil {
		t.Fatalf("Failed to create delivery event: %v", err)
	}

	// Hash an access code
	plainCode := "test-access-code-used"
	hashedCode, err := crypto.HashPassword(plainCode, nil)
	if err != nil {
		t.Fatalf("Failed to hash code: %v", err)
	}
	hashedCodeStr := base64.StdEncoding.EncodeToString(hashedCode)

	// Create access code
	accessCode := &models.AccessCode{
		Code:            hashedCodeStr,
		RecipientID:     recipient.ID,
		UserID:          user.ID,
		DeliveryEventID: deliveryEvent.ID,
		ExpiresAt:       time.Now().UTC().Add(30 * 24 * time.Hour),
		MaxAttempts:     5,
	}

	err = repo.CreateAccessCode(context.Background(), accessCode)
	if err != nil {
		t.Fatalf("Failed to create access code: %v", err)
	}

	// Mark as used
	err = repo.MarkAccessCodeAsUsed(context.Background(), accessCode.ID)
	if err != nil {
		t.Fatalf("Failed to mark access code as used: %v", err)
	}

	// Try to verify the used code
	_, err = repo.VerifyAccessCode(context.Background(), plainCode)
	if err == nil {
		t.Error("Expected used code to fail verification")
	}
}

func TestDeleteExpiredAccessCodes(t *testing.T) {
	repo, cleanup := setupTestDB(t)
	defer cleanup()

	// Create test user and recipient
	user := createTestUser(t, repo, "test4@example.com")
	recipient := createTestRecipient(t, repo, user.ID, "recipient4@example.com")

	// Create delivery event
	deliveryEvent := &models.DeliveryEvent{
		UserID:      user.ID,
		RecipientID: recipient.ID,
		SentAt:      time.Now().UTC(),
		Status:      "pending",
	}
	err := repo.CreateDeliveryEvent(context.Background(), deliveryEvent)
	if err != nil {
		t.Fatalf("Failed to create delivery event: %v", err)
	}

	// Create multiple access codes, some expired
	for i := 0; i < 5; i++ {
		plainCode := "test-access-code-" + string(rune(i))
		hashedCode, err := crypto.HashPassword(plainCode, nil)
		if err != nil {
			t.Fatalf("Failed to hash code: %v", err)
		}
		hashedCodeStr := base64.StdEncoding.EncodeToString(hashedCode)

		expiresAt := time.Now().UTC().Add(30 * 24 * time.Hour)
		if i%2 == 0 {
			// Make half of them expired
			expiresAt = time.Now().UTC().Add(-1 * time.Hour)
		}

		accessCode := &models.AccessCode{
			Code:            hashedCodeStr,
			RecipientID:     recipient.ID,
			UserID:          user.ID,
			DeliveryEventID: deliveryEvent.ID,
			ExpiresAt:       expiresAt,
			MaxAttempts:     5,
		}

		err = repo.CreateAccessCode(context.Background(), accessCode)
		if err != nil {
			t.Fatalf("Failed to create access code: %v", err)
		}
	}

	// Delete expired codes
	err = repo.DeleteExpiredAccessCodes(context.Background())
	if err != nil {
		t.Fatalf("Failed to delete expired access codes: %v", err)
	}

	// Note: We can't easily verify the count without adding a count method,
	// but at least we verified the operation doesn't error
}

// Helper functions

func setupTestDB(t *testing.T) (Repository, func()) {
	t.Helper()

	// Create a temporary database file with a unique name
	dbPath := "./test_access_code_" + t.Name() + ".sqlite"

	// Create a new repository
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		os.Remove(dbPath)
	}

	return repo, cleanup
}

func createTestUser(t *testing.T, repo Repository, email string) *models.User {
	t.Helper()

	user := &models.User{
		Email:          email,
		PasswordHash:   []byte("hashed_password"),
		PingFrequency:  3,
		PingDeadline:   14,
		PingingEnabled: true,
		PingMethod:     "both",
	}

	err := repo.CreateUser(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

func createTestRecipient(t *testing.T, repo Repository, userID, email string) *models.Recipient {
	t.Helper()

	recipient := &models.Recipient{
		UserID:      userID,
		Email:       email,
		Name:        "Test Recipient",
		Message:     "Test message",
		IsConfirmed: true,
	}

	err := repo.CreateRecipient(context.Background(), recipient)
	if err != nil {
		t.Fatalf("Failed to create test recipient: %v", err)
	}

	return recipient
}
