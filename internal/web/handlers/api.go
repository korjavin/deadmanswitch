package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

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
	if err := h.repo.UpdateUser(ctx, user); err != nil {
		log.Printf("Error updating user last activity: %v", err)
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"message":       "Check-in successful",
		"next_check_in": user.LastActivity.AddDate(0, 0, user.PingFrequency).Format(time.RFC3339),
	})
}
