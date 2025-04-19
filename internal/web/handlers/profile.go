package handlers

import (
	"log"
	"net/http"

	"github.com/korjavin/deadmanswitch/internal/config"
	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/web/middleware"
	"github.com/korjavin/deadmanswitch/internal/web/templates"
)

// ProfileHandler handles profile-related requests
type ProfileHandler struct {
	repo   storage.Repository
	config *config.Config
}

// NewProfileHandler creates a new ProfileHandler
func NewProfileHandler(repo storage.Repository, cfg *config.Config) *ProfileHandler {
	return &ProfileHandler{
		repo:   repo,
		config: cfg,
	}
}

// HandleProfile handles the profile page
func (h *ProfileHandler) HandleProfile(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the full user details from the database
	fullUser, err := h.repo.GetUserByID(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Error fetching user details", http.StatusInternalServerError)
		log.Printf("Error fetching user details: %v", err)
		return
	}

	// Prepare Telegram connection data
	telegramConnected := fullUser.TelegramID != ""
	telegramData := map[string]interface{}{
		"Connected":   telegramConnected,
		"BotUsername": h.config.TelegramBotUsername,
	}

	if telegramConnected {
		telegramData["Username"] = fullUser.TelegramUsername
		telegramData["ID"] = fullUser.TelegramID
	}

	// Create user data for the template
	userData := map[string]interface{}{
		"Email":     fullUser.Email,
		"Name":      fullUser.Email, // Use email as name since we don't have a separate name field
		"CreatedAt": fullUser.CreatedAt.Format("January 2, 2006"),
		"LastLogin": fullUser.LastActivity.Format("January 2, 2006 at 3:04 PM"),
	}

	// Prepare 2FA data
	twoFAData := map[string]interface{}{
		"Enabled": fullUser.TOTPEnabled,
	}

	data := templates.TemplateData{
		Title:           "My Profile",
		ActivePage:      "profile",
		IsAuthenticated: true,
		Data: map[string]interface{}{
			"User":           userData,
			"Telegram":       telegramData,
			"TwoFA":          twoFAData,
			"PingFrequency":  fullUser.PingFrequency,
			"PingDeadline":   fullUser.PingDeadline,
			"PingingEnabled": fullUser.PingingEnabled,
			"PingMethod":     fullUser.PingMethod,
			"NextPingDate":   fullUser.NextScheduledPing.Format("January 2, 2006 at 3:04 PM"),
		},
	}

	if err := templates.RenderTemplate(w, "profile.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering profile template: %v", err)
	}
}

// HandleUpdateProfile handles the profile update form submission
func (h *ProfileHandler) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
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

	// TODO: Implement profile update logic

	// Redirect back to the profile page
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}
