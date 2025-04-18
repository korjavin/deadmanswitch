package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/korjavin/deadmanswitch/internal/email"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/web/templates"
	"github.com/korjavin/deadmanswitch/internal/web/utils"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	repo        storage.Repository
	emailClient *email.Client
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(repo storage.Repository, emailClient *email.Client) *AuthHandler {
	return &AuthHandler{
		repo:        repo,
		emailClient: emailClient,
	}
}

// HandleLoginForm handles the login form page
func (h *AuthHandler) HandleLoginForm(w http.ResponseWriter, r *http.Request) {
	data := templates.TemplateData{
		Title:      "Login",
		ActivePage: "login",
		Data:       make(map[string]interface{}),
	}

	if err := templates.RenderTemplate(w, "login.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering login template: %v", err)
	}
}

// HandleLogin handles the login form submission
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	rememberMe := r.FormValue("remember") == "on"

	// Validate inputs
	if email == "" || password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Get the user from the database
	ctx := r.Context()
	user, err := h.repo.GetUserByEmail(ctx, email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Verify the password
	if !utils.VerifyPassword([]byte(user.PasswordHash), password) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate a session token
	sessionToken := utils.GenerateSecureToken()

	// Set session expiry
	expiresAt := time.Now().Add(24 * time.Hour) // Default expiry time
	if rememberMe {
		expiresAt = time.Now().Add(30 * 24 * time.Hour) // 30 days if "remember me" is checked
	}

	// Create a new session
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

	// Save the session to the database
	if err := h.repo.CreateSession(ctx, session); err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		log.Printf("Error creating session: %v", err)
		return
	}

	// Update the user's last activity time
	user.LastActivity = time.Now()
	if err := h.repo.UpdateUser(ctx, user); err != nil {
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

	// Create audit log entry
	auditLog := &models.AuditLog{
		ID:        utils.GenerateID(),
		UserID:    user.ID,
		Action:    "login",
		Timestamp: time.Now(),
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Details:   "User login",
	}

	if err := h.repo.CreateAuditLog(ctx, auditLog); err != nil {
		// Non-fatal error, just log it
		log.Printf("Error creating audit log for login: %v", err)
	}

	// Redirect to the dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// HandleRegisterForm handles the registration form page
func (h *AuthHandler) HandleRegisterForm(w http.ResponseWriter, r *http.Request) {
	data := templates.TemplateData{
		Title:      "Register",
		ActivePage: "register",
		Data:       make(map[string]interface{}),
	}

	if err := templates.RenderTemplate(w, "register.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering register template: %v", err)
	}
}

// HandleRegister handles the registration form submission
func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	name := r.FormValue("name")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirmPassword")

	// Validate inputs
	if email == "" || name == "" || password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	if password != confirmPassword {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	// Check if the user already exists
	ctx := r.Context()
	_, err := h.repo.GetUserByEmail(ctx, email)
	if err == nil {
		http.Error(w, "Email already registered", http.StatusBadRequest)
		return
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		log.Printf("Error hashing password: %v", err)
		return
	}

	// Create a new user
	user := &models.User{
		ID:                utils.GenerateID(),
		Email:             email,
		PasswordHash:      hashedPassword,
		CreatedAt:         time.Now(),
		LastActivity:      time.Now(),
		PingFrequency:     1, // Default to 1 day
		PingDeadline:      7, // Default to 7 days
		PingingEnabled:    false,
		NextScheduledPing: time.Time{},
	}

	// Save the user to the database
	if err := h.repo.CreateUser(ctx, user); err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		log.Printf("Error creating user: %v", err)
		return
	}

	// Generate a session token
	sessionToken := utils.GenerateSecureToken()

	// Create a new session
	session := &models.Session{
		ID:           utils.GenerateID(),
		UserID:       user.ID,
		Token:        sessionToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		LastActivity: time.Now(),
		IPAddress:    r.RemoteAddr,
		UserAgent:    r.UserAgent(),
	}

	// Save the session to the database
	if err := h.repo.CreateSession(ctx, session); err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		log.Printf("Error creating session: %v", err)
		return
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   r.TLS != nil, // Set Secure flag if using HTTPS
	})

	// Create audit log entry
	auditLog := &models.AuditLog{
		ID:        utils.GenerateID(),
		UserID:    user.ID,
		Action:    "register",
		Timestamp: time.Now(),
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Details:   "User registration",
	}

	if err := h.repo.CreateAuditLog(ctx, auditLog); err != nil {
		// Non-fatal error, just log it
		log.Printf("Error creating audit log for registration: %v", err)
	}

	// Send welcome email
	if h.emailClient != nil {
		welcomeMessage := "Welcome to Dead Man's Switch! Your account has been created successfully."

		if err := h.emailClient.SendEmailSimple([]string{email}, "Welcome to Dead Man's Switch", welcomeMessage, false); err != nil {
			log.Printf("Error sending welcome email: %v", err)
			// Continue anyway, this is not critical
		} else {
			log.Printf("Welcome email sent to %s", email)
		}
	} else {
		log.Printf("Email client not configured, skipping welcome email")
	}

	// Redirect to the dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// HandleLogout handles the logout request
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Get the session cookie
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Get the session token
		sessionToken := cookie.Value

		// Get the session from the database
		ctx := r.Context()
		session, err := h.repo.GetSessionByToken(ctx, sessionToken)
		if err == nil {
			// Delete the session
			if err := h.repo.DeleteSession(ctx, session.ID); err != nil {
				log.Printf("Error deleting session: %v", err)
				// Continue anyway, this is not critical
			}

			// Create audit log entry
			auditLog := &models.AuditLog{
				ID:        utils.GenerateID(),
				UserID:    session.UserID,
				Action:    "logout",
				Timestamp: time.Now(),
				IPAddress: r.RemoteAddr,
				UserAgent: r.UserAgent(),
				Details:   "User logout",
			}

			if err := h.repo.CreateAuditLog(ctx, auditLog); err != nil {
				// Non-fatal error, just log it
				log.Printf("Error creating audit log for logout: %v", err)
			}
		}
	}

	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   r.TLS != nil, // Set Secure flag if using HTTPS
	})

	// Redirect to the home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
