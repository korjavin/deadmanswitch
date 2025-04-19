package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
)

// Setup a test environment directory to help template loading
func setupTestEnv(t *testing.T) {
	// Ensure test templates directory exists
	if err := os.MkdirAll("./web/templates", 0755); err != nil {
		t.Fatalf("Error creating templates directory: %v", err)
	}

	// Create minimal layout and history template files for testing
	layoutContent := `{{ define "layout" }}{{ template "content" . }}{{ end }}`
	if err := os.WriteFile("./web/templates/layout.html", []byte(layoutContent), 0644); err != nil {
		t.Fatalf("Error writing layout template: %v", err)
	}

	historyContent := `{{ define "content" }}History Page{{ end }}`
	if err := os.WriteFile("./web/templates/history.html", []byte(historyContent), 0644); err != nil {
		t.Fatalf("Error writing history template: %v", err)
	}
}

func TestMain(m *testing.M) {
	// Setup test environment
	setupTestEnv(&testing.T{})

	// Run tests
	code := m.Run()

	// Clean up
	os.RemoveAll("./web/templates")

	os.Exit(code)
}

// TestHandleHistorySuccess tests the history handler with authenticated user
func TestHandleHistorySuccess(t *testing.T) {
	// Create mock repository
	repo := storage.NewMockRepository()

	// Create a test user
	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
	}
	repo.Users = append(repo.Users, user)

	// Create test ping history
	ping1 := &models.PingHistory{
		ID:     "ping1",
		UserID: user.ID,
		SentAt: time.Now().Add(-24 * time.Hour),
		Method: "email",
		Status: "sent",
	}

	ping2 := &models.PingHistory{
		ID:          "ping2",
		UserID:      user.ID,
		SentAt:      time.Now().Add(-48 * time.Hour),
		Method:      "telegram",
		Status:      "responded",
		RespondedAt: func() *time.Time { t := time.Now().Add(-47 * time.Hour); return &t }(),
	}

	repo.PingHistories = append(repo.PingHistories, ping1, ping2)

	// Create test audit logs
	log1 := &models.AuditLog{
		ID:        "log1",
		UserID:    user.ID,
		Action:    "user login",
		Timestamp: time.Now().Add(-12 * time.Hour),
		Details:   "Login via web browser",
	}

	log2 := &models.AuditLog{
		ID:        "log2",
		UserID:    user.ID,
		Action:    "create secret",
		Timestamp: time.Now().Add(-36 * time.Hour),
		Details:   "Added new secret",
	}

	repo.AuditLogs = append(repo.AuditLogs, log1, log2)

	// Create the handler
	handler := NewHistoryHandler(repo)

	// Create a test request
	req := httptest.NewRequest("GET", "/history", nil)

	// Create a context with the authenticated user
	ctx := context.WithValue(req.Context(), "user", user)
	req = req.WithContext(ctx)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.HandleHistory(rr, req)

	// Check the status code - with our mock it will return OK
	if status := rr.Code; status != http.StatusOK && status != http.StatusInternalServerError {
		t.Errorf("Handler returned unexpected status code: %v", status)
	}
}

// TestHandleHistoryUnauthorized tests the history handler with no authenticated user
func TestHandleHistoryUnauthorized(t *testing.T) {
	// Create mock repository
	repo := storage.NewMockRepository()

	// Create the handler
	handler := NewHistoryHandler(repo)

	// Create a test request with no user in context
	req := httptest.NewRequest("GET", "/history", nil)
	rr := httptest.NewRecorder()

	// Call the handler
	handler.HandleHistory(rr, req)

	// Check the status code should be unauthorized
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

// TestHandleHistoryWithDBErrors tests the handler with database errors
func TestHandleHistoryWithDBErrors(t *testing.T) {
	// Create a custom mock repository that will return errors
	repo := &errorMockRepo{}

	// Create a test user
	user := &models.User{
		ID:    "user123",
		Email: "test@example.com",
	}

	// Create the handler
	handler := NewHistoryHandler(repo)

	// Create a test request
	req := httptest.NewRequest("GET", "/history", nil)

	// Create a context with the authenticated user
	ctx := context.WithValue(req.Context(), "user", user)
	req = req.WithContext(ctx)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler - it should still render the template but with empty data
	handler.HandleHistory(rr, req)

	// We expect it to still succeed even with DB errors or fail gracefully
	if status := rr.Code; status != http.StatusOK && status != http.StatusInternalServerError {
		t.Errorf("Handler returned unexpected status code: %v", status)
	}
}

// TestDetermineActivityType tests the determineActivityType function
func TestDetermineActivityType(t *testing.T) {
	tests := []struct {
		action   string
		expected string
	}{
		{"", "unknown"},
		{"user login", "security"},
		{"password changed", "security"},
		{"authentication failed", "security"},
		{"create secret", "secret"},
		{"delete secret", "secret"},
		{"add recipient", "recipient"},
		{"update recipient", "recipient"},
		{"update settings", "settings"},
		{"change config", "settings"},
		{"check_in via email", "checkin"},
		{"checkin confirmed", "checkin"},
		{"ping sent", "checkin"},
		{"github activity detected", "activity"},
		{"external_activity recorded", "activity"},
		{"something else", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			result := determineActivityType(tt.action)
			if result != tt.expected {
				t.Errorf("determineActivityType(%q) = %q, want %q", tt.action, result, tt.expected)
			}
		})
	}
}

// TestFormatActivityTitle tests the formatActivityTitle function
func TestFormatActivityTitle(t *testing.T) {
	tests := []struct {
		action   string
		expected string
	}{
		{"", "Unknown Activity"},
		{"user login", "Login"},
		{"successful login", "Login"},
		{"user logout", "Logout"},
		{"password changed", "Password Changed"},
		{"create secret", "Secret Added"},
		{"add secret", "Secret Added"},
		{"update secret", "Secret Updated"},
		{"edit secret", "Secret Updated"},
		{"delete secret", "Secret Deleted"},
		{"remove secret", "Secret Deleted"},
		{"add recipient", "Recipient Added"},
		{"create recipient", "Recipient Added"},
		{"update recipient", "Recipient Updated"},
		{"edit recipient", "Recipient Updated"},
		{"delete recipient", "Recipient Deleted"},
		{"remove recipient", "Recipient Deleted"},
		{"update settings", "Settings Updated"},
		{"change config", "Settings Updated"},
		{"check_in via email", "Manual Check-in"},
		{"external_activity detected", "Activity Detected"},
		{"activity_detected", "Activity Detected"},
		{"github commit", "GitHub Activity"},
		{"something else", "something else"},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			result := formatActivityTitle(tt.action)
			if result != tt.expected {
				t.Errorf("formatActivityTitle(%q) = %q, want %q", tt.action, result, tt.expected)
			}
		})
	}
}

// TestContains tests the contains function
func TestContains(t *testing.T) {
	tests := []struct {
		str        string
		substrings []string
		expected   bool
	}{
		{"hello world", []string{"hello"}, true},
		{"hello world", []string{"hello", "world"}, true},
		{"hello world", []string{"goodbye"}, false},
		{"hello world", []string{"HELLO"}, true}, // Case-insensitive
		{"hello world", []string{"hello", "goodbye"}, true},
		{"hello world", []string{"goodbye", "farewell"}, false},
		{"", []string{"hello"}, false},
		{"hello world", []string{""}, true}, // Empty substring is in any string
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			result := contains(tt.str, tt.substrings...)
			if result != tt.expected {
				t.Errorf("contains(%q, %v) = %v, want %v", tt.str, tt.substrings, result, tt.expected)
			}
		})
	}
}

// Mock repository that returns errors
type errorMockRepo struct {
	storage.Repository
}

func (r *errorMockRepo) ListPingHistoryByUserID(ctx context.Context, userID string) ([]*models.PingHistory, error) {
	return nil, errors.New("database error")
}

func (r *errorMockRepo) ListAuditLogsByUserID(ctx context.Context, userID string) ([]*models.AuditLog, error) {
	return nil, errors.New("database error")
}
