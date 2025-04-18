package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/web/middleware"
	"github.com/korjavin/deadmanswitch/internal/web/templates"
)

// DashboardHandler handles dashboard-related requests
type DashboardHandler struct {
	repo storage.Repository
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(repo storage.Repository) *DashboardHandler {
	return &DashboardHandler{
		repo: repo,
	}
}

// HandleDashboard handles the dashboard page
func (h *DashboardHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get actual counts from database
	secretCount := 0
	recipientCount := 0

	// Calculate days active
	daysActive := int(time.Since(user.CreatedAt).Hours() / 24)
	if daysActive < 1 {
		daysActive = 1 // At least 1 day
	}

	// Calculate next check-in time
	nextCheckIn := user.LastActivity.AddDate(0, 0, user.PingFrequency)

	data := templates.TemplateData{
		Title:           "Dashboard",
		ActivePage:      "dashboard",
		IsAuthenticated: true,
		User: map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		Data: map[string]interface{}{
			"Status":      "active",
			"NextCheckIn": nextCheckIn.Format("Jan 2, 2006 15:04 MST"),
			"Stats": map[string]interface{}{
				"TotalSecrets":     secretCount,
				"ActiveRecipients": recipientCount,
				"DaysActive":       daysActive,
			},
			"Activities": []map[string]string{
				{
					"Time":        user.CreatedAt.Format("Jan 2, 2006 15:04"),
					"Description": "Account created",
				},
			},
		},
	}

	if err := templates.RenderTemplate(w, "dashboard.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering dashboard template: %v", err)
	}
}
