package handlers

import (
	"log"
	"net/http"

	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/web/middleware"
	"github.com/korjavin/deadmanswitch/internal/web/templates"
)

// RecipientsHandler handles recipients-related requests
type RecipientsHandler struct {
	repo storage.Repository
}

// NewRecipientsHandler creates a new RecipientsHandler
func NewRecipientsHandler(repo storage.Repository) *RecipientsHandler {
	return &RecipientsHandler{
		repo: repo,
	}
}

// HandleListRecipients handles the recipients list page
func (h *RecipientsHandler) HandleListRecipients(w http.ResponseWriter, r *http.Request) {
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
		Title:           "Recipients",
		ActivePage:      "recipients",
		IsAuthenticated: true,
		User: map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		Data: map[string]interface{}{
			"Recipients": recipients,
		},
	}

	if err := templates.RenderTemplate(w, "recipients.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering recipients template: %v", err)
	}
}

// HandleNewRecipientForm handles the new recipient form page
func (h *RecipientsHandler) HandleNewRecipientForm(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	data := templates.TemplateData{
		Title:           "Add Recipient",
		ActivePage:      "recipients",
		IsAuthenticated: true,
		User: map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		Data: map[string]interface{}{},
	}

	if err := templates.RenderTemplate(w, "new-recipient.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering new-recipient template: %v", err)
	}
}

// HandleCreateRecipient handles the new recipient form submission
func (h *RecipientsHandler) HandleCreateRecipient(w http.ResponseWriter, r *http.Request) {
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

	// TODO: Implement recipient creation logic

	// Redirect to the recipients list page
	http.Redirect(w, r, "/recipients", http.StatusSeeOther)
}
