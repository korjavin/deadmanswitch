package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

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
		// Check-in is due. Three states fan out from here:
		// past deadline > danger; within 48h of deadline > caution-imminent;
		// past nudge but >48h from deadline > caution-quiet.
		if timeUntilDeadline <= 0 {
			status = "danger"
			statusMessage = "Your check-in deadline has passed! Your switch will trigger soon if you don't check in."
			triggerTime = deadline.Format("Jan 2, 2006 15:04 MST")
		} else if timeUntilDeadline <= 48*time.Hour {
			status = "caution"
			statusMessage = "Your check-in deadline is approaching. Please check in to keep your switch active."
		} else {
			status = "caution"
			statusMessage = "We haven't heard from you in a while. A quick check-in keeps your switch quiet."
		}
	}

	// Get recent activity logs. The repo returns rows ORDER BY timestamp DESC,
	// so iterate forward to pick the five newest. The synthetic "Account
	// created" row is appended at the end so it never displaces a fresher entry.
	activityLogs, err := h.repo.ListAuditLogsByUserID(r.Context(), user.ID)
	activities := make([]map[string]string, 0, 6)

	if err == nil && len(activityLogs) > 0 {
		for i := 0; i < len(activityLogs) && i < 5; i++ {
			log := activityLogs[i]
			activities = append(activities, map[string]string{
				"Time":        log.Timestamp.Format("Jan 2, 2006 15:04"),
				"Description": formatActivityDescription(log.Action, log.Details),
			})
		}
	}

	activities = append(activities, map[string]string{
		"Time":        user.CreatedAt.Format("Jan 2, 2006 15:04"),
		"Description": "Account created",
	})

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

	// Heartbeat soft-status variant ("ok" | "warn" | "crit") derived from
	// the existing alarm-style Status field. Drives both copy and pulse class.
	statusVariant := softStatusVariant(status)

	// Build the "circle" — top 4 confirmed-or-awaiting recipients with a
	// rough letter count, used by the dashboard's right-rail.
	circle := make([]map[string]interface{}, 0, 4)
	for i, rcp := range recipients {
		if i >= 4 {
			break
		}
		letterCount := 0
		if assigns, err := h.repo.ListSecretAssignmentsByRecipientID(r.Context(), rcp.ID); err == nil {
			letterCount = len(assigns)
		}
		circle = append(circle, map[string]interface{}{
			"Name":        rcp.Name,
			"Email":       rcp.Email,
			"LetterCount": letterCount,
			"Verified":    rcp.IsConfirmed,
		})
	}

	// Passkey presence — used by the "Passkey tap" channel card.
	hasPasskey := false
	if passkeys, err := h.repo.ListPasskeysByUserID(r.Context(), user.ID); err == nil {
		hasPasskey = len(passkeys) > 0
	}

	firstName := firstNameFromUser(user.Email)

	data := templates.TemplateData{
		Title:           "Dashboard",
		ActivePage:      "dashboard",
		IsAuthenticated: true,
		User: map[string]interface{}{
			"Email":             user.Email,
			"Name":              user.Email,
			"FirstName":         firstName,
			"GitHubUsername":    user.GitHubUsername,
			"GitHubConnected":   user.GitHubUsername != "",
			"TelegramUsername":  user.TelegramUsername,
			"TelegramConnected": user.TelegramID != "",
			"HasPasskey":        hasPasskey,
		},
		Data: map[string]interface{}{
			"Status":        status,
			"StatusVariant": statusVariant,
			"StatusMessage": statusMessage,
			"TodayLabel":    strings.ToUpper(now.Format("Mon, Jan 2")),
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
			"Circle":     circle,
			"Timeline": map[string]interface{}{
				"LastSeen": map[string]string{
					"Value": humanizeAgo(time.Since(user.LastActivity)),
					"Sub":   user.LastActivity.Format("Jan 2, 2006"),
				},
				"NextNudge": map[string]string{
					"Value": humanizeUntil(timeUntilNextCheckIn),
					"Sub":   nextCheckIn.Format("Jan 2, 2006"),
				},
				"Delivery": map[string]string{
					"Value": humanizeUntil(timeUntilDeadline),
					"Sub":   deadline.Format("Jan 2, 2006"),
				},
			},
		},
	}

	data.SetHeartbeat(user)

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

// softStatusVariant maps the alarm-style Status string to the soft
// "ok" / "warn" / "crit" variant used by the Heartbeat dashboard.
func softStatusVariant(status string) string {
	switch status {
	case "active":
		return "ok"
	case "caution":
		return "warn"
	default:
		return "crit"
	}
}

// firstNameFromUser returns a friendly first-name token. We don't have a
// real name field on User, so we fall back to the local-part of the email
// (capitalized). Used by the soft "You're good, {name}." copy.
func firstNameFromUser(email string) string {
	local := email
	if at := strings.IndexByte(email, '@'); at > 0 {
		local = email[:at]
	}
	if dot := strings.IndexByte(local, '.'); dot > 0 {
		local = local[:dot]
	}
	if local == "" {
		return "there"
	}
	r, size := utf8.DecodeRuneInString(local)
	if r == utf8.RuneError {
		return "there"
	}
	return string(unicode.ToUpper(r)) + local[size:]
}

// humanizeUntil formats a positive forward-looking duration for the
// "NEXT NUDGE" / "DELIVERY" cells of the timeline strip. Negative values
// collapse to "now".
func humanizeUntil(d time.Duration) string {
	if d <= 0 {
		return "now"
	}
	days := int(d.Hours() / 24)
	if days >= 1 {
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
	hours := int(d.Hours())
	if hours >= 1 {
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	return "soon"
}

// humanizeAgo formats a positive elapsed duration for the "LAST SEEN" cell.
func humanizeAgo(d time.Duration) string {
	if d <= 0 {
		return "just now"
	}
	days := int(d.Hours() / 24)
	if days >= 1 {
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	hours := int(d.Hours())
	if hours >= 1 {
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	mins := int(d.Minutes())
	if mins >= 1 {
		return fmt.Sprintf("%d min ago", mins)
	}
	return "just now"
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
