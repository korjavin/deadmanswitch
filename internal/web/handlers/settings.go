package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/web/middleware"
	"github.com/korjavin/deadmanswitch/internal/web/templates"
)

// SettingsHandler handles settings-related requests
type SettingsHandler struct {
	repo storage.Repository
}

// NewSettingsHandler creates a new SettingsHandler
func NewSettingsHandler(repo storage.Repository) *SettingsHandler {
	return &SettingsHandler{
		repo: repo,
	}
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
		"EmailCheckIn":     true,
		"EmailWarning":     true,
		"TwoFactorEnabled": false,
	}

	data := templates.TemplateData{
		Title:           "Account Settings",
		ActivePage:      "settings",
		IsAuthenticated: true,
		Data: map[string]interface{}{
			"User":     user,
			"Settings": settingsData,
		},
	}

	if err := templates.RenderTemplate(w, "settings.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering settings template: %v", err)
	}
}

// HandleUpdateDeadManSwitchSettings handles the dead man's switch settings update
func (h *SettingsHandler) HandleUpdateDeadManSwitchSettings(w http.ResponseWriter, r *http.Request) {
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

	// Get form values
	pingFrequencyStr := r.FormValue("pingFrequency")
	pingDeadlineStr := r.FormValue("pingDeadline")
	pingMethod := r.FormValue("pingMethod")
	pingingEnabled := r.FormValue("pingingEnabled") == "on"

	// Parse ping frequency
	pingFrequency, err := strconv.Atoi(pingFrequencyStr)
	if err != nil || pingFrequency < 1 || pingFrequency > 30 {
		pingFrequency = 7 // Default to weekly
	}

	// Parse ping deadline
	pingDeadline, err := strconv.Atoi(pingDeadlineStr)
	if err != nil || pingDeadline < 3 || pingDeadline > 30 {
		pingDeadline = 14 // Default to 2 weeks
	}

	// Validate ping method
	if pingMethod != "email" && pingMethod != "telegram" && pingMethod != "both" {
		pingMethod = "email" // Default to email
	}

	// Update user settings
	user.PingFrequency = pingFrequency
	user.PingDeadline = pingDeadline
	user.PingMethod = pingMethod
	user.PingingEnabled = pingingEnabled

	// Save user settings
	if err := h.repo.UpdateUser(r.Context(), user); err != nil {
		log.Printf("Error updating user settings: %v", err)
		http.Error(w, "Failed to update settings", http.StatusInternalServerError)
		return
	}

	// Redirect back to the settings page
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
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
