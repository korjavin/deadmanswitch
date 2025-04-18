package handlers

import (
	"log"
	"net/http"

	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/web/middleware"
	"github.com/korjavin/deadmanswitch/internal/web/templates"
)

// SecretsHandler handles secrets-related requests
type SecretsHandler struct {
	repo storage.Repository
}

// NewSecretsHandler creates a new SecretsHandler
func NewSecretsHandler(repo storage.Repository) *SecretsHandler {
	return &SecretsHandler{
		repo: repo,
	}
}

// HandleListSecrets handles the secrets list page
func (h *SecretsHandler) HandleListSecrets(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// In a real implementation, we would fetch the user's secrets from the database
	// For now, we'll just show an empty list
	secrets := []map[string]interface{}{}

	data := templates.TemplateData{
		Title:           "My Secrets",
		ActivePage:      "secrets",
		IsAuthenticated: true,
		User: map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		Data: map[string]interface{}{
			"Secrets": secrets,
		},
	}

	if err := templates.RenderTemplate(w, "secrets.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering secrets template: %v", err)
	}
}

// HandleNewSecretForm handles the new secret form page
func (h *SecretsHandler) HandleNewSecretForm(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// In a real implementation, we would fetch the user's recipients from the database
	// For now, we'll just show an empty list
	recipients := []map[string]interface{}{}

	data := templates.TemplateData{
		Title:           "Add New Secret",
		ActivePage:      "secrets",
		IsAuthenticated: true,
		User: map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		Data: map[string]interface{}{
			"Recipients": recipients,
		},
	}

	if err := templates.RenderTemplate(w, "new-secret.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering new-secret template: %v", err)
	}
}

// HandleCreateSecret handles the new secret form submission
func (h *SecretsHandler) HandleCreateSecret(w http.ResponseWriter, r *http.Request) {
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

	// TODO: Implement secret creation logic

	// Redirect to the secrets list page
	http.Redirect(w, r, "/secrets", http.StatusSeeOther)
}
