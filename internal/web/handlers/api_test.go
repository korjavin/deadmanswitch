package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"

	"github.com/korjavin/deadmanswitch/internal/web/middleware"
)

func TestHandleCheckInSuccess(t *testing.T) {
	// Create mock repository
	repo := storage.NewMockRepository()

	// Create a test user
	user := &models.User{
		ID:            "user123",
		Email:         "test@example.com",
		LastActivity:  time.Now().Add(-24 * time.Hour),
		PingFrequency: 7, // 7 days
	}
	repo.Users = append(repo.Users, user)

	// Create the handler
	handler := NewAPIHandler(repo)

	// Create a test request
	req := httptest.NewRequest("POST", "/api/check-in", nil)

	// Create a context with the authenticated user
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.HandleCheckIn(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the content type is JSON
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Handler returned wrong content type: got %v want %v", contentType, "application/json")
	}

	// Parse the response
	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Could not decode response: %v", err)
	}

	// Check the response
	if success, ok := response["success"].(bool); !ok || !success {
		t.Errorf("Expected successful response, got %v", response)
	}

	// Verify that the user was updated
	updatedUser, err := repo.GetUserByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("Could not get updated user: %v", err)
	}

	if !updatedUser.PingingEnabled {
		t.Error("Expected pinging to be enabled, but it wasn't")
	}

	expectedNextPing := time.Now().AddDate(0, 0, user.PingFrequency)
	nextPingDiff := updatedUser.NextScheduledPing.Sub(expectedNextPing)
	if nextPingDiff < -time.Minute || nextPingDiff > time.Minute {
		t.Errorf("Next ping time not set correctly. Expected around %v, got %v (diff: %v)",
			expectedNextPing, updatedUser.NextScheduledPing, nextPingDiff)
	}

	// Verify that a ping history entry was created
	pingHistories, err := repo.ListPingHistoryByUserID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("Could not list ping history: %v", err)
	}
	if len(pingHistories) != 1 {
		t.Errorf("Expected 1 ping history entry, got %d", len(pingHistories))
	} else {
		ping := pingHistories[0]
		if ping.Method != "web" {
			t.Errorf("Expected ping method 'web', got '%s'", ping.Method)
		}
		if ping.Status != "responded" {
			t.Errorf("Expected ping status 'responded', got '%s'", ping.Status)
		}
		if ping.RespondedAt == nil {
			t.Error("Expected ping responded_at to be set")
		}
	}

	// Verify that an audit log entry was created
	auditLogs, err := repo.ListAuditLogsByUserID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("Could not list audit logs: %v", err)
	}
	if len(auditLogs) != 1 {
		t.Errorf("Expected 1 audit log entry, got %d", len(auditLogs))
	} else {
		log := auditLogs[0]
		if log.Action != "check_in" {
			t.Errorf("Expected log action 'check_in', got '%s'", log.Action)
		}
	}
}

func TestHandleCheckInUnauthorized(t *testing.T) {
	// Create mock repository
	repo := storage.NewMockRepository()

	// Create the handler
	handler := NewAPIHandler(repo)

	// Create a test request with no user in context
	req := httptest.NewRequest("POST", "/api/check-in", nil)
	rr := httptest.NewRecorder()

	// Call the handler
	handler.HandleCheckIn(rr, req)

	// Check the status code should be unauthorized
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestHandleCheckInUpdateUserError(t *testing.T) {
	// Create a mock repository that returns an error on UpdateUser
	repo := &mockErrorRepo{
		errorOn: "UpdateUser",
	}

	// Create a test user
	user := &models.User{
		ID:            "user123",
		Email:         "test@example.com",
		LastActivity:  time.Now().Add(-24 * time.Hour),
		PingFrequency: 7, // 7 days
	}

	// Create the handler
	handler := NewAPIHandler(repo)

	// Create a test request
	req := httptest.NewRequest("POST", "/api/check-in", nil)

	// Create a context with the authenticated user
	ctx := context.WithValue(req.Context(), middleware.UserContextKey, user)
	req = req.WithContext(ctx)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.HandleCheckIn(rr, req)

	// Check the status code should be an error
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

// mockErrorRepo is a mock repository that returns errors for specific operations
type mockErrorRepo struct {
	storage.Repository
	errorOn string
}

func (m *mockErrorRepo) UpdateUser(ctx context.Context, user *models.User) error {
	if m.errorOn == "UpdateUser" {
		return errors.New("mock error")
	}
	return nil
}

func (m *mockErrorRepo) CreatePingHistory(ctx context.Context, ping *models.PingHistory) error {
	if m.errorOn == "CreatePingHistory" {
		return errors.New("mock error")
	}
	return nil
}

func (m *mockErrorRepo) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	if m.errorOn == "CreateAuditLog" {
		return errors.New("mock error")
	}
	return nil
}
