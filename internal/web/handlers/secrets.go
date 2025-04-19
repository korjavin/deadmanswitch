package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/korjavin/deadmanswitch/internal/crypto"
	"github.com/korjavin/deadmanswitch/internal/models"
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

	// Fetch the user's secrets from the database
	dbSecrets, err := h.repo.ListSecretsByUserID(context.Background(), user.ID)
	if err != nil {
		http.Error(w, "Error fetching secrets", http.StatusInternalServerError)
		log.Printf("Error fetching secrets: %v", err)
		return
	}

	// Convert to template-friendly format
	secrets := make([]map[string]interface{}, 0, len(dbSecrets))
	for _, s := range dbSecrets {
		// Get assigned recipients for this secret
		assignments, err := h.repo.ListSecretAssignmentsBySecretID(context.Background(), s.ID)
		if err != nil {
			log.Printf("Error fetching secret assignments for secret %s: %v", s.ID, err)
			// Continue anyway, don't fail the whole request
		}

		// Create a list of recipients
		recipients := make([]map[string]interface{}, 0, len(assignments))
		for _, a := range assignments {
			recipient, err := h.repo.GetRecipientByID(context.Background(), a.RecipientID)
			if err != nil {
				log.Printf("Error fetching recipient %s: %v", a.RecipientID, err)
				continue
			}

			recipients = append(recipients, map[string]interface{}{
				"ID":    recipient.ID,
				"Name":  recipient.Name,
				"Email": recipient.Email,
			})
		}

		// Create the secret entry with basic metadata
		// We don't decrypt the content here for security reasons
		secretEntry := map[string]interface{}{
			"ID":             s.ID,
			"Title":          s.Name,
			"Type":           "encrypted",
			"Description":    "Encrypted secret",
			"Content":        "This content is encrypted", // Dummy content for the template
			"CreatedAt":      s.CreatedAt,
			"UpdatedAt":      s.UpdatedAt,
			"EncryptionType": s.EncryptionType,
			"Recipients":     recipients,
		}

		secrets = append(secrets, secretEntry)
	}

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
		recipientEntry := map[string]interface{}{
			"ID":    r.ID,
			"Name":  r.Name,
			"Email": r.Email,
		}
		recipients = append(recipients, recipientEntry)
	}

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

// HandleManageRecipients handles the manage recipients page for a secret
func (h *SecretsHandler) HandleManageRecipients(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the secret ID from the URL
	secretID := r.PathValue("id")
	if secretID == "" {
		http.Error(w, "Secret ID is required", http.StatusBadRequest)
		return
	}

	// Fetch the secret from the database
	secret, err := h.repo.GetSecretByID(context.Background(), secretID)
	if err != nil {
		http.Error(w, "Error fetching secret", http.StatusInternalServerError)
		log.Printf("Error fetching secret: %v", err)
		return
	}

	// Verify that the secret belongs to the user
	if secret.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch all recipients for the user
	dbRecipients, err := h.repo.ListRecipientsByUserID(context.Background(), user.ID)
	if err != nil {
		http.Error(w, "Error fetching recipients", http.StatusInternalServerError)
		log.Printf("Error fetching recipients: %v", err)
		return
	}

	// Fetch all secret assignments for the secret
	assignments, err := h.repo.ListSecretAssignmentsBySecretID(context.Background(), secretID)
	if err != nil {
		http.Error(w, "Error fetching secret assignments", http.StatusInternalServerError)
		log.Printf("Error fetching secret assignments: %v", err)
		return
	}

	// Create a map of assigned recipient IDs for quick lookup
	assignedRecipientIDs := make(map[string]bool)
	for _, a := range assignments {
		assignedRecipientIDs[a.RecipientID] = true
	}

	// Convert to template-friendly format
	recipients := make([]map[string]interface{}, 0, len(dbRecipients))
	for _, r := range dbRecipients {
		recipientEntry := map[string]interface{}{
			"ID":         r.ID,
			"Name":       r.Name,
			"Email":      r.Email,
			"IsAssigned": assignedRecipientIDs[r.ID],
		}

		recipients = append(recipients, recipientEntry)
	}

	// Convert secret to template-friendly format
	secretData := map[string]interface{}{
		"ID":    secret.ID,
		"Title": secret.Name,
	}

	data := templates.TemplateData{
		Title:           "Manage Recipients for " + secret.Name,
		ActivePage:      "secrets",
		IsAuthenticated: true,
		User: map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		Data: map[string]interface{}{
			"Secret":     secretData,
			"Recipients": recipients,
		},
	}

	if err := templates.RenderTemplate(w, "manage-secret-recipients.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering manage-secret-recipients template: %v", err)
	}
}

// HandleUpdateSecretRecipients handles the update of recipients assigned to a secret
func (h *SecretsHandler) HandleUpdateSecretRecipients(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the secret ID from the URL
	secretID := r.PathValue("id")
	if secretID == "" {
		http.Error(w, "Secret ID is required", http.StatusBadRequest)
		return
	}

	// Fetch the secret from the database
	secret, err := h.repo.GetSecretByID(context.Background(), secretID)
	if err != nil {
		http.Error(w, "Error fetching secret", http.StatusInternalServerError)
		log.Printf("Error fetching secret: %v", err)
		return
	}

	// Verify that the secret belongs to the user
	if secret.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get the selected recipient IDs
	selectedRecipientIDs := r.Form["recipients"]

	// Use the shared helper function to update assignments
	err = updateEntityAssignments(
		context.Background(),
		h.repo,
		user.ID,
		secretID,
		"secret", // We're updating a secret's recipients
		secret.Name,
		selectedRecipientIDs,
	)

	if err != nil {
		http.Error(w, "Error updating secret recipients", http.StatusInternalServerError)
		log.Printf("Error updating secret recipients: %v", err)
		return
	}

	// Redirect to the secrets list page
	http.Redirect(w, r, "/secrets", http.StatusSeeOther)
}

// HandleDeleteSecret handles the deletion of a secret
func (h *SecretsHandler) HandleDeleteSecret(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the secret ID from the URL
	secretID := r.PathValue("id")
	if secretID == "" {
		http.Error(w, "Secret ID is required", http.StatusBadRequest)
		return
	}

	// Fetch the secret from the database
	secret, err := h.repo.GetSecretByID(context.Background(), secretID)
	if err != nil {
		http.Error(w, "Error fetching secret", http.StatusInternalServerError)
		log.Printf("Error fetching secret: %v", err)
		return
	}

	// Verify that the secret belongs to the user
	if secret.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete all secret assignments first
	assignments, err := h.repo.ListSecretAssignmentsBySecretID(context.Background(), secretID)
	if err != nil {
		log.Printf("Error fetching secret assignments: %v", err)
		// Continue anyway, don't fail the whole request
	}

	for _, assignment := range assignments {
		if err := h.repo.DeleteSecretAssignment(context.Background(), assignment.ID); err != nil {
			log.Printf("Error deleting secret assignment: %v", err)
			// Continue anyway, don't fail the whole request
		}
	}

	// Delete the secret
	if err := h.repo.DeleteSecret(context.Background(), secretID); err != nil {
		http.Error(w, "Error deleting secret", http.StatusInternalServerError)
		log.Printf("Error deleting secret: %v", err)
		return
	}

	// Create an audit log entry
	auditLog := &models.AuditLog{
		UserID:    user.ID,
		Action:    "delete_secret",
		Timestamp: time.Now(),
		Details:   "Deleted secret: " + secret.Name,
	}

	if err := h.repo.CreateAuditLog(context.Background(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Redirect to the secrets list page
	http.Redirect(w, r, "/secrets", http.StatusSeeOther)
}

// HandleViewSecretForm handles the view/edit secret form page
func (h *SecretsHandler) HandleViewSecretForm(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the secret ID from the URL
	secretID := r.PathValue("id")
	if secretID == "" {
		http.Error(w, "Secret ID is required", http.StatusBadRequest)
		return
	}

	// Fetch the secret from the database
	secret, err := h.repo.GetSecretByID(context.Background(), secretID)
	if err != nil {
		http.Error(w, "Error fetching secret", http.StatusInternalServerError)
		log.Printf("Error fetching secret: %v", err)
		return
	}

	// Verify that the secret belongs to the user
	if secret.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Fetch all recipients for the user
	dbRecipients, err := h.repo.ListRecipientsByUserID(context.Background(), user.ID)
	if err != nil {
		http.Error(w, "Error fetching recipients", http.StatusInternalServerError)
		log.Printf("Error fetching recipients: %v", err)
		return
	}

	// Fetch all secret assignments for the secret
	assignments, err := h.repo.ListSecretAssignmentsBySecretID(context.Background(), secretID)
	if err != nil {
		http.Error(w, "Error fetching secret assignments", http.StatusInternalServerError)
		log.Printf("Error fetching secret assignments: %v", err)
		return
	}

	// Create a map of assigned recipient IDs for quick lookup
	assignedRecipientIDs := make(map[string]bool)
	for _, a := range assignments {
		assignedRecipientIDs[a.RecipientID] = true
	}

	// Convert to template-friendly format
	recipients := make([]map[string]interface{}, 0, len(dbRecipients))
	for _, r := range dbRecipients {
		recipientEntry := map[string]interface{}{
			"ID":         r.ID,
			"Name":       r.Name,
			"Email":      r.Email,
			"IsAssigned": assignedRecipientIDs[r.ID],
		}

		recipients = append(recipients, recipientEntry)
	}

	// Decrypt the secret content for editing
	// In a real implementation, we would get the master key from the user's session
	// For now, we'll use a dummy master key for demonstration
	masterKey := []byte("this-is-a-dummy-master-key-for-demo-only")

	// Log the secret details for debugging
	log.Printf("Secret details - ID: %s, Name: %s, EncryptionType: %s, EncryptedData length: %d",
		secret.ID, secret.Name, secret.EncryptionType, len(secret.EncryptedData))

	// Decrypt the secret content
	decryptedContent := ""
	log.Printf("Attempting to decrypt secret %s with encrypted data length: %d", secret.ID, len(secret.EncryptedData))
	if secret.EncryptedData != "" {
		decryptedBytes, err := crypto.DecryptSecret(secret.EncryptedData, masterKey)
		if err != nil {
			log.Printf("Error decrypting secret %s: %v", secret.ID, err)
			// If decryption fails, we'll show a placeholder
			decryptedContent = "[Unable to decrypt content. The encryption key may have changed.]"
		} else {
			decryptedContent = string(decryptedBytes)
			log.Printf("Successfully decrypted secret %s, content length: %d", secret.ID, len(decryptedContent))
		}
	} else {
		log.Printf("Secret %s has no encrypted data", secret.ID)
	}

	secretData := map[string]interface{}{
		"ID":             secret.ID,
		"Name":           secret.Name,
		"Type":           "note", // Default type
		"Content":        decryptedContent,
		"CreatedAt":      secret.CreatedAt,
		"LastModified":   secret.UpdatedAt,
		"EncryptionType": secret.EncryptionType,
	}

	data := templates.TemplateData{
		Title:           "Edit Secret",
		ActivePage:      "secrets",
		IsAuthenticated: true,
		User: map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		Data: map[string]interface{}{
			"Secret":     secretData,
			"Recipients": recipients,
		},
	}

	if err := templates.RenderTemplate(w, "view-secret.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering view-secret template: %v", err)
	}
}

// Helper function to validate user ownership of a secret
func (h *SecretsHandler) validateSecretOwnership(ctx context.Context, secretID string, userID string) (*models.Secret, error) {
	// Fetch the secret from the database
	secret, err := h.repo.GetSecretByID(ctx, secretID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch secret: %w", err)
	}

	// Verify that the secret belongs to the user
	if secret.UserID != userID {
		return nil, fmt.Errorf("user does not own this secret")
	}

	return secret, nil
}

// Helper function to update secret content with encryption
func updateSecretContent(secret *models.Secret, title string, content string, masterKey []byte) error {
	// Update the title
	secret.Name = title

	// Only re-encrypt if content was provided
	if content != "" {
		// Encrypt the secret content
		encryptedData, err := crypto.EncryptSecret([]byte(content), masterKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt secret: %w", err)
		}
		secret.EncryptedData = encryptedData
	}

	// Update timestamp
	secret.UpdatedAt = time.Now().UTC()

	return nil
}

// Helper function to update secret recipient assignments
func (h *SecretsHandler) updateSecretRecipientAssignments(ctx context.Context,
	secretID string, userID string, recipientIDs []string) error {
	// Fetch all current assignments for the secret
	currentAssignments, err := h.repo.ListSecretAssignmentsBySecretID(ctx, secretID)
	if err != nil {
		return fmt.Errorf("failed to fetch secret assignments: %w", err)
	}

	// Create maps for quick lookups
	currentAssignmentMap := make(map[string]*models.SecretAssignment)
	for _, a := range currentAssignments {
		currentAssignmentMap[a.RecipientID] = a
	}

	selectedRecipientMap := make(map[string]bool)
	for _, id := range recipientIDs {
		selectedRecipientMap[id] = true
	}

	// Remove assignments that are no longer selected
	for recipientID, assignment := range currentAssignmentMap {
		if !selectedRecipientMap[recipientID] {
			if err := h.repo.DeleteSecretAssignment(ctx, assignment.ID); err != nil {
				log.Printf("Error deleting secret assignment: %v", err)
				// Continue anyway, don't fail the whole request
			}
		}
	}

	// Add new assignments for newly selected recipients
	for _, recipientID := range recipientIDs {
		if _, exists := currentAssignmentMap[recipientID]; !exists {
			// Create a new assignment
			assignment := &models.SecretAssignment{
				SecretID:    secretID,
				RecipientID: recipientID,
				UserID:      userID,
			}

			if err := h.repo.CreateSecretAssignment(ctx, assignment); err != nil {
				log.Printf("Error creating secret assignment: %v", err)
				// Continue anyway, don't fail the whole request
			}
		}
	}

	return nil
}

// Helper function to create an audit log entry
func (h *SecretsHandler) createAuditLogEntry(ctx context.Context,
	userID string, action string, details string) {
	auditLog := &models.AuditLog{
		UserID:    userID,
		Action:    action,
		Timestamp: time.Now(),
		Details:   details,
	}

	if err := h.repo.CreateAuditLog(ctx, auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}
}

// HandleUpdateSecret handles the update of a secret
func (h *SecretsHandler) HandleUpdateSecret(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the secret ID from the URL
	secretID := r.PathValue("id")
	if secretID == "" {
		http.Error(w, "Secret ID is required", http.StatusBadRequest)
		return
	}

	// Validate secret ownership
	secret, err := h.validateSecretOwnership(ctx, secretID, user.ID)
	if err != nil {
		if err.Error() == "user does not own this secret" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		} else {
			http.Error(w, "Error fetching secret", http.StatusInternalServerError)
			log.Printf("Error fetching secret: %v", err)
		}
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get form values
	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// In a real implementation, we would get the master key from the user's session
	// For now, we'll use a dummy master key for demonstration
	masterKey := []byte("this-is-a-dummy-master-key-for-demo-only")

	// Update the secret content with encryption if needed
	if err := updateSecretContent(secret, title, content, masterKey); err != nil {
		http.Error(w, "Error updating secret content", http.StatusInternalServerError)
		log.Printf("Error updating secret content: %v", err)
		return
	}

	// Save the updated secret to the database
	if err := h.repo.UpdateSecret(ctx, secret); err != nil {
		http.Error(w, "Error updating secret", http.StatusInternalServerError)
		log.Printf("Error updating secret: %v", err)
		return
	}

	// Process recipient assignments
	recipientIDs := r.Form["recipients"]
	if err := h.updateSecretRecipientAssignments(ctx, secretID, user.ID, recipientIDs); err != nil {
		log.Printf("Error updating recipient assignments: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Create an audit log entry
	h.createAuditLogEntry(ctx, user.ID, "update_secret", "Updated secret: "+secret.Name)

	// Redirect to the secrets list page
	http.Redirect(w, r, "/secrets", http.StatusSeeOther)
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

	// Get form values
	title := r.FormValue("title")
	content := r.FormValue("content")

	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// In a real implementation, we would get the master key from the user's session
	// For now, we'll use a dummy master key for demonstration
	masterKey := []byte("this-is-a-dummy-master-key-for-demo-only")

	// Encrypt the secret content
	encryptedData, err := crypto.EncryptSecret([]byte(content), masterKey)
	if err != nil {
		http.Error(w, "Error encrypting secret", http.StatusInternalServerError)
		log.Printf("Error encrypting secret: %v", err)
		return
	}

	log.Printf("Successfully encrypted content of length %d, resulting in encrypted data of length %d",
		len(content), len(encryptedData))

	// Create the secret in the database
	secret := &models.Secret{
		UserID:         user.ID,
		Name:           title,
		EncryptedData:  encryptedData,
		EncryptionType: "aes-256-gcm",
	}

	if err := h.repo.CreateSecret(context.Background(), secret); err != nil {
		http.Error(w, "Error creating secret", http.StatusInternalServerError)
		log.Printf("Error creating secret: %v", err)
		return
	}

	// Process recipient assignments if any were selected
	recipientIDs := r.Form["recipients"]
	log.Printf("Selected recipient IDs: %v", recipientIDs)

	if len(recipientIDs) == 0 {
		log.Printf("No recipients selected for secret %s", secret.ID)
	}

	for _, recipientID := range recipientIDs {
		assignment := &models.SecretAssignment{
			SecretID:    secret.ID,
			RecipientID: recipientID,
			UserID:      user.ID,
		}

		log.Printf("Creating secret assignment: Secret ID %s, Recipient ID %s", secret.ID, recipientID)

		if err := h.repo.CreateSecretAssignment(context.Background(), assignment); err != nil {
			log.Printf("Error creating secret assignment: %v", err)
			// Continue anyway, don't fail the whole request
		} else {
			log.Printf("Successfully created secret assignment: Secret ID %s, Recipient ID %s", secret.ID, recipientID)
		}
	}

	// Create an audit log entry
	auditLog := &models.AuditLog{
		UserID:    user.ID,
		Action:    "create_secret",
		Timestamp: time.Now(),
		Details:   "Created secret: " + secret.Name,
	}

	if err := h.repo.CreateAuditLog(context.Background(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Redirect to the secrets list page
	http.Redirect(w, r, "/secrets", http.StatusSeeOther)
}
