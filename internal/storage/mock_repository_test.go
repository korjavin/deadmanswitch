package storage

import (
	"context"
	"testing"

	"github.com/korjavin/deadmanswitch/internal/models"
)

// Helper function for testing CRUD operations on a generic entity
func testMockRepositoryCRUDOperations(t *testing.T,
	createFn func(ctx context.Context, mockRepo *MockRepository, id string) (interface{}, error),
	getFn func(ctx context.Context, mockRepo *MockRepository, id string) (interface{}, error),
	listFn func(ctx context.Context, mockRepo *MockRepository, userId string) (interface{}, error),
	updateFn func(ctx context.Context, mockRepo *MockRepository, entity interface{}) error,
	deleteFn func(ctx context.Context, mockRepo *MockRepository, id string) error,
	entityName string, // For error messages
	entityID string,
	userID string,
	updateFieldName string, // Name of the field to update for testing
	updateFieldValue string, // New value for the field
	getEntityField func(entity interface{}) string, // Function to get the field value from the entity
	getEntitiesLength func(entities interface{}) int, // Function to get the length of entities list
) {
	repo := NewMockRepository()
	ctx := context.Background()

	// Test Create
	entity, err := createFn(ctx, repo, entityID)
	if err != nil {
		t.Fatalf("Create%s failed: %v", entityName, err)
	}

	// Test GetByID
	retrievedEntity, err := getFn(ctx, repo, entityID)
	if err != nil {
		t.Fatalf("Get%sByID failed: %v", entityName, err)
	}
	if retrievedEntity == nil {
		t.Fatalf("Expected non-nil %s", entityName)
	}

	// Test List
	entities, err := listFn(ctx, repo, userID)
	if err != nil {
		t.Fatalf("List%ssByUserID failed: %v", entityName, err)
	}
	if getEntitiesLength(entities) != 1 {
		t.Errorf("Expected 1 %s, got %d", entityName, getEntitiesLength(entities))
	}

	// Test Update
	if err = updateFn(ctx, repo, entity); err != nil {
		t.Fatalf("Update%s failed: %v", entityName, err)
	}

	// Verify the update
	retrievedEntity, err = getFn(ctx, repo, entityID)
	if err != nil {
		t.Fatalf("Get%sByID failed after update: %v", entityName, err)
	}
	if getEntityField(retrievedEntity) != updateFieldValue {
		t.Errorf("Expected updated %s to be '%s', got '%s'",
			updateFieldName, updateFieldValue, getEntityField(retrievedEntity))
	}

	// Test Delete
	if err = deleteFn(ctx, repo, entityID); err != nil {
		t.Fatalf("Delete%s failed: %v", entityName, err)
	}

	// Verify the deletion
	entities, err = listFn(ctx, repo, userID)
	if err != nil {
		t.Fatalf("List%ssByUserID failed after deletion: %v", entityName, err)
	}
	if getEntitiesLength(entities) != 0 {
		t.Errorf("Expected 0 %ss after deletion, got %d", entityName, getEntitiesLength(entities))
	}
}

// Helper function for testing user fetching methods with similar patterns
func testUserFetchingMethod(t *testing.T, methodName string,
	fetchFn func(ctx context.Context, repo *MockRepository) ([]*models.User, error),
	setupUsersFn func(repo *MockRepository)) {

	repo := NewMockRepository()
	ctx := context.Background()

	// Add users using the provided setup function
	setupUsersFn(repo)

	// Test the fetch method
	users, err := fetchFn(ctx, repo)
	if err != nil {
		t.Fatalf("%s failed: %v", methodName, err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users for %s, got %d", methodName, len(users))
	}
}

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
	testMockRepositoryCRUDOperations(t,
		func(ctx context.Context, mockRepo *MockRepository, id string) (interface{}, error) {
			user := &models.User{
				ID:    id,
				Email: "user1@example.com",
			}
			return user, mockRepo.CreateUser(ctx, user)
		},
		func(ctx context.Context, mockRepo *MockRepository, id string) (interface{}, error) {
			return mockRepo.GetUserByID(ctx, id)
		},
		func(ctx context.Context, mockRepo *MockRepository, userId string) (interface{}, error) {
			return mockRepo.ListUsers(ctx)
		},
		func(ctx context.Context, mockRepo *MockRepository, entity interface{}) error {
			user := entity.(*models.User)
			user.Email = "updated@example.com"
			return mockRepo.UpdateUser(ctx, user)
		},
		func(ctx context.Context, mockRepo *MockRepository, id string) error {
			return mockRepo.DeleteUser(ctx, id)
		},
		"User", "user1", "user1", "Email", "updated@example.com",
		func(entity interface{}) string {
			return entity.(*models.User).Email
		},
		func(entities interface{}) int {
			return len(entities.([]*models.User))
		},
	)
}

func TestMockRepositorySecretMethods(t *testing.T) {
	testMockRepositoryCRUDOperations(t,
		func(ctx context.Context, mockRepo *MockRepository, id string) (interface{}, error) {
			secret := &models.Secret{
				ID:            id,
				UserID:        "user1",
				Name:          "Test Secret",
				EncryptedData: "encrypted-data",
			}
			return secret, mockRepo.CreateSecret(ctx, secret)
		},
		func(ctx context.Context, mockRepo *MockRepository, id string) (interface{}, error) {
			return mockRepo.GetSecretByID(ctx, id)
		},
		func(ctx context.Context, mockRepo *MockRepository, userId string) (interface{}, error) {
			return mockRepo.ListSecretsByUserID(ctx, userId)
		},
		func(ctx context.Context, mockRepo *MockRepository, entity interface{}) error {
			secret := entity.(*models.Secret)
			secret.Name = "Updated Secret"
			return mockRepo.UpdateSecret(ctx, secret)
		},
		func(ctx context.Context, mockRepo *MockRepository, id string) error {
			return mockRepo.DeleteSecret(ctx, id)
		},
		"Secret", "secret1", "user1", "Name", "Updated Secret",
		func(entity interface{}) string {
			return entity.(*models.Secret).Name
		},
		func(entities interface{}) int {
			return len(entities.([]*models.Secret))
		},
	)
}

func TestMockRepositoryRecipientMethods(t *testing.T) {
	testMockRepositoryCRUDOperations(t,
		func(ctx context.Context, mockRepo *MockRepository, id string) (interface{}, error) {
			recipient := &models.Recipient{
				ID:     id,
				UserID: "user1",
				Name:   "Test Recipient",
				Email:  "recipient@example.com",
			}
			return recipient, mockRepo.CreateRecipient(ctx, recipient)
		},
		func(ctx context.Context, mockRepo *MockRepository, id string) (interface{}, error) {
			return mockRepo.GetRecipientByID(ctx, id)
		},
		func(ctx context.Context, mockRepo *MockRepository, userId string) (interface{}, error) {
			return mockRepo.ListRecipientsByUserID(ctx, userId)
		},
		func(ctx context.Context, mockRepo *MockRepository, entity interface{}) error {
			recipient := entity.(*models.Recipient)
			recipient.Name = "Updated Recipient"
			return mockRepo.UpdateRecipient(ctx, recipient)
		},
		func(ctx context.Context, mockRepo *MockRepository, id string) error {
			return mockRepo.DeleteRecipient(ctx, id)
		},
		"Recipient", "recipient1", "user1", "Name", "Updated Recipient",
		func(entity interface{}) string {
			return entity.(*models.Recipient).Name
		},
		func(entities interface{}) int {
			return len(entities.([]*models.Recipient))
		},
	)
}

func TestMockRepositoryGetUsersForPinging(t *testing.T) {
	testUserFetchingMethod(t, "GetUsersForPinging",
		func(ctx context.Context, repo *MockRepository) ([]*models.User, error) {
			return repo.GetUsersForPinging(ctx)
		},
		func(repo *MockRepository) {
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
		})
}

func TestMockRepositoryGetUsersWithExpiredPings(t *testing.T) {
	testUserFetchingMethod(t, "GetUsersWithExpiredPings",
		func(ctx context.Context, repo *MockRepository) ([]*models.User, error) {
			return repo.GetUsersWithExpiredPings(ctx)
		},
		func(repo *MockRepository) {
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
		})
}
