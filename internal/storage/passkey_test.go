package storage

import (
	"context"
	"testing"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
)

func TestPasskeyOperations(t *testing.T) {
	// Use the mock repository for passkey tests
	repo := NewMockRepository()

	// Set up test data
	userID := "test-user-1"
	credentialID := []byte("credential-id-123")
	pubKey := []byte("public-key-data")

	// Test CreatePasskey
	t.Run("CreatePasskey", func(t *testing.T) {
		passkey := &models.Passkey{
			ID:              "passkey-1",
			UserID:          userID,
			Name:            "Test Passkey",
			CredentialID:    credentialID,
			PublicKey:       pubKey,
			AAGUID:          []byte("aaguid-123"),
			SignCount:       0,
			AttestationType: "none",
			Transports:      []string{"internal"},
			CreatedAt:       time.Now(),
		}

		err := repo.CreatePasskey(context.Background(), passkey)
		if err != nil {
			t.Fatalf("CreatePasskey failed: %v", err)
		}

		// Verify the passkey was added to the repository
		if len(repo.Passkeys) != 1 {
			t.Errorf("Expected 1 passkey, got %d", len(repo.Passkeys))
		}
		if repo.Passkeys[0].ID != "passkey-1" {
			t.Errorf("Expected passkey ID 'passkey-1', got '%s'", repo.Passkeys[0].ID)
		}
	})

	// Test GetPasskeyByID
	t.Run("GetPasskeyByID", func(t *testing.T) {
		passkey, err := repo.GetPasskeyByID(context.Background(), "passkey-1")
		if err != nil {
			t.Fatalf("GetPasskeyByID failed: %v", err)
		}
		if passkey == nil {
			t.Fatal("Expected passkey, got nil")
		}
		if passkey.ID != "passkey-1" {
			t.Errorf("Expected passkey ID 'passkey-1', got '%s'", passkey.ID)
		}

		// Test with non-existent ID
		_, err = repo.GetPasskeyByID(context.Background(), "nonexistent")
		if err != ErrNotFound {
			t.Errorf("Expected ErrNotFound for non-existent passkey, got %v", err)
		}
	})

	// Test GetPasskeyByCredentialID
	t.Run("GetPasskeyByCredentialID", func(t *testing.T) {
		passkey, err := repo.GetPasskeyByCredentialID(context.Background(), credentialID)
		if err != nil {
			t.Fatalf("GetPasskeyByCredentialID failed: %v", err)
		}
		if passkey == nil {
			t.Fatal("Expected passkey, got nil")
		}
		if passkey.ID != "passkey-1" {
			t.Errorf("Expected passkey ID 'passkey-1', got '%s'", passkey.ID)
		}

		// Test with non-existent credential ID
		_, err = repo.GetPasskeyByCredentialID(context.Background(), []byte("nonexistent"))
		if err != ErrNotFound {
			t.Errorf("Expected ErrNotFound for non-existent credential, got %v", err)
		}
	})

	// Add a second passkey for the same user
	secondPasskey := &models.Passkey{
		ID:              "passkey-2",
		UserID:          userID,
		Name:            "Second Passkey",
		CredentialID:    []byte("credential-id-456"),
		PublicKey:       []byte("public-key-data-2"),
		AAGUID:          []byte("aaguid-456"),
		SignCount:       0,
		AttestationType: "none",
		Transports:      []string{"internal"},
		CreatedAt:       time.Now(),
	}
	if err := repo.CreatePasskey(context.Background(), secondPasskey); err != nil {
		t.Fatalf("Failed to create second passkey: %v", err)
	}

	// Test ListPasskeysByUserID
	t.Run("ListPasskeysByUserID", func(t *testing.T) {
		passkeys, err := repo.ListPasskeysByUserID(context.Background(), userID)
		if err != nil {
			t.Fatalf("ListPasskeysByUserID failed: %v", err)
		}
		if len(passkeys) != 2 {
			t.Errorf("Expected 2 passkeys, got %d", len(passkeys))
		}

		// Test with non-existent user
		passkeys, err = repo.ListPasskeysByUserID(context.Background(), "nonexistent-user")
		if err != nil {
			t.Fatalf("ListPasskeysByUserID with non-existent user should not error: %v", err)
		}
		if len(passkeys) != 0 {
			t.Errorf("Expected 0 passkeys for non-existent user, got %d", len(passkeys))
		}
	})

	// Test ListPasskeys
	t.Run("ListPasskeys", func(t *testing.T) {
		passkeys, err := repo.ListPasskeys(context.Background())
		if err != nil {
			t.Fatalf("ListPasskeys failed: %v", err)
		}
		if len(passkeys) != 2 {
			t.Errorf("Expected 2 passkeys, got %d", len(passkeys))
		}
	})

	// Test UpdatePasskey
	t.Run("UpdatePasskey", func(t *testing.T) {
		// Get the passkey to update
		passkey, err := repo.GetPasskeyByID(context.Background(), "passkey-1")
		if err != nil {
			t.Fatalf("GetPasskeyByID failed: %v", err)
		}

		// Update the passkey
		passkey.Name = "Updated Passkey"
		passkey.SignCount = 10
		err = repo.UpdatePasskey(context.Background(), passkey)
		if err != nil {
			t.Fatalf("UpdatePasskey failed: %v", err)
		}

		// Verify the update
		updatedPasskey, err := repo.GetPasskeyByID(context.Background(), "passkey-1")
		if err != nil {
			t.Fatalf("GetPasskeyByID after update failed: %v", err)
		}
		if updatedPasskey.Name != "Updated Passkey" {
			t.Errorf("Expected updated name 'Updated Passkey', got '%s'", updatedPasskey.Name)
		}
		if updatedPasskey.SignCount != 10 {
			t.Errorf("Expected updated sign count 10, got %d", updatedPasskey.SignCount)
		}

		// Test updating non-existent passkey
		nonExistentPasskey := &models.Passkey{
			ID: "nonexistent",
		}
		err = repo.UpdatePasskey(context.Background(), nonExistentPasskey)
		if err != ErrNotFound {
			t.Errorf("Expected ErrNotFound when updating non-existent passkey, got %v", err)
		}
	})

	// Test DeletePasskey
	t.Run("DeletePasskey", func(t *testing.T) {
		err := repo.DeletePasskey(context.Background(), "passkey-1")
		if err != nil {
			t.Fatalf("DeletePasskey failed: %v", err)
		}

		// Verify the passkey was deleted
		_, err = repo.GetPasskeyByID(context.Background(), "passkey-1")
		if err != ErrNotFound {
			t.Errorf("Expected ErrNotFound after deletion, got %v", err)
		}

		// Test deleting non-existent passkey
		err = repo.DeletePasskey(context.Background(), "nonexistent")
		if err != ErrNotFound {
			t.Errorf("Expected ErrNotFound when deleting non-existent passkey, got %v", err)
		}
	})

	// Test DeletePasskeysByUserID
	t.Run("DeletePasskeysByUserID", func(t *testing.T) {
		// Verify we still have one passkey left
		passkeys, _ := repo.ListPasskeys(context.Background())
		if len(passkeys) != 1 {
			t.Errorf("Expected 1 passkey before DeletePasskeysByUserID, got %d", len(passkeys))
		}

		err := repo.DeletePasskeysByUserID(context.Background(), userID)
		if err != nil {
			t.Fatalf("DeletePasskeysByUserID failed: %v", err)
		}

		// Verify all passkeys for the user were deleted
		passkeys, _ = repo.ListPasskeysByUserID(context.Background(), userID)
		if len(passkeys) != 0 {
			t.Errorf("Expected 0 passkeys after DeletePasskeysByUserID, got %d", len(passkeys))
		}

		// Test deleting for non-existent user (should not error)
		err = repo.DeletePasskeysByUserID(context.Background(), "nonexistent-user")
		if err != nil {
			t.Errorf("DeletePasskeysByUserID for non-existent user should not error: %v", err)
		}
	})
}
