package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/korjavin/deadmanswitch/internal/web/middleware"
	"github.com/korjavin/deadmanswitch/internal/web/templates"
)

// HistoryHandler handles history-related requests
type HistoryHandler struct{}

// NewHistoryHandler creates a new HistoryHandler
func NewHistoryHandler() *HistoryHandler {
	return &HistoryHandler{}
}

// HandleHistory handles the activity history page
func (h *HistoryHandler) HandleHistory(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Mock activity data
	now := time.Now()
	activities := []map[string]interface{}{
		{
			"Type":        "checkin",
			"Title":       "Check-in Completed",
			"Description": "You successfully checked in to your Dead Man's Switch.",
			"Timestamp":   now.Format("Jan 2, 2006 at 3:04 PM"),
			"Details":     nil,
		},
		{
			"Type":        "settings",
			"Title":       "Settings Updated",
			"Description": "You changed your check-in interval from 14 days to 30 days.",
			"Timestamp":   now.Add(-48 * time.Hour).Format("Jan 2, 2006 at 3:04 PM"),
			"Details":     "Previous value: 14 days\nNew value: 30 days",
		},
		{
			"Type":        "secret",
			"Title":       "Secret Added",
			"Description": "You added a new secret: 'Banking Credentials'",
			"Timestamp":   now.Add(-72 * time.Hour).Format("Jan 2, 2006 at 3:04 PM"),
			"Details":     nil,
		},
		{
			"Type":        "recipient",
			"Title":       "Recipient Added",
			"Description": "You added a new recipient: 'John Doe'",
			"Timestamp":   now.Add(-96 * time.Hour).Format("Jan 2, 2006 at 3:04 PM"),
			"Details":     nil,
		},
		{
			"Type":        "security",
			"Title":       "Login from New Device",
			"Description": "You logged in from a new device or location.",
			"Timestamp":   now.Add(-120 * time.Hour).Format("Jan 2, 2006 at 3:04 PM"),
			"Details":     "IP Address: 192.168.1.1\nLocation: New York, USA\nDevice: Chrome on macOS",
		},
	}

	data := templates.TemplateData{
		Title:           "Activity History",
		ActivePage:      "history",
		IsAuthenticated: true,
		Data: map[string]interface{}{
			"User":       map[string]interface{}{"Email": user.Email},
			"Activities": activities,
		},
	}

	if err := templates.RenderTemplate(w, "history.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering history template: %v", err)
	}
}
