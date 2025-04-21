package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/web/middleware"
)

// mockRepository is a mock implementation of storage.Repository
type mockRepository struct {
	sessions map[string]*models.Session
	users    map[string]*models.User
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		sessions: make(map[string]*models.Session),
		users:    make(map[string]*models.User),
	}
}

func (m *mockRepository) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	session, ok := m.sessions[token]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return session, nil
}

func (m *mockRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return user, nil
}

func (m *mockRepository) UpdateUser(ctx context.Context, user *models.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockRepository) UpdateSessionActivity(ctx context.Context, sessionID string) error {
	return nil
}

// Add a session to the mock repository
func (m *mockRepository) AddSession(session *models.Session) {
	m.sessions[session.Token] = session
}

// Add a user to the mock repository
func (m *mockRepository) AddUser(user *models.User) {
	m.users[user.ID] = user
}

// TestRouter tests the router functionality
func TestRouter(t *testing.T) {
	// Create a mock repository
	repo := newMockRepository()

	// Create a test user and session
	user := &models.User{
		ID:           "test-user-id",
		Email:        "test@example.com",
		LastActivity: time.Now(),
	}
	session := &models.Session{
		ID:        "test-session-id",
		UserID:    user.ID,
		Token:     "test-session-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	repo.AddUser(user)
	repo.AddSession(session)

	// Create a router
	router := New(repo)

	// Register a test route
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test route"))
	})
	router.RegisterRoutes([]Route{
		GET("test", "/test", testHandler),
	})

	// Create a test server
	server := httptest.NewServer(router.Handler())
	defer server.Close()

	// Test the route
	t.Run("Test Route", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL+"/test", nil)
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

	// Test the auth middleware
	t.Run("Auth Middleware", func(t *testing.T) {
		// Override the GetCurrentTime function for testing
		originalGetCurrentTime := middleware.GetCurrentTime
		defer func() { middleware.GetCurrentTime = originalGetCurrentTime }()
		middleware.GetCurrentTime = func() time.Time {
			return time.Now()
		}

		// Register a protected route
		protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := middleware.GetUserFromContext(r.Context())
			if !ok {
				t.Error("User not found in context")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(user.Email))
		})

		router.mux.Handle("/protected", router.AuthMiddleware()(protectedHandler)).Methods("GET")

		// Test with valid session
		req, err := http.NewRequest("GET", server.URL+"/protected", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: session.Token,
		})

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		// Test with invalid session
		req, err = http.NewRequest("GET", server.URL+"/protected", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: "invalid-token",
		})

		resp, err = http.DefaultClient.Do(req)
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

	// Test route groups
	t.Run("Route Groups", func(t *testing.T) {
		// Create a new router for this test
		router := New(repo)

		// Create a group middleware
		groupMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Group", "true")
				next.ServeHTTP(w, r)
			})
		}

		// Create a group
		group := router.Group("/api", groupMiddleware)

		// Add a route to the group
		group.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Group route"))
		}).Methods("GET")

		// Create a test server
		server := httptest.NewServer(router.Handler())
		defer server.Close()

		// Test the group route
		req, err := http.NewRequest("GET", server.URL+"/api/test", nil)
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
		if resp.Header.Get("X-Group") != "true" {
			t.Errorf("Expected X-Group header to be set")
		}
	})

	// Test route parameters
	t.Run("Route Parameters", func(t *testing.T) {
		// Create a new router for this test
		router := New(repo)

		// Add a route with parameters
		router.mux.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			id := vars["id"]
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(id))
		}).Methods("GET")

		// Create a test server
		server := httptest.NewServer(router.Handler())
		defer server.Close()

		// Test the route with parameters
		req, err := http.NewRequest("GET", server.URL+"/users/123", nil)
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
}
