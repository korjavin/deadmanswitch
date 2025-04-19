package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/web/middleware"
	"github.com/korjavin/deadmanswitch/internal/web/templates"
)

// HistoryHandler handles history-related requests
type HistoryHandler struct {
	repo storage.Repository
}

// NewHistoryHandler creates a new HistoryHandler
func NewHistoryHandler(repo storage.Repository) *HistoryHandler {
	return &HistoryHandler{
		repo: repo,
	}
}

// HandleHistory handles the activity history page
func (h *HistoryHandler) HandleHistory(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get ping history from the database
	pingHistory, err := h.repo.ListPingHistoryByUserID(context.Background(), user.ID)
	if err != nil {
		log.Printf("Error fetching ping history: %v", err)
		// Continue anyway, we'll just show an empty list
	}

	// Get audit logs from the database
	auditLogs, err := h.repo.ListAuditLogsByUserID(context.Background(), user.ID)
	if err != nil {
		log.Printf("Error fetching audit logs: %v", err)
		// Continue anyway, we'll just show an empty list
	}

	// Combine ping history and audit logs into a single activity list
	activities := make([]map[string]interface{}, 0)

	// Add ping history entries
	for _, ping := range pingHistory {
		activity := map[string]interface{}{
			"Type":        "checkin",
			"Title":       "Check-in " + ping.Status,
			"Description": "Check-in via " + ping.Method,
			"Timestamp":   ping.SentAt.Format("Jan 2, 2006 at 3:04 PM"),
			"Details":     nil,
		}

		if ping.RespondedAt != nil {
			activity["Details"] = "Responded at: " + ping.RespondedAt.Format("Jan 2, 2006 at 3:04 PM")
		}

		activities = append(activities, activity)
	}

	// Add audit log entries
	for _, log := range auditLogs {
		activity := map[string]interface{}{
			"Type":        determineActivityType(log.Action),
			"Title":       formatActivityTitle(log.Action),
			"Description": log.Action,
			"Timestamp":   log.Timestamp.Format("Jan 2, 2006 at 3:04 PM"),
			"Details":     log.Details,
		}

		activities = append(activities, activity)
	}

	// Sort activities by timestamp (newest first)
	// In a real implementation, we would sort by timestamp
	// For now, we'll just use the order they were added

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

// determineActivityType determines the activity type based on the action
func determineActivityType(action string) string {
	if action == "" {
		return "unknown"
	}

	// Check for common action prefixes
	switch {
	case contains(action, "login", "auth", "password"):
		return "security"
	case contains(action, "secret"):
		return "secret"
	case contains(action, "recipient"):
		return "recipient"
	case contains(action, "setting", "config"):
		return "settings"
	case contains(action, "check_in", "check-in", "checkin", "ping"):
		return "checkin"
	case contains(action, "github", "external_activity", "activity_detected"):
		return "activity"
	default:
		return "other"
	}
}

// Define a mapping of keyword patterns to activity titles
var activityTitleMappings = []struct {
	pattern []string
	title   string
}{
	{[]string{"login"}, "Login"},
	{[]string{"logout"}, "Logout"},
	{[]string{"password"}, "Password Changed"},
	{[]string{"create secret", "add secret"}, "Secret Added"},
	{[]string{"update secret", "edit secret"}, "Secret Updated"},
	{[]string{"delete secret", "remove secret"}, "Secret Deleted"},
	{[]string{"create recipient", "add recipient"}, "Recipient Added"},
	{[]string{"update recipient", "edit recipient"}, "Recipient Updated"},
	{[]string{"delete recipient", "remove recipient"}, "Recipient Deleted"},
	{[]string{"setting", "config"}, "Settings Updated"},
	{[]string{"check_in"}, "Manual Check-in"},
	{[]string{"external_activity", "activity_detected"}, "Activity Detected"},
	{[]string{"github"}, "GitHub Activity"},
}

// formatActivityTitle formats the activity title based on the action
func formatActivityTitle(action string) string {
	if action == "" {
		return "Unknown Activity"
	}

	// Convert action to lowercase for case-insensitive matching
	lowercaseAction := strings.ToLower(action)

	// Check each mapping pattern
	for _, mapping := range activityTitleMappings {
		for _, pattern := range mapping.pattern {
			if strings.Contains(lowercaseAction, pattern) {
				return mapping.title
			}
		}
	}

	// If no match is found, return the original action
	return action
}

// contains checks if any of the substrings are in the string
func contains(s string, substrings ...string) bool {
	for _, substr := range substrings {
		if strings.Contains(strings.ToLower(s), strings.ToLower(substr)) {
			return true
		}
	}
	return false
}
