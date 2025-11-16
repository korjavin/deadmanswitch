package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/web/middleware"
)

// APIHandler handles API requests
type APIHandler struct {
	repo storage.Repository
}

// NewAPIHandler creates a new APIHandler
func NewAPIHandler(repo storage.Repository) *APIHandler {
	return &APIHandler{
		repo: repo,
	}
}

// HandleCheckIn handles the check-in API endpoint
func (h *APIHandler) HandleCheckIn(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	ctx := r.Context()
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Update the user's last activity time
	user.LastActivity = time.Now()

	// Calculate the next scheduled ping based on the user's ping frequency
	user.NextScheduledPing = time.Now().AddDate(0, 0, user.PingFrequency)

	// Enable pinging if it wasn't already enabled
	if !user.PingingEnabled {
		user.PingingEnabled = true
	}

	if err := h.repo.UpdateUser(ctx, user); err != nil {
		log.Printf("Error updating user last activity: %v", err)
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	// Create a ping history entry
	pingHistory := &models.PingHistory{
		ID:          uuid.New().String(),
		UserID:      user.ID,
		SentAt:      time.Now().UTC(),
		Method:      "web",
		Status:      "responded",
		RespondedAt: &user.LastActivity,
	}

	if err := h.repo.CreatePingHistory(ctx, pingHistory); err != nil {
		log.Printf("Error creating ping history during check-in: %v", err)
		// Non-fatal error, continue
	}

	// Create audit log entry
	auditLog := &models.AuditLog{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Action:    "check_in",
		Timestamp: time.Now().UTC(),
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Details:   "Manual user check-in via web interface",
	}

	if err := h.repo.CreateAuditLog(ctx, auditLog); err != nil {
		log.Printf("Error creating audit log for check-in: %v", err)
		// Non-fatal error, continue
	}

	// Format the next check-in time for display
	nextCheckInFormatted := user.NextScheduledPing.Format("Jan 2, 2006 15:04 MST")

	// Calculate and format the deadline
	deadline := user.LastActivity.AddDate(0, 0, user.PingDeadline)
	deadlineFormatted := deadline.Format("Jan 2, 2006 15:04 MST")

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"message":       "Check-in successful",
		"next_check_in": user.NextScheduledPing.Format(time.RFC3339),
		"nextCheckIn":   nextCheckInFormatted,
		"deadline":      deadlineFormatted,
	}); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}
