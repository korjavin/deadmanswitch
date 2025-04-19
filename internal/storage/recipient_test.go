package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
)

// TestSQLiteRepository_RecipientOperations tests recipient operations
func TestSQLiteRepository_RecipientOperations(t *testing.T) {
	// Create a temporary database file
	dbPath := "./test_recipient_ops.sqlite"
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
		IsConfirmed:      false,
		ConfirmationCode: "abc123",
	}

	// Test CreateRecipient
	err = repo.CreateRecipient(ctx, recipient)
	if err != nil {
		t.Fatalf("Failed to create recipient: %v", err)
	}

	// Check that the recipient ID was generated
	if recipient.ID == "" {
		t.Fatal("Recipient ID was not generated")
	}

	// Test GetRecipientByID
	retrievedRecipient, err := repo.GetRecipientByID(ctx, recipient.ID)
	if err != nil {
		t.Fatalf("Failed to get recipient by ID: %v", err)
	}
	if retrievedRecipient.Name != recipient.Name {
		t.Errorf("Expected name %s, got %s", recipient.Name, retrievedRecipient.Name)
	}
	if retrievedRecipient.Email != recipient.Email {
		t.Errorf("Expected email %s, got %s", recipient.Email, retrievedRecipient.Email)
	}
	if retrievedRecipient.PhoneNumber != recipient.PhoneNumber {
		t.Errorf("Expected phone number %s, got %s", recipient.PhoneNumber, retrievedRecipient.PhoneNumber)
	}
	if retrievedRecipient.IsConfirmed != recipient.IsConfirmed {
		t.Errorf("Expected IsConfirmed %v, got %v", recipient.IsConfirmed, retrievedRecipient.IsConfirmed)
	}
	if retrievedRecipient.ConfirmationCode != recipient.ConfirmationCode {
		t.Errorf("Expected confirmation code %s, got %s", recipient.ConfirmationCode, retrievedRecipient.ConfirmationCode)
	}
	// Both should be nil since we didn't set it
	if retrievedRecipient.ConfirmationSentAt != nil {
		t.Errorf("Expected nil ConfirmationSentAt, got %v", *retrievedRecipient.ConfirmationSentAt)
	}

	// Test ListRecipientsByUserID
	recipients, err := repo.ListRecipientsByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to list recipients by user ID: %v", err)
	}
	if len(recipients) != 1 {
		t.Errorf("Expected 1 recipient, got %d", len(recipients))
	}

	// Test UpdateRecipient
	confirmedAt := time.Now()
	recipient.Name = "Updated Recipient"
	recipient.IsConfirmed = true
	recipient.ConfirmedAt = &confirmedAt
	err = repo.UpdateRecipient(ctx, recipient)
	if err != nil {
		t.Fatalf("Failed to update recipient: %v", err)
	}

	// Verify the update
	retrievedRecipient, err = repo.GetRecipientByID(ctx, recipient.ID)
	if err != nil {
		t.Fatalf("Failed to get updated recipient: %v", err)
	}
	if retrievedRecipient.Name != "Updated Recipient" {
		t.Errorf("Expected updated name 'Updated Recipient', got %s", retrievedRecipient.Name)
	}
	if !retrievedRecipient.IsConfirmed {
		t.Errorf("Expected IsConfirmed to be true")
	}
	if retrievedRecipient.ConfirmedAt == nil {
		t.Errorf("Expected non-nil ConfirmedAt")
	} else if recipient.ConfirmedAt == nil {
		t.Errorf("Expected non-nil recipient.ConfirmedAt")
	} else if !retrievedRecipient.ConfirmedAt.Equal(*recipient.ConfirmedAt) {
		t.Errorf("Expected confirmed at %v, got %v", *recipient.ConfirmedAt, *retrievedRecipient.ConfirmedAt)
	}

	// Test DeleteRecipient
	err = repo.DeleteRecipient(ctx, recipient.ID)
	if err != nil {
		t.Fatalf("Failed to delete recipient: %v", err)
	}

	// Verify the deletion
	_, err = repo.GetRecipientByID(ctx, recipient.ID)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

// TestSQLiteRepository_SecretAssignmentOperations tests secret assignment operations
func TestSQLiteRepository_SecretAssignmentOperations(t *testing.T) {
	// Create a temporary database file
	dbPath := "./test_secret_assignment_ops.sqlite"
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

	// Create a test secret
	secret := &models.Secret{
		UserID:         user.ID,
		Name:           "Test Secret",
		EncryptedData:  "encrypted_data",
		EncryptionType: "aes-256-gcm",
	}
	err = repo.CreateSecret(ctx, secret)
	if err != nil {
		t.Fatalf("Failed to create secret: %v", err)
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

	// Create a test secret assignment
	assignment := &models.SecretAssignment{
		UserID:      user.ID,
		SecretID:    secret.ID,
		RecipientID: recipient.ID,
	}

	// Test CreateSecretAssignment
	err = repo.CreateSecretAssignment(ctx, assignment)
	if err != nil {
		t.Fatalf("Failed to create secret assignment: %v", err)
	}

	// Check that the assignment ID was generated
	if assignment.ID == "" {
		t.Fatal("Assignment ID was not generated")
	}

	// Test GetSecretAssignmentByID
	retrievedAssignment, err := repo.GetSecretAssignmentByID(ctx, assignment.ID)
	if err != nil {
		t.Fatalf("Failed to get secret assignment by ID: %v", err)
	}
	if retrievedAssignment.SecretID != assignment.SecretID {
		t.Errorf("Expected secret ID %s, got %s", assignment.SecretID, retrievedAssignment.SecretID)
	}
	if retrievedAssignment.RecipientID != assignment.RecipientID {
		t.Errorf("Expected recipient ID %s, got %s", assignment.RecipientID, retrievedAssignment.RecipientID)
	}

	// Test ListSecretAssignmentsBySecretID
	secretAssignments, err := repo.ListSecretAssignmentsBySecretID(ctx, secret.ID)
	if err != nil {
		t.Fatalf("Failed to list secret assignments by secret ID: %v", err)
	}
	if len(secretAssignments) != 1 {
		t.Errorf("Expected 1 secret assignment, got %d", len(secretAssignments))
	}

	// Test ListSecretAssignmentsByRecipientID
	recipientAssignments, err := repo.ListSecretAssignmentsByRecipientID(ctx, recipient.ID)
	if err != nil {
		t.Fatalf("Failed to list secret assignments by recipient ID: %v", err)
	}
	if len(recipientAssignments) != 1 {
		t.Errorf("Expected 1 secret assignment, got %d", len(recipientAssignments))
	}

	// Test ListSecretAssignmentsByUserID
	userAssignments, err := repo.ListSecretAssignmentsByUserID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to list secret assignments by user ID: %v", err)
	}
	if len(userAssignments) != 1 {
		t.Errorf("Expected 1 secret assignment, got %d", len(userAssignments))
	}

	// Test DeleteSecretAssignment
	err = repo.DeleteSecretAssignment(ctx, assignment.ID)
	if err != nil {
		t.Fatalf("Failed to delete secret assignment: %v", err)
	}

	// Verify the deletion
	_, err = repo.GetSecretAssignmentByID(ctx, assignment.ID)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}
