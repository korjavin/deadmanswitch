package storage_test

import (
	"context"
	// "database/sql" // Removed as it's not directly used
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	// Ensure SQLite driver is loaded
	_ "modernc.org/sqlite"
)

// newTestStore creates a new SQLiteRepository backed by a temporary database file.
// It ensures that migrations are run because NewSQLiteRepository calls an initialize method.
func newTestStore(t *testing.T) *storage.SQLiteRepository {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), "test_*.db")
	require.NoError(t, err, "Failed to create temp db file")
	dbPath := tmpFile.Name()
	// Close the file immediately; NewSQLiteRepository will open and manage its own connection.
	require.NoError(t, tmpFile.Close(), "Failed to close temp db file")

	// NewSQLiteRepository internally calls initialize(), which runs migrations.
	storeInstance, err := storage.NewSQLiteRepository(dbPath)
	require.NoError(t, err, "Failed to create SQLiteRepository")

	return storeInstance
}

// Helper functions to create prerequisite records
func createTestUser(t *testing.T, store storage.Repository) *models.User {
	t.Helper()
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        uuid.New().String() + "@example.com",
		PasswordHash: []byte("hash"),
		LastActivity: time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	err := store.CreateUser(context.Background(), user)
	require.NoError(t, err)
	return user
}

func createTestSecret(t *testing.T, store storage.Repository, userID string) *models.Secret {
	t.Helper()
	secret := &models.Secret{
		ID:             uuid.New().String(),
		UserID:         userID,
		Name:           "Test Secret",
		EncryptedData:  "encrypteddata",
		EncryptionType: "aes",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	err := store.CreateSecret(context.Background(), secret)
	require.NoError(t, err)
	return secret
}

func createTestRecipient(t *testing.T, store storage.Repository, userID string) *models.Recipient {
	t.Helper()
	recipient := &models.Recipient{
		ID:        uuid.New().String(),
		UserID:    userID,
		Email:     uuid.New().String() + "@recipient.com",
		Name:      "Test Recipient",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := store.CreateRecipient(context.Background(), recipient)
	require.NoError(t, err)
	return recipient
}

func createTestSecretAssignment(t *testing.T, store storage.Repository, userID, secretID, recipientID string) *models.SecretAssignment {
	t.Helper()
	assignment := &models.SecretAssignment{
		ID:          uuid.New().String(),
		SecretID:    secretID,
		RecipientID: recipientID,
		UserID:      userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := store.CreateSecretAssignment(context.Background(), assignment)
	require.NoError(t, err)
	return assignment
}

// createTestSecretQuestionSet is a helper function to create a SecretQuestionSet for testing.
// It now also creates the necessary prerequisite User, Secret, Recipient, and SecretAssignment records.
func createTestSecretQuestionSet(t *testing.T, store storage.Repository, setDetails *models.SecretQuestionSet) *models.SecretQuestionSet {
	t.Helper()

	user := createTestUser(t, store)
	secret := createTestSecret(t, store, user.ID)
	recipient := createTestRecipient(t, store, user.ID)
	assignment := createTestSecretAssignment(t, store, user.ID, secret.ID, recipient.ID)

	fullSet := &models.SecretQuestionSet{
		ID:                 setDetails.ID, // Use provided ID or generate if empty
		SecretAssignmentID: assignment.ID, // THIS IS THE CRUCIAL LINK
		Threshold:          setDetails.Threshold,
		TotalQuestions:     setDetails.TotalQuestions,
		TimelockRound:      setDetails.TimelockRound,
		EncryptedBlob:      setDetails.EncryptedBlob,
	}

	if fullSet.ID == "" {
		fullSet.ID = uuid.New().String()
	}

	err := store.CreateSecretQuestionSet(context.Background(), fullSet)
	require.NoError(t, err)
	require.NotEmpty(t, fullSet.ID)
	require.NotZero(t, fullSet.CreatedAt)
	require.NotZero(t, fullSet.UpdatedAt)
	return fullSet
}

// TestCreateSecretQuestionSet will be implemented here
func TestCreateSecretQuestionSet(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		setDetails := &models.SecretQuestionSet{ // Details to pass to our helper
			Threshold:      2,
			TotalQuestions: 3,
			TimelockRound:  100,
			EncryptedBlob:  []byte("testblob"),
		}
		// createTestSecretQuestionSet now handles creating prerequisites and the actual SecretQuestionSet
		createdSet := createTestSecretQuestionSet(t, store, setDetails)
		
		assert.NotEmpty(t, createdSet.ID)
		assert.NotZero(t, createdSet.CreatedAt)
		assert.NotZero(t, createdSet.UpdatedAt)
		assert.Equal(t, createdSet.CreatedAt, createdSet.UpdatedAt)

		// Verify by fetching
		fetchedSet, err := store.GetSecretQuestionSet(ctx, createdSet.ID)
		require.NoError(t, err)
		assert.Equal(t, createdSet.ID, fetchedSet.ID)
		assert.Equal(t, createdSet.SecretAssignmentID, fetchedSet.SecretAssignmentID) // Foreign key
		assert.Equal(t, setDetails.Threshold, fetchedSet.Threshold)
		assert.Equal(t, setDetails.TotalQuestions, fetchedSet.TotalQuestions)
		assert.Equal(t, setDetails.TimelockRound, fetchedSet.TimelockRound)
		assert.Equal(t, setDetails.EncryptedBlob, fetchedSet.EncryptedBlob)
	})

	t.Run("creation with empty ID", func(t *testing.T) {
		setDetails := &models.SecretQuestionSet{
			Threshold:      1,
			TotalQuestions: 1,
			TimelockRound:  200,
			EncryptedBlob:  []byte("anotherblob"),
		}
		createdSet := createTestSecretQuestionSet(t, store, setDetails) // ID will be generated by helper/Create method
		
		assert.NotEmpty(t, createdSet.ID, "ID should be generated")
		assert.NotZero(t, createdSet.CreatedAt)
		assert.NotZero(t, createdSet.UpdatedAt)
	})

	t.Run("CreatedAt and UpdatedAt are set", func(t *testing.T) {
		startTime := time.Now().Truncate(time.Second) 
		setDetails := &models.SecretQuestionSet{
			Threshold:      1,
			TotalQuestions: 1,
			TimelockRound:  300,
			EncryptedBlob:  []byte("timeblob"),
		}
		createdSet := createTestSecretQuestionSet(t, store, setDetails)

		assert.NotZero(t, createdSet.CreatedAt)
		assert.NotZero(t, createdSet.UpdatedAt)
		assert.False(t, createdSet.CreatedAt.Before(startTime), "CreatedAt should be after or equal to test start time")
		assert.False(t, createdSet.UpdatedAt.Before(startTime), "UpdatedAt should be after or equal to test start time")
		assert.Equal(t, createdSet.CreatedAt, createdSet.UpdatedAt)
	})

	// Based on the schema, SecretAssignmentID is NOT NULL (handled by createTestSecretQuestionSet)
	// EncryptedBlob is NOT NULL.
	t.Run("missing mandatory fields", func(t *testing.T) {
		// SecretAssignmentID is automatically created by the helper, so we can't easily test its absence here
		// without a more direct call to store.CreateSecretQuestionSet with a nil/empty SecretAssignmentID.
		// For now, this sub-test focuses on EncryptedBlob.

		t.Run("missing EncryptedBlob", func(t *testing.T) {
			user := createTestUser(t, store)
			secret := createTestSecret(t, store, user.ID)
			recipient := createTestRecipient(t, store, user.ID)
			assignment := createTestSecretAssignment(t, store, user.ID, secret.ID, recipient.ID)

			set := &models.SecretQuestionSet{
				ID:                 uuid.New().String(),
				SecretAssignmentID: assignment.ID,
				Threshold:          1,
				TotalQuestions:     1,
				TimelockRound:      1,
				// EncryptedBlob is missing
			}
			err := store.CreateSecretQuestionSet(ctx, set)
			// This should fail due to NOT NULL constraint in DB (assuming schema has EncryptedBlob NOT NULL)
			// The current schema in sqlite.go for secret_question_sets does not explicitly state NOT NULL for encrypted_blob
			// Let's check the actual migration 'add_secret_questions.go'
			// If it's nullable, this test is invalid. If NOT NULL, it's valid.
			// Assuming it's NOT NULL based on typical design for such fields.
			// The migration `add_secret_questions.go` indeed has `encrypted_blob BLOB NOT NULL`. So this test is valid.
			require.Error(t, err) 
		})
	})
}

// TestGetSecretQuestionSet will be implemented here
func TestGetSecretQuestionSet(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	t.Run("successful retrieval", func(t *testing.T) {
		setDetails := &models.SecretQuestionSet{
			Threshold:      2,
			TotalQuestions: 3,
			TimelockRound:  100,
			EncryptedBlob:  []byte("retrievetest"),
		}
		createdSet := createTestSecretQuestionSet(t, store, setDetails)

		fetchedSet, err := store.GetSecretQuestionSet(ctx, createdSet.ID)
		require.NoError(t, err)
		assert.Equal(t, createdSet.ID, fetchedSet.ID)
		assert.Equal(t, createdSet.SecretAssignmentID, fetchedSet.SecretAssignmentID)
		assert.Equal(t, setDetails.Threshold, fetchedSet.Threshold)
		assert.Equal(t, setDetails.TotalQuestions, fetchedSet.TotalQuestions)
		assert.Equal(t, setDetails.TimelockRound, fetchedSet.TimelockRound)
		assert.Equal(t, setDetails.EncryptedBlob, fetchedSet.EncryptedBlob)
		assert.Equal(t, createdSet.CreatedAt.Unix(), fetchedSet.CreatedAt.Unix()) 
		assert.Equal(t, createdSet.UpdatedAt.Unix(), fetchedSet.UpdatedAt.Unix()) 
	})

	t.Run("non-existent set", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		_, err := store.GetSecretQuestionSet(ctx, nonExistentID)
		assert.ErrorIs(t, err, storage.ErrNotFound)
	})
}

// TestGetSecretQuestionSetByAssignmentID will be implemented here
func TestGetSecretQuestionSetByAssignmentID(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	// Create a unique assignment for this test
	user := createTestUser(t, store)
	secret := createTestSecret(t, store, user.ID)
	recipient := createTestRecipient(t, store, user.ID)
	assignment := createTestSecretAssignment(t, store, user.ID, secret.ID, recipient.ID)

	setDetails := &models.SecretQuestionSet{
		SecretAssignmentID: assignment.ID, // Use the specific assignment ID
		Threshold:          1,
		TotalQuestions:     2,
		TimelockRound:      250,
		EncryptedBlob:      []byte("byassignmentid"),
	}
	// Manually create the SecretQuestionSet with the specific assignment.ID, not using the main helper here.
	fullSet := &models.SecretQuestionSet{
		ID:                 uuid.New().String(),
		SecretAssignmentID: assignment.ID,
		Threshold:          setDetails.Threshold,
		TotalQuestions:     setDetails.TotalQuestions,
		TimelockRound:      setDetails.TimelockRound,
		EncryptedBlob:      setDetails.EncryptedBlob,
	}
	err := store.CreateSecretQuestionSet(context.Background(), fullSet)
	require.NoError(t, err)


	t.Run("successful retrieval by assignment ID", func(t *testing.T) {
		fetchedSet, err := store.GetSecretQuestionSetByAssignmentID(ctx, assignment.ID)
		require.NoError(t, err)
		assert.Equal(t, fullSet.ID, fetchedSet.ID)
		assert.Equal(t, assignment.ID, fetchedSet.SecretAssignmentID)
		assert.Equal(t, fullSet.Threshold, fetchedSet.Threshold)
		assert.Equal(t, fullSet.TotalQuestions, fetchedSet.TotalQuestions)
		assert.Equal(t, fullSet.TimelockRound, fetchedSet.TimelockRound)
		assert.Equal(t, fullSet.EncryptedBlob, fetchedSet.EncryptedBlob)
	})

	t.Run("non-existent assignment ID", func(t *testing.T) {
		nonExistentAssignmentID := uuid.New().String()
		_, err := store.GetSecretQuestionSetByAssignmentID(ctx, nonExistentAssignmentID)
		assert.ErrorIs(t, err, storage.ErrNotFound)
	})
}

// TestUpdateSecretQuestionSet will be implemented here
func TestUpdateSecretQuestionSet(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	initialDetails := &models.SecretQuestionSet{
		Threshold:      2,
		TotalQuestions: 3,
		TimelockRound:  100,
		EncryptedBlob:  []byte("initialblob"),
	}
	initialSet := createTestSecretQuestionSet(t, store, initialDetails)
	
	initialCreatedAt := initialSet.CreatedAt
	initialUpdatedAt := initialSet.UpdatedAt
	assert.Equal(t, initialCreatedAt.Unix(), initialUpdatedAt.Unix())


	t.Run("successful update", func(t *testing.T) {
		time.Sleep(10 * time.Millisecond)

		update := &models.SecretQuestionSet{
			ID:                 initialSet.ID, // Must provide ID for update
			SecretAssignmentID: initialSet.SecretAssignmentID, // Should not change and is not part of update SET clause
			Threshold:          3,
			TotalQuestions:     5,
			TimelockRound:      200,
			EncryptedBlob:      []byte("updatedblob"),
		}

		err := store.UpdateSecretQuestionSet(ctx, update)
		require.NoError(t, err)

		fetchedSet, err := store.GetSecretQuestionSet(ctx, initialSet.ID)
		require.NoError(t, err)

		assert.Equal(t, initialSet.ID, fetchedSet.ID)
		assert.Equal(t, initialSet.SecretAssignmentID, fetchedSet.SecretAssignmentID) // SecretAssignmentID should not change
		assert.Equal(t, update.Threshold, fetchedSet.Threshold)
		assert.Equal(t, update.TotalQuestions, fetchedSet.TotalQuestions)
		assert.Equal(t, update.TimelockRound, fetchedSet.TimelockRound)
		assert.Equal(t, update.EncryptedBlob, fetchedSet.EncryptedBlob)
		
		assert.Equal(t, initialCreatedAt.Unix(), fetchedSet.CreatedAt.Unix(), "CreatedAt should not change on update")
		assert.True(t, fetchedSet.UpdatedAt.After(initialUpdatedAt), "UpdatedAt should be updated")
	})

	t.Run("updating a non-existent set", func(t *testing.T) {
		nonExistentSet := &models.SecretQuestionSet{
			ID:                 uuid.New().String(), 
			SecretAssignmentID: uuid.New().String(), // Needs a valid FK if it were to be created
			Threshold:          1,
			TotalQuestions:     1,
			TimelockRound:      1,
			EncryptedBlob:      []byte("nonexistent"),
		}
		err := store.UpdateSecretQuestionSet(ctx, nonExistentSet)
		require.Error(t, err) // UpdateSecretQuestionSet checks RowsAffected and returns ErrNotFound
		assert.ErrorIs(t, err, storage.ErrNotFound) 

		_, getErr := store.GetSecretQuestionSet(ctx, nonExistentSet.ID)
		assert.ErrorIs(t, getErr, storage.ErrNotFound, "Set should still not exist after failed update attempt")
	})
}


// TestDeleteSecretQuestionSet will be implemented here
func TestDeleteSecretQuestionSet(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	setToDeleteDetails := &models.SecretQuestionSet{
		Threshold:      1,
		TotalQuestions: 1,
		TimelockRound:  50,
		EncryptedBlob:  []byte("deleteme"),
	}
	setToDelete := createTestSecretQuestionSet(t, store, setToDeleteDetails)

	t.Run("successful deletion", func(t *testing.T) {
		err := store.DeleteSecretQuestionSet(ctx, setToDelete.ID)
		require.NoError(t, err)

		_, err = store.GetSecretQuestionSet(ctx, setToDelete.ID)
		assert.ErrorIs(t, err, storage.ErrNotFound, "Set should be deleted")
	})

	t.Run("deleting a non-existent set", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		err := store.DeleteSecretQuestionSet(ctx, nonExistentID)
		assert.ErrorIs(t, err, storage.ErrNotFound)
	})
}

// TestListSecretQuestionSetsNeedingReencryption will be implemented here
func TestListSecretQuestionSetsNeedingReencryption(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	nowUnix := time.Now().Unix()

	set1Details := &models.SecretQuestionSet{Threshold: 1, TotalQuestions: 1, EncryptedBlob: []byte("blob1"), TimelockRound: uint64(nowUnix - 1000)}
	set1_past := createTestSecretQuestionSet(t, store, set1Details)

	set2Details := &models.SecretQuestionSet{Threshold: 1, TotalQuestions: 1, EncryptedBlob: []byte("blob2"), TimelockRound: uint64(nowUnix - 1)}
	set2_just_past := createTestSecretQuestionSet(t, store, set2Details)

	set3Details := &models.SecretQuestionSet{Threshold: 1, TotalQuestions: 1, EncryptedBlob: []byte("blob3"), TimelockRound: uint64(nowUnix + 10000)}
	set3_future := createTestSecretQuestionSet(t, store, set3Details)
	
	set4Details := &models.SecretQuestionSet{Threshold: 1, TotalQuestions: 1, EncryptedBlob: []byte("blob4"), TimelockRound: 0}
	set4_zero_timelock := createTestSecretQuestionSet(t, store, set4Details)


	t.Run("list sets needing re-encryption (no margin)", func(t *testing.T) {
		time.Sleep(5 * time.Millisecond) 
		
		sets, err := store.ListSecretQuestionSetsNeedingReencryption(ctx, 0) 
		require.NoError(t, err)

		var foundIDs []string
		for _, s := range sets { foundIDs = append(foundIDs, s.ID) }

		assert.Contains(t, foundIDs, set1_past.ID)
		assert.Contains(t, foundIDs, set2_just_past.ID)
		assert.Contains(t, foundIDs, set4_zero_timelock.ID)
		assert.NotContains(t, foundIDs, set3_future.ID)
		assert.Len(t, sets, 3)
	})

	t.Run("list with positive safe margin", func(t *testing.T) {
		slightlyFutureTime := time.Now().Unix() + 50
		set5Details := &models.SecretQuestionSet{Threshold: 1, TotalQuestions: 1, EncryptedBlob: []byte("blob5"), TimelockRound: uint64(slightlyFutureTime)}
		set5_slightly_future := createTestSecretQuestionSet(t, store, set5Details)
		
		time.Sleep(5 * time.Millisecond) 
		safeMargin := int64(100) 
		sets, err := store.ListSecretQuestionSetsNeedingReencryption(ctx, safeMargin)
		require.NoError(t, err)

		var foundIDs []string
		for _, s := range sets { foundIDs = append(foundIDs, s.ID) }

		assert.Contains(t, foundIDs, set1_past.ID)
		assert.Contains(t, foundIDs, set2_just_past.ID)
		assert.Contains(t, foundIDs, set4_zero_timelock.ID)
		assert.Contains(t, foundIDs, set5_slightly_future.ID)
		assert.NotContains(t, foundIDs, set3_future.ID) 
		assert.Len(t, sets, 4)
	})
	
	t.Run("list with negative safe margin", func(t *testing.T) {
		negativeMargin := int64(-500) 
		time.Sleep(5 * time.Millisecond)
		sets, err := store.ListSecretQuestionSetsNeedingReencryption(ctx, negativeMargin)
		require.NoError(t, err)

		var foundIDs []string
		for _, s := range sets { foundIDs = append(foundIDs, s.ID) }
		
		assert.Contains(t, foundIDs, set1_past.ID)
		assert.NotContains(t, foundIDs, set2_just_past.ID)
		assert.Contains(t, foundIDs, set4_zero_timelock.ID)
		assert.Len(t, sets, 2)
	})

	t.Run("no sets in DB", func(t *testing.T) {
		emptyStore := newTestStore(t) 
		sets, err := emptyStore.ListSecretQuestionSetsNeedingReencryption(ctx, 0)
		require.NoError(t, err)
		assert.Empty(t, sets)
	})
}

// Additional helper to compare SecretQuestionSet ignoring time fields for some tests
func assertSecretQuestionSetEqualValues(t *testing.T, expected, actual *models.SecretQuestionSet, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Equal(t, expected.ID, actual.ID, msgAndArgs...)
	assert.Equal(t, expected.SecretAssignmentID, actual.SecretAssignmentID, msgAndArgs...)
	assert.Equal(t, expected.Threshold, actual.Threshold, msgAndArgs...)
	assert.Equal(t, expected.TotalQuestions, actual.TotalQuestions, msgAndArgs...)
	assert.Equal(t, expected.TimelockRound, actual.TimelockRound, msgAndArgs...)
	assert.Equal(t, expected.EncryptedBlob, actual.EncryptedBlob, msgAndArgs...)
}
