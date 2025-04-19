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

	// Create the middleware
	authMiddlewareWrapper := Auth(repo)

	// Create a test handler for the middleware to wrap
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

		w.Write([]byte("test handler called"))
	})

	// Test with valid session
	t.Run("Valid Session", func(t *testing.T) {
		testAuthMiddlewareSessionScenario(t, "Valid Session", validSession.Token, false, testHandler, authMiddlewareWrapper)
	})

	// Test with expired session
	t.Run("Expired Session", func(t *testing.T) {
		testAuthMiddlewareSessionScenario(t, "Expired Session", expiredSession.Token, true, testHandler, authMiddlewareWrapper)
	})

	// Test with invalid user ID
	t.Run("Invalid User ID", func(t *testing.T) {
		testAuthMiddlewareSessionScenario(t, "Invalid User ID", invalidUserSession.Token, true, testHandler, authMiddlewareWrapper)
	})

	// Test with no session cookie
	t.Run("No Session Cookie", func(t *testing.T) {
		testAuthMiddlewareSessionScenario(t, "No Session Cookie", "", true, testHandler, authMiddlewareWrapper)
	})

	// Test with invalid session token
	t.Run("Invalid Session Token", func(t *testing.T) {
		testAuthMiddlewareSessionScenario(t, "Invalid Session Token", "invalid-token", true, testHandler, authMiddlewareWrapper)
	})
}

// testAuthMiddlewareSessionScenario is a helper function to test various session scenarios
func testAuthMiddlewareSessionScenario(
	t *testing.T,
	scenarioName string,
	sessionToken string,
	expectRedirect bool,
	testHandler http.HandlerFunc,
	authMiddlewareWrapper func(http.HandlerFunc) http.HandlerFunc,
) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add the session cookie if provided
	if sessionToken != "" {
		req.AddCookie(&http.Cookie{
			Name:  "session_token",
			Value: sessionToken,
		})
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Serve the request
	authMiddlewareWrapper(testHandler).ServeHTTP(rr, req)

	// Check the status code based on expected behavior
	if expectRedirect {
		if status := rr.Code; status != http.StatusSeeOther {
			t.Errorf("%s: Expected status code %d, got %d", scenarioName, http.StatusSeeOther, status)
		}

		// Check the redirect location
		if location := rr.Header().Get("Location"); location != "/login" {
			t.Errorf("%s: Expected redirect to '/login', got '%s'", scenarioName, location)
		}
	} else {
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("%s: Expected status code %d, got %d", scenarioName, http.StatusOK, status)
		}

		// Check the response body
		expected := "test handler called"
		if rr.Body.String() != expected {
			t.Errorf("%s: Expected body '%s', got '%s'", scenarioName, expected, rr.Body.String())
		}
	}
}

func TestGetUserFromContext(t *testing.T) {
	// Create a test user
	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
	}

	// Update test to use the proper context key type
	ctx := context.WithValue(context.Background(), userContextKey, user)
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
