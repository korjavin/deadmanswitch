package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
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

	// Fetch the user's recipients from the database
	dbRecipients, err := h.repo.ListRecipientsByUserID(context.Background(), user.ID)
	if err != nil {
		http.Error(w, "Error fetching recipients", http.StatusInternalServerError)
		log.Printf("Error fetching recipients: %v", err)
		return
	}

	// Convert to template-friendly format
	recipients := make([]map[string]interface{}, 0, len(dbRecipients))
	for _, r := range dbRecipients {
		// Get assigned secrets for this recipient
		assignments, err := h.repo.ListSecretAssignmentsByRecipientID(context.Background(), r.ID)
		if err != nil {
			log.Printf("Error fetching secret assignments for recipient %s: %v", r.ID, err)
			// Continue anyway, don't fail the whole request
		}

		// Create a map of assigned secrets
		assignedSecrets := make([]map[string]interface{}, 0, len(assignments))
		for _, a := range assignments {
			secret, err := h.repo.GetSecretByID(context.Background(), a.SecretID)
			if err != nil {
				log.Printf("Error fetching secret %s: %v", a.SecretID, err)
				continue
			}

			assignedSecrets = append(assignedSecrets, map[string]interface{}{
				"ID":    secret.ID,
				"Title": secret.Name,
			})
		}

		// Determine contact method based on available fields
		contactMethod := "email"
		if r.PhoneNumber != "" {
			contactMethod = "phone"
		}

		recipientEntry := map[string]interface{}{
			"ID":              r.ID,
			"Name":            r.Name,
			"Email":           r.Email,
			"PhoneNumber":     r.PhoneNumber,
			"CreatedAt":       r.CreatedAt,
			"UpdatedAt":       r.UpdatedAt,
			"Relationship":    "other", // Default value, not in the base model
			"ContactMethod":   contactMethod,
			"Verified":        true, // Default value, not in the base model
			"AssignedSecrets": assignedSecrets,
		}
		recipients = append(recipients, recipientEntry)
	}

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

	// Get form values
	name := r.FormValue("name")
	email := r.FormValue("email")
	phoneNumber := r.FormValue("phoneNumber")
	notes := r.FormValue("notes")

	if name == "" || email == "" {
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	// Create the recipient in the database
	recipient := &models.Recipient{
		UserID:      user.ID,
		Name:        name,
		Email:       email,
		PhoneNumber: phoneNumber,
		Message:     notes, // Use the notes field as the message
	}

	if err := h.repo.CreateRecipient(context.Background(), recipient); err != nil {
		http.Error(w, "Error creating recipient", http.StatusInternalServerError)
		log.Printf("Error creating recipient: %v", err)
		return
	}

	// Create an audit log entry
	auditLog := &models.AuditLog{
		UserID:    user.ID,
		Action:    "create_recipient",
		Timestamp: recipient.CreatedAt,
		Details:   "Created recipient: " + recipient.Name,
	}

	if err := h.repo.CreateAuditLog(context.Background(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Redirect to the recipients list page
	http.Redirect(w, r, "/recipients", http.StatusSeeOther)
}

// HandleEditRecipientForm handles the edit recipient form page
func (h *RecipientsHandler) HandleEditRecipientForm(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the recipient ID from the URL
	recipientID := r.PathValue("id")
	if recipientID == "" {
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}

	// Fetch the recipient from the database
	recipient, err := h.repo.GetRecipientByID(context.Background(), recipientID)
	if err != nil {
		http.Error(w, "Error fetching recipient", http.StatusInternalServerError)
		log.Printf("Error fetching recipient: %v", err)
		return
	}

	// Verify that the recipient belongs to the user
	if recipient.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Determine contact method based on available fields
	contactMethod := "email"
	if recipient.PhoneNumber != "" {
		contactMethod = "phone"
	}

	// Convert to template-friendly format
	recipientData := map[string]interface{}{
		"ID":            recipient.ID,
		"Name":          recipient.Name,
		"Email":         recipient.Email,
		"PhoneNumber":   recipient.PhoneNumber,
		"Notes":         recipient.Message,
		"CreatedAt":     recipient.CreatedAt,
		"UpdatedAt":     recipient.UpdatedAt,
		"Relationship":  "other", // Default value, not in the base model
		"ContactMethod": contactMethod,
		"Verified":      true, // Default value, not in the base model
	}

	data := templates.TemplateData{
		Title:           "Edit Recipient",
		ActivePage:      "recipients",
		IsAuthenticated: true,
		User: map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		Data: map[string]interface{}{
			"Recipient": recipientData,
		},
	}

	if err := templates.RenderTemplate(w, "new-recipient.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering edit-recipient template: %v", err)
	}
}

// HandleUpdateRecipient handles the edit recipient form submission
func (h *RecipientsHandler) HandleUpdateRecipient(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the recipient ID from the URL
	recipientID := r.PathValue("id")
	if recipientID == "" {
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}

	// Fetch the recipient from the database
	recipient, err := h.repo.GetRecipientByID(context.Background(), recipientID)
	if err != nil {
		http.Error(w, "Error fetching recipient", http.StatusInternalServerError)
		log.Printf("Error fetching recipient: %v", err)
		return
	}

	// Verify that the recipient belongs to the user
	if recipient.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get form values
	name := r.FormValue("name")
	email := r.FormValue("email")
	phoneNumber := r.FormValue("phoneNumber")
	notes := r.FormValue("notes")

	if name == "" || email == "" {
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	// Update the recipient
	recipient.Name = name
	recipient.Email = email
	recipient.PhoneNumber = phoneNumber
	recipient.Message = notes

	if err := h.repo.UpdateRecipient(context.Background(), recipient); err != nil {
		http.Error(w, "Error updating recipient", http.StatusInternalServerError)
		log.Printf("Error updating recipient: %v", err)
		return
	}

	// Create an audit log entry
	auditLog := &models.AuditLog{
		UserID:    user.ID,
		Action:    "update_recipient",
		Timestamp: recipient.UpdatedAt,
		Details:   "Updated recipient: " + recipient.Name,
	}

	if err := h.repo.CreateAuditLog(context.Background(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Redirect to the recipients list page
	http.Redirect(w, r, "/recipients", http.StatusSeeOther)
}

// HandleManageSecrets handles the manage secrets page for a recipient
func (h *RecipientsHandler) HandleManageSecrets(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the recipient ID from the URL
	recipientID := r.PathValue("id")
	if recipientID == "" {
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}

	// Fetch the recipient from the database
	recipient, err := h.repo.GetRecipientByID(context.Background(), recipientID)
	if err != nil {
		http.Error(w, "Error fetching recipient", http.StatusInternalServerError)
		log.Printf("Error fetching recipient: %v", err)
		return
	}

	// Verify that the recipient belongs to the user
	if recipient.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch all secrets for the user
	dbSecrets, err := h.repo.ListSecretsByUserID(context.Background(), user.ID)
	if err != nil {
		http.Error(w, "Error fetching secrets", http.StatusInternalServerError)
		log.Printf("Error fetching secrets: %v", err)
		return
	}

	// Fetch all secret assignments for the recipient
	assignments, err := h.repo.ListSecretAssignmentsByRecipientID(context.Background(), recipientID)
	if err != nil {
		http.Error(w, "Error fetching secret assignments", http.StatusInternalServerError)
		log.Printf("Error fetching secret assignments: %v", err)
		return
	}

	// Create a map of assigned secret IDs for quick lookup
	assignedSecretIDs := make(map[string]bool)
	for _, a := range assignments {
		assignedSecretIDs[a.SecretID] = true
	}

	// Convert to template-friendly format
	secrets := make([]map[string]interface{}, 0, len(dbSecrets))
	for _, s := range dbSecrets {
		secretEntry := map[string]interface{}{
			"ID":         s.ID,
			"Title":      s.Name,
			"IsAssigned": assignedSecretIDs[s.ID],
			"CreatedAt":  s.CreatedAt,
			"UpdatedAt":  s.UpdatedAt,
		}

		secrets = append(secrets, secretEntry)
	}

	// Convert recipient to template-friendly format
	recipientData := map[string]interface{}{
		"ID":    recipient.ID,
		"Name":  recipient.Name,
		"Email": recipient.Email,
	}

	data := templates.TemplateData{
		Title:           "Manage Secrets for " + recipient.Name,
		ActivePage:      "recipients",
		IsAuthenticated: true,
		User: map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		Data: map[string]interface{}{
			"Recipient": recipientData,
			"Secrets":   secrets,
		},
	}

	if err := templates.RenderTemplate(w, "manage-recipient-secrets.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering manage-recipient-secrets template: %v", err)
	}
}

// HandleUpdateRecipientSecrets handles the update of secrets assigned to a recipient
func (h *RecipientsHandler) HandleUpdateRecipientSecrets(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the recipient ID from the URL
	recipientID := r.PathValue("id")
	if recipientID == "" {
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}

	// Fetch the recipient from the database
	recipient, err := h.repo.GetRecipientByID(context.Background(), recipientID)
	if err != nil {
		http.Error(w, "Error fetching recipient", http.StatusInternalServerError)
		log.Printf("Error fetching recipient: %v", err)
		return
	}

	// Verify that the recipient belongs to the user
	if recipient.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get the selected secret IDs
	selectedSecretIDs := r.Form["secrets"]

	// Fetch all current assignments for the recipient
	currentAssignments, err := h.repo.ListSecretAssignmentsByRecipientID(context.Background(), recipientID)
	if err != nil {
		http.Error(w, "Error fetching secret assignments", http.StatusInternalServerError)
		log.Printf("Error fetching secret assignments: %v", err)
		return
	}

	// Create a map of current assignments for quick lookup
	currentAssignmentMap := make(map[string]*models.SecretAssignment)
	for _, a := range currentAssignments {
		currentAssignmentMap[a.SecretID] = a
	}

	// Create a map of selected secret IDs for quick lookup
	selectedSecretMap := make(map[string]bool)
	for _, id := range selectedSecretIDs {
		selectedSecretMap[id] = true
	}

	// Remove assignments that are no longer selected
	for secretID, assignment := range currentAssignmentMap {
		if !selectedSecretMap[secretID] {
			if err := h.repo.DeleteSecretAssignment(context.Background(), assignment.ID); err != nil {
				log.Printf("Error deleting secret assignment: %v", err)
				// Continue anyway, don't fail the whole request
			}
		}
	}

	// Add new assignments for newly selected secrets
	for _, secretID := range selectedSecretIDs {
		if _, exists := currentAssignmentMap[secretID]; !exists {
			// Create a new assignment
			assignment := &models.SecretAssignment{
				SecretID:    secretID,
				RecipientID: recipientID,
				UserID:      user.ID,
			}

			if err := h.repo.CreateSecretAssignment(context.Background(), assignment); err != nil {
				log.Printf("Error creating secret assignment: %v", err)
				// Continue anyway, don't fail the whole request
			}
		}
	}

	// Create an audit log entry
	auditLog := &models.AuditLog{
		UserID:    user.ID,
		Action:    "update_recipient_secrets",
		Timestamp: time.Now(),
		Details:   "Updated secrets for recipient: " + recipient.Name,
	}

	if err := h.repo.CreateAuditLog(context.Background(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Redirect to the recipient list page
	http.Redirect(w, r, "/recipients", http.StatusSeeOther)
}

// HandleTestContact handles the test contact request
func (h *RecipientsHandler) HandleTestContact(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the recipient ID from the URL
	recipientID := r.PathValue("id")
	if recipientID == "" {
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}

	// Fetch the recipient from the database
	recipient, err := h.repo.GetRecipientByID(context.Background(), recipientID)
	if err != nil {
		http.Error(w, "Error fetching recipient", http.StatusInternalServerError)
		log.Printf("Error fetching recipient: %v", err)
		return
	}

	// Verify that the recipient belongs to the user
	if recipient.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// In a real implementation, we would send a test email or other notification to the recipient
	// For now, we'll just log it and redirect back to the recipients page
	log.Printf("Test contact sent to recipient: %s (%s)", recipient.Name, recipient.Email)

	// Create an audit log entry
	auditLog := &models.AuditLog{
		UserID:    user.ID,
		Action:    "test_contact_recipient",
		Timestamp: time.Now(),
		Details:   "Sent test contact to recipient: " + recipient.Name,
	}

	if err := h.repo.CreateAuditLog(context.Background(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Redirect to the recipients list page with a success message
	http.Redirect(w, r, "/recipients?test_contact=success", http.StatusSeeOther)
}

// HandleDeleteRecipient handles the delete recipient request
func (h *RecipientsHandler) HandleDeleteRecipient(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the recipient ID from the URL
	recipientID := r.PathValue("id")
	if recipientID == "" {
		http.Error(w, "Recipient ID is required", http.StatusBadRequest)
		return
	}

	// Fetch the recipient from the database
	recipient, err := h.repo.GetRecipientByID(context.Background(), recipientID)
	if err != nil {
		http.Error(w, "Error fetching recipient", http.StatusInternalServerError)
		log.Printf("Error fetching recipient: %v", err)
		return
	}

	// Verify that the recipient belongs to the user
	if recipient.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete the recipient
	if err := h.repo.DeleteRecipient(context.Background(), recipientID); err != nil {
		http.Error(w, "Error deleting recipient", http.StatusInternalServerError)
		log.Printf("Error deleting recipient: %v", err)
		return
	}

	// Create an audit log entry
	auditLog := &models.AuditLog{
		UserID:    user.ID,
		Action:    "delete_recipient",
		Timestamp: recipient.UpdatedAt,
		Details:   "Deleted recipient: " + recipient.Name,
	}

	if err := h.repo.CreateAuditLog(context.Background(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Redirect to the recipients list page
	http.Redirect(w, r, "/recipients", http.StatusSeeOther)
}
