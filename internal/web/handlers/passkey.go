package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/korjavin/deadmanswitch/internal/auth"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/web/middleware"
	"github.com/korjavin/deadmanswitch/internal/web/templates"
	"github.com/korjavin/deadmanswitch/internal/web/utils"
)

// PasskeyHandler handles passkey-related requests
type PasskeyHandler struct {
	repo            storage.Repository
	webAuthnService *auth.WebAuthnService
}

// NewPasskeyHandler creates a new PasskeyHandler
func NewPasskeyHandler(repo storage.Repository, webAuthnService *auth.WebAuthnService) *PasskeyHandler {
	return &PasskeyHandler{
		repo:            repo,
		webAuthnService: webAuthnService,
	}
}

// HandlePasskeyManagement handles the passkey management page
func (h *PasskeyHandler) HandlePasskeyManagement(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the user's passkeys
	passkeys, err := h.repo.ListPasskeysByUserID(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Error retrieving passkeys", http.StatusInternalServerError)
		log.Printf("Error retrieving passkeys: %v", err)
		return
	}

	// Prepare passkey data for the template
	passkeysData := make([]map[string]interface{}, len(passkeys))
	for i, passkey := range passkeys {
		passkeysData[i] = map[string]interface{}{
			"ID":           passkey.ID,
			"Name":         passkey.Name,
			"CreatedAt":    passkey.CreatedAt.Format("January 2, 2006"),
			"LastUsedAt":   passkey.LastUsedAt.Format("January 2, 2006 at 3:04 PM"),
			"CredentialID": auth.CredentialIDToString(passkey.CredentialID),
		}
	}

	// Prepare template data
	data := templates.TemplateData{
		Title:           "Manage Passkeys",
		ActivePage:      "profile",
		IsAuthenticated: true,
		User: map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		Data: map[string]interface{}{
			"Passkeys": passkeysData,
		},
	}

	if err := templates.RenderTemplate(w, "passkeys.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering passkeys template: %v", err)
	}
}

// HandleBeginRegistration handles the beginning of passkey registration
func (h *PasskeyHandler) HandleBeginRegistration(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Begin registration
	options, err := h.webAuthnService.BeginRegistration(r.Context(), user, w)
	if err != nil {
		http.Error(w, "Error beginning registration", http.StatusInternalServerError)
		log.Printf("Error beginning registration: %v", err)
		return
	}

	// Note: In a real implementation, we would store the session data in a secure session
	// For this demo, we'll rely on the context in the WebAuthnService

	// Convert options to JSON
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		http.Error(w, "Error marshaling options", http.StatusInternalServerError)
		log.Printf("Error marshaling options: %v", err)
		return
	}

	// Return options as JSON
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(optionsJSON); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// HandleFinishRegistration handles the completion of passkey registration
func (h *PasskeyHandler) HandleFinishRegistration(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleFinishRegistration called with method: %s, content-type: %s", r.Method, r.Header.Get("Content-Type"))

	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		log.Printf("User not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	log.Printf("User found in context: %s", user.Email)

	// Variable to store the passkey name
	var passkeyName string

	// Check if this is a JSON request
	if r.Header.Get("Content-Type") == "application/json" {
		// Create a copy of the request body for logging
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}
		// Restore the body for further processing
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		log.Printf("Request body: %s", string(bodyBytes))

		// Parse JSON request
		var requestData struct {
			Credential json.RawMessage `json:"credential"`
			Name       string          `json:"name"`
		}

		decoder := json.NewDecoder(bytes.NewBuffer(bodyBytes))
		if err := decoder.Decode(&requestData); err != nil {
			log.Printf("Error decoding JSON request: %v", err)
			http.Error(w, "Invalid JSON data", http.StatusBadRequest)
			return
		}
		log.Printf("Received JSON request with name: %s and credential data length: %d",
			requestData.Name, len(requestData.Credential))

		// Use the name from the JSON request
		if requestData.Name == "" {
			log.Printf("Passkey name is required in JSON request")
			http.Error(w, "Passkey name is required", http.StatusBadRequest)
			return
		}

		// Continue with the name from JSON
		passkeyName = requestData.Name
	} else {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form data: %v", err)
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Get the passkey name from form
		passkeyName = r.FormValue("name")
		if passkeyName == "" {
			log.Printf("Passkey name is required in form data")
			http.Error(w, "Passkey name is required", http.StatusBadRequest)
			return
		}
		log.Printf("Received form request with name: %s", passkeyName)
	}

	// Finish registration
	log.Printf("Calling FinishRegistration with name: %s", passkeyName)
	_, err := h.webAuthnService.FinishRegistration(r.Context(), user, passkeyName, r)
	if err != nil {
		log.Printf("Error finishing registration: %v", err)
		http.Error(w, "Error finishing registration: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("FinishRegistration completed successfully")

	// Create an audit log entry
	auditLog := &models.AuditLog{
		ID:        utils.GenerateID(),
		UserID:    user.ID,
		Action:    "register_passkey",
		Timestamp: time.Now(),
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Details:   "Registered new passkey: " + passkeyName,
	}

	if err := h.repo.CreateAuditLog(context.Background(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(`{"success": true, "message": "Passkey registered successfully"}`)); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// HandleDeletePasskey handles the deletion of a passkey
func (h *PasskeyHandler) HandleDeletePasskey(w http.ResponseWriter, r *http.Request) {
	// Get the passkey ID from the URL parameter
	passkeyID := utils.GetLastURLSegment(r)
	if passkeyID == "" {
		http.Error(w, "Missing passkey ID", http.StatusBadRequest)
		return
	}

	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the passkey to verify ownership
	passkey, err := h.repo.GetPasskeyByID(r.Context(), passkeyID)
	if err != nil {
		if err == storage.ErrNotFound {
			http.Error(w, "Passkey not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error retrieving passkey", http.StatusInternalServerError)
			log.Printf("Error retrieving passkey: %v", err)
		}
		return
	}

	// Verify ownership
	if passkey.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete the passkey
	if err := h.repo.DeletePasskey(r.Context(), passkeyID); err != nil {
		http.Error(w, "Error deleting passkey", http.StatusInternalServerError)
		log.Printf("Error deleting passkey: %v", err)
		return
	}

	// Create an audit log entry
	auditLog := &models.AuditLog{
		ID:        utils.GenerateID(),
		UserID:    user.ID,
		Action:    "delete_passkey",
		Timestamp: time.Now(),
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Details:   "Deleted passkey: " + passkey.Name,
	}

	if err := h.repo.CreateAuditLog(context.Background(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Redirect back to the passkey management page
	http.Redirect(w, r, "/profile/passkeys", http.StatusSeeOther)
}

// HandleBeginLogin handles the beginning of passkey login
func (h *PasskeyHandler) HandleBeginLogin(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleBeginLogin called with method: %s, content-type: %s", r.Method, r.Header.Get("Content-Type"))

	// Variable to store the email
	var email string

	// Check if this is a JSON request
	if r.Header.Get("Content-Type") == "application/json" {
		// Create a copy of the request body for logging
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}
		// Restore the body for further processing
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		// Request body logging removed for security

		// Parse JSON request
		var requestData struct {
			Email string `json:"email"`
		}

		decoder := json.NewDecoder(bytes.NewBuffer(bodyBytes))
		if err := decoder.Decode(&requestData); err != nil {
			log.Printf("Error decoding JSON request: %v", err)
			http.Error(w, "Invalid JSON data", http.StatusBadRequest)
			return
		}
		log.Printf("Received JSON request for passkey begin login")

		// Use the email from the JSON request
		if requestData.Email == "" {
			log.Printf("Email is required in JSON request")
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}

		// Continue with the email from JSON
		email = requestData.Email
	} else {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form data: %v", err)
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Get the email from form
		email = r.FormValue("email")
		if email == "" {
			log.Printf("Email is required in form data")
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}
		log.Printf("Received form request for passkey begin login")
	}

	// Get the user
	user, err := h.repo.GetUserByEmail(r.Context(), email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Begin login
	options, err := h.webAuthnService.BeginLogin(r.Context(), user, w)
	if err != nil {
		http.Error(w, "Error beginning login", http.StatusInternalServerError)
		log.Printf("Error beginning login: %v", err)
		return
	}

	// Note: In a real implementation, we would store the session data in a secure session
	// For this demo, we'll rely on the context in the WebAuthnService

	// Convert options to JSON
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		http.Error(w, "Error marshaling options", http.StatusInternalServerError)
		log.Printf("Error marshaling options: %v", err)
		return
	}

	// Return options as JSON
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(optionsJSON); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// HandleFinishLogin handles the completion of passkey login
func (h *PasskeyHandler) HandleFinishLogin(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleFinishLogin called with method: %s, content-type: %s", r.Method, r.Header.Get("Content-Type"))

	// Variable to store the email
	var email string

	// Check if this is a JSON request
	if r.Header.Get("Content-Type") == "application/json" {
		// Create a copy of the request body for logging
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading request body: %v", err)
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}
		// Restore the body for further processing
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		// Request body logging removed for security

		// Parse JSON request
		var requestData struct {
			Credential json.RawMessage `json:"credential"`
			Email      string          `json:"email"`
		}

		decoder := json.NewDecoder(bytes.NewBuffer(bodyBytes))
		if err := decoder.Decode(&requestData); err != nil {
			log.Printf("Error decoding JSON request: %v", err)
			http.Error(w, "Invalid JSON data", http.StatusBadRequest)
			return
		}
		log.Printf("Received JSON request for passkey login")

		// Use the email from the JSON request
		if requestData.Email == "" {
			log.Printf("Email is required in JSON request")
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}

		// Continue with the email from JSON
		email = requestData.Email
	} else {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form data: %v", err)
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Get the email from form
		email = r.FormValue("email")
		if email == "" {
			log.Printf("Email is required in form data")
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}
		log.Printf("Received form request for passkey finish login")
	}

	// Get the user
	user, err := h.repo.GetUserByEmail(r.Context(), email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Finish login
	passkey, err := h.webAuthnService.FinishLogin(r.Context(), user, r)
	if err != nil {
		http.Error(w, "Error finishing login", http.StatusInternalServerError)
		log.Printf("Error finishing login: %v", err)
		return
	}

	// Create a new session
	sessionToken := utils.GenerateSecureToken()
	expiresAt := time.Now().Add(24 * time.Hour)

	// Create the session in the database
	session := &models.Session{
		ID:           utils.GenerateID(),
		UserID:       user.ID,
		Token:        sessionToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    expiresAt,
		LastActivity: time.Now(),
		IPAddress:    r.RemoteAddr,
		UserAgent:    r.UserAgent(),
	}

	if err := h.repo.CreateSession(r.Context(), session); err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		log.Printf("Error creating session: %v", err)
		return
	}

	// Update the user's last activity time
	user.LastActivity = time.Now()
	if err := h.repo.UpdateUser(r.Context(), user); err != nil {
		log.Printf("Error updating user last activity: %v", err)
		// Continue anyway, this is not critical
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   r.TLS != nil, // Set Secure flag if using HTTPS
	})

	// Create an audit log entry
	auditLog := &models.AuditLog{
		ID:        utils.GenerateID(),
		UserID:    user.ID,
		Action:    "login_passkey",
		Timestamp: time.Now(),
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Details:   "Logged in with passkey: " + passkey.Name,
	}

	if err := h.repo.CreateAuditLog(r.Context(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Return success
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(`{"success": true, "redirect": "/dashboard"}`)); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
