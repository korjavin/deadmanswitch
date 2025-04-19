package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
)

// MockRepository is a mock implementation of the storage.Repository interface
type MockRepository struct {
	sessions map[string]*models.Session
	users    map[string]*models.User
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		sessions: make(map[string]*models.Session),
		users:    make(map[string]*models.User),
	}
}

func (m *MockRepository) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	session, ok := m.sessions[token]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return session, nil
}

func (m *MockRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return user, nil
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockRepository) UpdateSessionActivity(ctx context.Context, id string) error {
	for _, session := range m.sessions {
		if session.ID == id {
			session.LastActivity = time.Now()
			return nil
		}
	}
	return storage.ErrNotFound
}

// Implement other methods of the Repository interface with empty implementations
func (m *MockRepository) CreateUser(ctx context.Context, user *models.User) error { return nil }
func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) { return nil, nil }
func (m *MockRepository) GetUserByTelegramID(ctx context.Context, telegramID string) (*models.User, error) { return nil, nil }
func (m *MockRepository) DeleteUser(ctx context.Context, id string) error { return nil }
func (m *MockRepository) ListUsers(ctx context.Context) ([]*models.User, error) { return nil, nil }
func (m *MockRepository) CreateSecret(ctx context.Context, secret *models.Secret) error { return nil }
func (m *MockRepository) GetSecretByID(ctx context.Context, id string) (*models.Secret, error) { return nil, nil }
func (m *MockRepository) ListSecretsByUserID(ctx context.Context, userID string) ([]*models.Secret, error) { return nil, nil }
func (m *MockRepository) UpdateSecret(ctx context.Context, secret *models.Secret) error { return nil }
func (m *MockRepository) DeleteSecret(ctx context.Context, id string) error { return nil }
func (m *MockRepository) CreateRecipient(ctx context.Context, recipient *models.Recipient) error { return nil }
func (m *MockRepository) GetRecipientByID(ctx context.Context, id string) (*models.Recipient, error) { return nil, nil }
func (m *MockRepository) ListRecipientsByUserID(ctx context.Context, userID string) ([]*models.Recipient, error) { return nil, nil }
func (m *MockRepository) UpdateRecipient(ctx context.Context, recipient *models.Recipient) error { return nil }
func (m *MockRepository) DeleteRecipient(ctx context.Context, id string) error { return nil }
func (m *MockRepository) CreateSecretAssignment(ctx context.Context, assignment *models.SecretAssignment) error { return nil }
func (m *MockRepository) GetSecretAssignmentByID(ctx context.Context, id string) (*models.SecretAssignment, error) { return nil, nil }
func (m *MockRepository) ListSecretAssignmentsBySecretID(ctx context.Context, secretID string) ([]*models.SecretAssignment, error) { return nil, nil }
func (m *MockRepository) ListSecretAssignmentsByRecipientID(ctx context.Context, recipientID string) ([]*models.SecretAssignment, error) { return nil, nil }
func (m *MockRepository) ListSecretAssignmentsByUserID(ctx context.Context, userID string) ([]*models.SecretAssignment, error) { return nil, nil }
func (m *MockRepository) DeleteSecretAssignment(ctx context.Context, id string) error { return nil }
func (m *MockRepository) CreatePingHistory(ctx context.Context, ping *models.PingHistory) error { return nil }
func (m *MockRepository) UpdatePingHistory(ctx context.Context, ping *models.PingHistory) error { return nil }
func (m *MockRepository) GetLatestPingByUserID(ctx context.Context, userID string) (*models.PingHistory, error) { return nil, nil }
func (m *MockRepository) ListPingHistoryByUserID(ctx context.Context, userID string) ([]*models.PingHistory, error) { return nil, nil }
func (m *MockRepository) CreatePingVerification(ctx context.Context, verification *models.PingVerification) error { return nil }
func (m *MockRepository) GetPingVerificationByCode(ctx context.Context, code string) (*models.PingVerification, error) { return nil, nil }
func (m *MockRepository) UpdatePingVerification(ctx context.Context, verification *models.PingVerification) error { return nil }
func (m *MockRepository) CreateDeliveryEvent(ctx context.Context, event *models.DeliveryEvent) error { return nil }
func (m *MockRepository) ListDeliveryEventsByUserID(ctx context.Context, userID string) ([]*models.DeliveryEvent, error) { return nil, nil }
func (m *MockRepository) CreateAuditLog(ctx context.Context, log *models.AuditLog) error { return nil }
func (m *MockRepository) ListAuditLogsByUserID(ctx context.Context, userID string) ([]*models.AuditLog, error) { return nil, nil }
func (m *MockRepository) CreateSession(ctx context.Context, session *models.Session) error { return nil }
func (m *MockRepository) DeleteSession(ctx context.Context, id string) error { return nil }
func (m *MockRepository) DeleteExpiredSessions(ctx context.Context) error { return nil }
func (m *MockRepository) GetUsersForPinging(ctx context.Context) ([]*models.User, error) { return nil, nil }
func (m *MockRepository) GetUsersWithExpiredPings(ctx context.Context) ([]*models.User, error) { return nil, nil }
func (m *MockRepository) BeginTx(ctx context.Context) (storage.Transaction, error) { return nil, nil }
func (m *MockRepository) ListPasskeysByUserID(ctx context.Context, userID string) ([]*models.Passkey, error) { return nil, nil }
func (m *MockRepository) ListPasskeys(ctx context.Context) ([]*models.Passkey, error) { return nil, nil }
func (m *MockRepository) GetPasskeyByCredentialID(ctx context.Context, credentialID []byte) (*models.Passkey, error) { return nil, nil }
func (m *MockRepository) CreatePasskey(ctx context.Context, passkey *models.Passkey) error { return nil }
func (m *MockRepository) UpdatePasskey(ctx context.Context, passkey *models.Passkey) error { return nil }
func (m *MockRepository) GetPasskeyByID(ctx context.Context, id string) (*models.Passkey, error) { return nil, nil }
func (m *MockRepository) DeletePasskey(ctx context.Context, id string) error { return nil }
func (m *MockRepository) DeletePasskeysByUserID(ctx context.Context, userID string) error { return nil }

func TestAuth(t *testing.T) {
	// Create a mock repository
	repo := NewMockRepository()

	// Create a test user
	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
	}
	repo.users[user.ID] = user

	// Create a valid session
	validSession := &models.Session{
		ID:           "session123",
		UserID:       user.ID,
		Token:        "valid-token",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		LastActivity: time.Now(),
	}
	repo.sessions[validSession.Token] = validSession

	// Create an expired session
	expiredSession := &models.Session{
		ID:           "session456",
		UserID:       user.ID,
		Token:        "expired-token",
		ExpiresAt:    time.Now().Add(-1 * time.Hour),
		LastActivity: time.Now().Add(-2 * time.Hour),
	}
	repo.sessions[expiredSession.Token] = expiredSession

	// Create a session with an invalid user ID
	invalidUserSession := &models.Session{
		ID:           "session789",
		UserID:       "invalid-user",
		Token:        "invalid-user-token",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		LastActivity: time.Now(),
	}
	repo.sessions[invalidUserSession.Token] = invalidUserSession

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the user is in the context
		user, ok := GetUserFromContext(r)
		if !ok {
			t.Error("Expected user in context, got none")
		} else if user.ID != "user123" {
			t.Errorf("Expected user ID 'user123', got '%s'", user.ID)
		}

		// Check if the session is in the context
		session, ok := r.Context().Value("session").(*models.Session)
		if !ok {
			t.Error("Expected session in context, got none")
		} else if session.ID != "session123" {
			t.Errorf("Expected session ID 'session123', got '%s'", session.ID)
		}

		w.WriteHeader(http.StatusOK)
	})

	// Create the middleware
	authMiddleware := Auth(repo)

	// Test with valid session
	t.Run("Valid Session", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		// Add the session cookie
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: validSession.Token,
		})

		// Create a response recorder
		rr := httptest.NewRecorder()

		// Serve the request
		authMiddleware(testHandler).ServeHTTP(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
		}
	})

	// Test with expired session
	t.Run("Expired Session", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		// Add the session cookie
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: expiredSession.Token,
		})

		// Create a response recorder
		rr := httptest.NewRecorder()

		// Serve the request
		authMiddleware(testHandler).ServeHTTP(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusSeeOther {
			t.Errorf("Expected status code %d, got %d", http.StatusSeeOther, status)
		}

		// Check the redirect location
		if location := rr.Header().Get("Location"); location != "/login" {
			t.Errorf("Expected redirect to '/login', got '%s'", location)
		}
	})

	// Test with invalid user ID
	t.Run("Invalid User ID", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		// Add the session cookie
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: invalidUserSession.Token,
		})

		// Create a response recorder
		rr := httptest.NewRecorder()

		// Serve the request
		authMiddleware(testHandler).ServeHTTP(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusSeeOther {
			t.Errorf("Expected status code %d, got %d", http.StatusSeeOther, status)
		}

		// Check the redirect location
		if location := rr.Header().Get("Location"); location != "/login" {
			t.Errorf("Expected redirect to '/login', got '%s'", location)
		}
	})

	// Test with no session cookie
	t.Run("No Session Cookie", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		// Create a response recorder
		rr := httptest.NewRecorder()

		// Serve the request
		authMiddleware(testHandler).ServeHTTP(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusSeeOther {
			t.Errorf("Expected status code %d, got %d", http.StatusSeeOther, status)
		}

		// Check the redirect location
		if location := rr.Header().Get("Location"); location != "/login" {
			t.Errorf("Expected redirect to '/login', got '%s'", location)
		}
	})

	// Test with invalid session token
	t.Run("Invalid Session Token", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		// Add the session cookie
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: "invalid-token",
		})

		// Create a response recorder
		rr := httptest.NewRecorder()

		// Serve the request
		authMiddleware(testHandler).ServeHTTP(rr, req)

		// Check the status code
		if status := rr.Code; status != http.StatusSeeOther {
			t.Errorf("Expected status code %d, got %d", http.StatusSeeOther, status)
		}

		// Check the redirect location
		if location := rr.Header().Get("Location"); location != "/login" {
			t.Errorf("Expected redirect to '/login', got '%s'", location)
		}
	})
}

func TestGetUserFromContext(t *testing.T) {
	// Create a test user
	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
	}

	// Create a context with the user
	ctx := context.WithValue(context.Background(), "user", user)
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(ctx)

	// Get the user from the context
	contextUser, ok := GetUserFromContext(req)
	if !ok {
		t.Error("Expected user in context, got none")
	}
	if contextUser.ID != user.ID {
		t.Errorf("Expected user ID '%s', got '%s'", user.ID, contextUser.ID)
	}
	if contextUser.Email != user.Email {
		t.Errorf("Expected user email '%s', got '%s'", user.Email, contextUser.Email)
	}

	// Test with no user in context
	req, err = http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	contextUser, ok = GetUserFromContext(req)
	if ok {
		t.Error("Expected no user in context, got one")
	}
	if contextUser != nil {
		t.Errorf("Expected nil user, got %v", contextUser)
	}
}
