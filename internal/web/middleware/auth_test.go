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

func TestAuth(t *testing.T) {
	// Create a mock repository
	repo := storage.NewMockRepository()

	// Create a test user
	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
	}
	repo.Users = append(repo.Users, user)

	// Create a valid session
	validSession := &models.Session{
		ID:           "session123",
		UserID:       user.ID,
		Token:        "valid-token",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		LastActivity: time.Now(),
	}
	repo.Sessions = append(repo.Sessions, validSession)

	// Create an expired session
	expiredSession := &models.Session{
		ID:           "session456",
		UserID:       user.ID,
		Token:        "expired-token",
		ExpiresAt:    time.Now().Add(-1 * time.Hour),
		LastActivity: time.Now().Add(-2 * time.Hour),
	}
	repo.Sessions = append(repo.Sessions, expiredSession)

	// Create a session with an invalid user ID
	invalidUserSession := &models.Session{
		ID:           "session789",
		UserID:       "invalid-user",
		Token:        "invalid-user-token",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		LastActivity: time.Now(),
	}
	repo.Sessions = append(repo.Sessions, invalidUserSession)

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
	ctx := context.WithValue(context.Background(), UserContextKey, user)
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
