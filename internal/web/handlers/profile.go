package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/korjavin/deadmanswitch/internal/config"
	"github.com/korjavin/deadmanswitch/internal/models"
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

	// Prepare GitHub connection data
	githubConnected := fullUser.GitHubUsername != ""
	log.Printf("GitHub connected: %v, username: '%s'", githubConnected, fullUser.GitHubUsername)
	githubData := map[string]interface{}{
		"Connected": githubConnected,
	}

	if githubConnected {
		githubData["Username"] = fullUser.GitHubUsername
		log.Printf("Added GitHub username to template data: %s", fullUser.GitHubUsername)
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

	// Create template data
	templateData := map[string]interface{}{
		"User":           userData,
		"Telegram":       telegramData,
		"GitHub":         githubData,
		"TwoFA":          twoFAData,
		"PingFrequency":  fullUser.PingFrequency,
		"PingDeadline":   fullUser.PingDeadline,
		"PingingEnabled": fullUser.PingingEnabled,
		"PingMethod":     fullUser.PingMethod,
		"NextPingDate":   fullUser.NextScheduledPing.Format("January 2, 2006 at 3:04 PM"),
	}

	// Log GitHub data for debugging
	log.Printf("GitHub data in template: %+v", githubData)
	log.Printf("Template data: %+v", templateData)

	data := templates.TemplateData{
		Title:           "My Profile",
		ActivePage:      "profile",
		IsAuthenticated: true,
		Data:            templateData,
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

	// Get the full user details from the database
	fullUser, err := h.repo.GetUserByID(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Error fetching user details", http.StatusInternalServerError)
		log.Printf("Error fetching user details: %v", err)
		return
	}

	// Update other fields if needed
	// Name field currently not implemented in User model

	// Ping frequency currently not implemented in this handler
	// If we want to implement it, we would parse and update it here

	// Update GitHub username if provided
	githubUsername := r.FormValue("github_username")
	log.Printf("GitHub username from form: '%s'", githubUsername)
	if githubUsername != "" {
		log.Printf("Updating GitHub username to: %s", githubUsername)
		fullUser.GitHubUsername = githubUsername
		fullUser.UpdatedAt = time.Now()

		// Create an audit log entry
		auditLog := &models.AuditLog{
			ID:        generateID(),
			UserID:    fullUser.ID,
			Action:    "update_github_username",
			Timestamp: time.Now(),
			IPAddress: r.RemoteAddr,
			UserAgent: r.UserAgent(),
			Details:   "Updated GitHub username to: " + githubUsername,
		}

		if err := h.repo.CreateAuditLog(r.Context(), auditLog); err != nil {
			log.Printf("Error creating audit log: %v", err)
			// Continue anyway, don't fail the whole request
		}
	}

	// Save the updated user
	log.Printf("Saving user with GitHub username: %s", fullUser.GitHubUsername)
	if err := h.repo.UpdateUser(r.Context(), fullUser); err != nil {
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		log.Printf("Error updating user: %v", err)
		return
	}

	// Redirect back to the profile page
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// HandleDisconnectGitHub handles disconnecting a GitHub account
func (h *ProfileHandler) HandleDisconnectGitHub(w http.ResponseWriter, r *http.Request) {
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

	// Clear the GitHub username
	oldUsername := fullUser.GitHubUsername
	fullUser.GitHubUsername = ""
	fullUser.UpdatedAt = time.Now()

	// Save the updated user
	if err := h.repo.UpdateUser(r.Context(), fullUser); err != nil {
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		log.Printf("Error updating user: %v", err)
		return
	}

	// Create an audit log entry
	auditLog := &models.AuditLog{
		ID:        generateID(),
		UserID:    fullUser.ID,
		Action:    "disconnect_github",
		Timestamp: time.Now(),
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Details:   "Disconnected GitHub account: " + oldUsername,
	}

	if err := h.repo.CreateAuditLog(r.Context(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Redirect back to the profile page
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// Helper function to generate a unique ID
func generateID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		log.Printf("Error generating UUID: %v", err)
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(uuid)
}
