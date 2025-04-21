package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/korjavin/deadmanswitch/internal/config"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
)

// mockRepository is a mock implementation of storage.Repository for testing
type mockRepository struct {
	users    map[string]*models.User
	sessions map[string]*models.Session
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		users:    make(map[string]*models.User),
		sessions: make(map[string]*models.Session),
	}
}

// Implement the necessary methods from storage.Repository
func (m *mockRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return user, nil
}

func (m *mockRepository) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	session, ok := m.sessions[token]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return session, nil
}

func (m *mockRepository) UpdateUser(ctx context.Context, user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockRepository) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	// No-op for testing
	return nil
}

// Add a user to the mock repository
func (m *mockRepository) AddUser(user *models.User) {
	m.users[user.ID] = user
}

// Add a session to the mock repository
func (m *mockRepository) AddSession(session *models.Session) {
	m.sessions[session.Token] = session
}

// TestServerNew tests the new server implementation
func TestServerNew(t *testing.T) {
	// Create a mock repository
	repo := newMockRepository()

	// Create a test user and session
	user := &models.User{
		ID:           "test-user-id",
		Email:        "test@example.com",
		LastActivity: time.Now(),
		CreatedAt:    time.Now().Add(-24 * time.Hour), // 1 day ago
	}
	session := &models.Session{
		ID:        "test-session-id",
		UserID:    user.ID,
		Token:     "test-session-token",
		ExpiresAt: time.Now().Add(24 * time.Hour), // 1 day in the future
	}
	repo.AddUser(user)
	repo.AddSession(session)

	// Create a test config
	cfg := &config.Config{
		BaseDomain: "localhost",
		Debug:      true,
	}

	// Create a new server
	server := NewServerWithRouter(cfg, repo, nil, nil, nil)

	// Create a test server
	testServer := httptest.NewServer(server.router.Handler())
	defer testServer.Close()

	// Test the home page
	t.Run("Home Page", func(t *testing.T) {
		req, err := http.NewRequest("GET", testServer.URL+"/", nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})

	// Test a protected route without authentication
	t.Run("Protected Route Without Auth", func(t *testing.T) {
		req, err := http.NewRequest("GET", testServer.URL+"/dashboard", nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		// Should redirect to login
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("Expected status code %d, got %d", http.StatusSeeOther, resp.StatusCode)
		}
		if resp.Header.Get("Location") != "/login" {
			t.Errorf("Expected redirect to /login, got %s", resp.Header.Get("Location"))
		}
	})

	// Test a protected route with authentication
	t.Run("Protected Route With Auth", func(t *testing.T) {
		req, err := http.NewRequest("GET", testServer.URL+"/dashboard", nil)
		if err != nil {
			t.Fatal(err)
		}

		// Add the session cookie
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: session.Token,
		})

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		// Should be successful
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}
	})
}
