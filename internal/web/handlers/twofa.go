package handlers

import (
	"context"
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

// TwoFAHandler handles 2FA-related requests
type TwoFAHandler struct {
	repo storage.Repository
}

// NewTwoFAHandler creates a new TwoFAHandler
func NewTwoFAHandler(repo storage.Repository) *TwoFAHandler {
	return &TwoFAHandler{
		repo: repo,
	}
}

// HandleSetup handles the 2FA setup page
func (h *TwoFAHandler) HandleSetup(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if 2FA is already enabled
	if user.TOTPEnabled {
		// Redirect to profile page with a message
		http.Redirect(w, r, "/profile?message=2fa_already_enabled", http.StatusSeeOther)
		return
	}

	// Generate a new TOTP secret if one doesn't exist
	var qrCode string
	if user.TOTPSecret == "" {
		config := auth.DefaultTOTPConfig()
		secret, qr, err := auth.GenerateTOTPSecret(user.Email, config)
		if err != nil {
			http.Error(w, "Error generating 2FA secret", http.StatusInternalServerError)
			log.Printf("Error generating 2FA secret: %v", err)
			return
		}

		// Save the secret to the user
		user.TOTPSecret = secret
		user.TOTPEnabled = false
		user.TOTPVerified = false

		if err := h.repo.UpdateUser(context.Background(), user); err != nil {
			http.Error(w, "Error updating user", http.StatusInternalServerError)
			log.Printf("Error updating user with 2FA secret: %v", err)
			return
		}

		qrCode = qr
	} else {
		// Regenerate QR code from existing secret
		config := auth.DefaultTOTPConfig()
		_, qr, err := auth.GenerateTOTPSecret(user.Email, config)
		if err != nil {
			http.Error(w, "Error generating QR code", http.StatusInternalServerError)
			log.Printf("Error generating QR code: %v", err)
			return
		}
		qrCode = qr
	}

	// Prepare template data
	data := templates.TemplateData{
		Title:           "Set Up Two-Factor Authentication",
		ActivePage:      "profile",
		IsAuthenticated: true,
		User: map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		Data: map[string]interface{}{
			"QRCode":     qrCode,
			"TOTPSecret": user.TOTPSecret,
		},
	}

	if err := templates.RenderTemplate(w, "2fa-setup.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering 2FA setup template: %v", err)
	}
}

// HandleVerify handles the 2FA verification
func (h *TwoFAHandler) HandleVerify(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get the verification code
	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "Verification code is required", http.StatusBadRequest)
		return
	}

	// Validate the code
	config := auth.DefaultTOTPConfig()
	if !auth.ValidateTOTP(user.TOTPSecret, code, config) {
		// Render the setup page again with an error
		data := templates.TemplateData{
			Title:           "Set Up Two-Factor Authentication",
			ActivePage:      "profile",
			IsAuthenticated: true,
			User: map[string]interface{}{
				"Email": user.Email,
				"Name":  user.Email,
			},
			Data: map[string]interface{}{
				"Error":      "Invalid verification code. Please try again.",
				"TOTPSecret": user.TOTPSecret,
			},
		}

		if err := templates.RenderTemplate(w, "2fa-setup.html", data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			log.Printf("Error rendering 2FA setup template: %v", err)
		}
		return
	}

	// Code is valid, enable 2FA for the user
	user.TOTPEnabled = true
	user.TOTPVerified = true

	if err := h.repo.UpdateUser(context.Background(), user); err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		log.Printf("Error enabling 2FA for user: %v", err)
		return
	}

	// Create an audit log entry
	auditLog := &models.AuditLog{
		ID:        utils.GenerateID(),
		UserID:    user.ID,
		Action:    "enable_2fa",
		Timestamp: time.Now(),
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Details:   "Enabled two-factor authentication",
	}
	if err := h.repo.CreateAuditLog(context.Background(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Redirect to profile page with success message
	http.Redirect(w, r, "/profile?message=2fa_enabled", http.StatusSeeOther)
}

// HandleDisable handles disabling 2FA
func (h *TwoFAHandler) HandleDisable(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if 2FA is enabled
	if !user.TOTPEnabled {
		http.Redirect(w, r, "/profile?message=2fa_not_enabled", http.StatusSeeOther)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get the verification code
	code := r.FormValue("code")
	if code == "" {
		http.Error(w, "Verification code is required", http.StatusBadRequest)
		return
	}

	// Validate the code
	config := auth.DefaultTOTPConfig()
	if !auth.ValidateTOTP(user.TOTPSecret, code, config) {
		http.Redirect(w, r, "/profile?message=invalid_2fa_code", http.StatusSeeOther)
		return
	}

	// Code is valid, disable 2FA for the user
	user.TOTPEnabled = false
	user.TOTPVerified = false
	user.TOTPSecret = "" // Clear the secret

	if err := h.repo.UpdateUser(context.Background(), user); err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		log.Printf("Error disabling 2FA for user: %v", err)
		return
	}

	// Create an audit log entry
	auditLog := &models.AuditLog{
		ID:        utils.GenerateID(),
		UserID:    user.ID,
		Action:    "disable_2fa",
		Timestamp: time.Now(),
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Details:   "Disabled two-factor authentication",
	}
	if err := h.repo.CreateAuditLog(context.Background(), auditLog); err != nil {
		log.Printf("Error creating audit log: %v", err)
		// Continue anyway, don't fail the whole request
	}

	// Redirect to profile page with success message
	http.Redirect(w, r, "/profile?message=2fa_disabled", http.StatusSeeOther)
}
