package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/korjavin/deadmanswitch/internal/web/middleware"
	"github.com/korjavin/deadmanswitch/internal/web/templates"
)

// ProfileHandler handles profile-related requests
type ProfileHandler struct{}

// NewProfileHandler creates a new ProfileHandler
func NewProfileHandler() *ProfileHandler {
	return &ProfileHandler{}
}

// HandleProfile handles the profile page
func (h *ProfileHandler) HandleProfile(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Create user data for the template
	userData := map[string]interface{}{
		"Email":     user.Email,
		"Name":      user.Email, // Use email as name since we don't have a separate name field
		"CreatedAt": user.CreatedAt.Format("January 2, 2006"),
		"LastLogin": time.Now().Add(-24 * time.Hour).Format("January 2, 2006 at 3:04 PM"), // Mock data
	}

	data := templates.TemplateData{
		Title:           "My Profile",
		ActivePage:      "profile",
		IsAuthenticated: true,
		Data: map[string]interface{}{
			"User": userData,
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
