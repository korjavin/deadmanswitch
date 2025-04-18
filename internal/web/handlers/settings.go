package handlers

import (
	"log"
	"net/http"

	"github.com/korjavin/deadmanswitch/internal/web/middleware"
	"github.com/korjavin/deadmanswitch/internal/web/templates"
)

// SettingsHandler handles settings-related requests
type SettingsHandler struct{}

// NewSettingsHandler creates a new SettingsHandler
func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

// HandleSettings handles the settings page
func (h *SettingsHandler) HandleSettings(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Mock settings data
	settingsData := map[string]interface{}{
		"EmailCheckIn":      true,
		"EmailWarning":      true,
		"CheckInInterval":   30,  // Monthly
		"GracePeriod":       7,   // 1 week
		"TwoFactorEnabled":  false,
	}

	data := templates.TemplateData{
		Title:           "Account Settings",
		ActivePage:      "settings",
		IsAuthenticated: true,
		Data: map[string]interface{}{
			"User":     map[string]interface{}{"Email": user.Email},
			"Settings": settingsData,
		},
	}

	if err := templates.RenderTemplate(w, "settings.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering settings template: %v", err)
	}
}

// HandleUpdateNotificationSettings handles the notification settings update
func (h *SettingsHandler) HandleUpdateNotificationSettings(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// TODO: Implement notification settings update logic

	// Redirect back to the settings page
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

// HandleUpdateSecuritySettings handles the security settings update
func (h *SettingsHandler) HandleUpdateSecuritySettings(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// TODO: Implement security settings update logic

	// Redirect back to the settings page
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}
