package handlers

import (
	"fmt"
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
	secrets, err := h.repo.ListSecretsByUserID(r.Context(), user.ID)
	secretCount := 0
	if err == nil {
		secretCount = len(secrets)
	}

	recipients, err := h.repo.ListRecipientsByUserID(r.Context(), user.ID)
	recipientCount := 0
	if err == nil {
		recipientCount = len(recipients)
	}

	// Calculate days active
	daysActive := int(time.Since(user.CreatedAt).Hours() / 24)
	if daysActive < 1 {
		daysActive = 1 // At least 1 day
	}

	// Calculate next check-in time
	nextCheckIn := user.LastActivity.AddDate(0, 0, user.PingFrequency)

	// Calculate deadline time
	deadline := user.LastActivity.AddDate(0, 0, user.PingDeadline)

	// Determine status based on time until deadline
	now := time.Now()
	timeUntilNextCheckIn := nextCheckIn.Sub(now)
	timeUntilDeadline := deadline.Sub(now)

	status := "active"
	statusMessage := "Your dead man's switch is active and all systems are operational."
	triggerTime := ""

	if timeUntilNextCheckIn <= 0 {
		// Check-in is due
		if timeUntilDeadline <= 48*time.Hour && timeUntilDeadline > 0 {
			// Less than 48 hours until deadline
			status = "caution"
			statusMessage = "Your check-in deadline is approaching. Please check in to keep your switch active."
		} else if timeUntilDeadline <= 0 {
			// Past deadline
			status = "danger"
			statusMessage = "Your check-in deadline has passed! Your switch will trigger soon if you don't check in."
			triggerTime = deadline.Format("Jan 2, 2006 15:04 MST")
		}
	}

	// Get recent activity logs
	activityLogs, err := h.repo.ListAuditLogsByUserID(r.Context(), user.ID)
	activities := []map[string]string{{
		"Time":        user.CreatedAt.Format("Jan 2, 2006 15:04"),
		"Description": "Account created",
	}}

	if err == nil && len(activityLogs) > 0 {
		// Add the most recent 5 activity logs
		count := 0
		for i := len(activityLogs) - 1; i >= 0 && count < 5; i-- {
			log := activityLogs[i]
			activities = append(activities, map[string]string{
				"Time":        log.Timestamp.Format("Jan 2, 2006 15:04"),
				"Description": formatActivityDescription(log.Action, log.Details),
			})
			count++
		}
	}

	// Get the latest ping history
	latestPing, err := h.repo.GetLatestPingByUserID(r.Context(), user.ID)
	latestPingInfo := map[string]string{
		"Time":   "",
		"Method": "",
		"Status": "",
	}

	if err == nil && latestPing != nil {
		latestPingInfo["Time"] = latestPing.SentAt.Format("Jan 2, 2006 15:04 MST")
		latestPingInfo["Method"] = formatPingMethod(latestPing.Method)
		latestPingInfo["Status"] = formatPingStatus(latestPing.Status)

		if latestPing.RespondedAt != nil {
			latestPingInfo["RespondedAt"] = latestPing.RespondedAt.Format("Jan 2, 2006 15:04 MST")
		}
	}

	data := templates.TemplateData{
		Title:           "Dashboard",
		ActivePage:      "dashboard",
		IsAuthenticated: true,
		User: map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		Data: map[string]interface{}{
			"Status":        status,
			"StatusMessage": statusMessage,
			"NextCheckIn":   nextCheckIn.Format("Jan 2, 2006 15:04 MST"),
			"Deadline":      deadline.Format("Jan 2, 2006 15:04 MST"),
			"TriggerTime":   triggerTime,
			"TimeRemaining": formatDuration(timeUntilDeadline),
			"LastActivity":  user.LastActivity.Format("Jan 2, 2006 15:04 MST"),
			"PingFrequency": user.PingFrequency,
			"PingDeadline":  user.PingDeadline,
			"PingMethod":    formatPingMethod(user.PingMethod),
			"LatestPing":    latestPingInfo,
			"Stats": map[string]interface{}{
				"TotalSecrets":     secretCount,
				"ActiveRecipients": recipientCount,
				"DaysActive":       daysActive,
			},
			"Activities": activities,
		},
	}

	if err := templates.RenderTemplate(w, "dashboard.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering dashboard template: %v", err)
	}
}

// formatActivityDescription returns a user-friendly description of an activity
func formatActivityDescription(action, details string) string {
	switch action {
	case "login":
		return "Logged in"
	case "logout":
		return "Logged out"
	case "password_changed":
		return "Changed password"
	case "check_in":
		return "Checked in"
	case "reminder_sent":
		return "Reminder sent"
	case "urgent_reminder_sent":
		return "Urgent reminder sent"
	case "final_warning_sent":
		return "Final warning sent"
	case "switch_triggered":
		return "Switch triggered"
	case "switch_trigger_cancelled":
		return "Switch trigger cancelled"
	default:
		if details != "" {
			return details
		}
		return action
	}
}

// formatPingMethod returns a user-friendly description of a ping method
func formatPingMethod(method string) string {
	switch method {
	case "email":
		return "Email"
	case "telegram":
		return "Telegram"
	case "both":
		return "Email & Telegram"
	default:
		return "Email"
	}
}

// formatPingStatus returns a user-friendly description of a ping status
func formatPingStatus(status string) string {
	switch status {
	case "sent":
		return "Sent"
	case "delivered":
		return "Delivered"
	case "responded":
		return "Responded"
	default:
		return status
	}
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < 0 {
		return "Expired"
	}

	d = d.Round(time.Minute)
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d days, %d hours", days, hours)
	} else if hours > 0 {
		return fmt.Sprintf("%d hours, %d minutes", hours, minutes)
	}
	return fmt.Sprintf("%d minutes", minutes)
}
