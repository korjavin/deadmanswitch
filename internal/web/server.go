package web

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/korjavin/deadmanswitch/internal/auth"
	"github.com/korjavin/deadmanswitch/internal/config"
	"github.com/korjavin/deadmanswitch/internal/email"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/scheduler"
	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/telegram"
	"github.com/korjavin/deadmanswitch/internal/web/handlers"
	authMiddleware "github.com/korjavin/deadmanswitch/internal/web/middleware"
	"github.com/korjavin/deadmanswitch/internal/web/utils"
)

// Server represents the web server
type Server struct {
	config      *config.Config
	repo        storage.Repository
	emailClient *email.Client
	telegramBot *telegram.Bot
	scheduler   *scheduler.Scheduler
	router      *chi.Mux
	httpServer  *http.Server
	handlers    struct {
		index      *handlers.IndexHandler
		auth       *handlers.AuthHandler
		dashboard  *handlers.DashboardHandler
		secrets    *handlers.SecretsHandler
		recipients *handlers.RecipientsHandler
		api        *handlers.APIHandler
		profile    *handlers.ProfileHandler
		settings   *handlers.SettingsHandler
		history    *handlers.HistoryHandler
		twofa      *handlers.TwoFAHandler
		passkey    *handlers.PasskeyHandler
	}
}

// NewServer creates a new web server
func NewServer(
	cfg *config.Config,
	repo storage.Repository,
	emailClient *email.Client,
	telegramBot *telegram.Bot,
	scheduler *scheduler.Scheduler,
) *Server {
	server := &Server{
		config:      cfg,
		repo:        repo,
		emailClient: emailClient,
		telegramBot: telegramBot,
		scheduler:   scheduler,
		router:      chi.NewRouter(),
	}

	// Initialize WebAuthn service
	// Determine if we're in a development environment (localhost)
	isLocalhost := strings.Contains(cfg.BaseDomain, "localhost")

	// Set the origin based on environment
	var origin string
	if isLocalhost {
		// For localhost development, use HTTP and include the port
		// Note: We need to use the same origin that the browser sends
		// The browser is accessing the app at http://localhost:8082
		origin = "http://localhost:8082"
		log.Printf("Using development WebAuthn origin: %s", origin)
	} else {
		// For production, use HTTPS without port
		origin = fmt.Sprintf("https://%s", cfg.BaseDomain)
		log.Printf("Using production WebAuthn origin: %s", origin)
	}

	webAuthnConfig := auth.WebAuthnConfig{
		RPDisplayName: "Dead Man's Switch",
		RPID:          cfg.BaseDomain,
		RPOrigin:      origin,
	}
	webAuthnService, err := auth.NewWebAuthnService(webAuthnConfig, repo)
	if err != nil {
		log.Printf("Warning: Failed to create WebAuthn service: %v", err)
		// Continue without WebAuthn support
		webAuthnService = nil
	}

	// Initialize handlers
	server.handlers.index = handlers.NewIndexHandler()
	server.handlers.auth = handlers.NewAuthHandler(repo, emailClient)
	server.handlers.dashboard = handlers.NewDashboardHandler(repo)
	server.handlers.secrets = handlers.NewSecretsHandler(repo)
	server.handlers.recipients = handlers.NewRecipientsHandler(repo, emailClient)
	server.handlers.api = handlers.NewAPIHandler(repo)
	server.handlers.profile = handlers.NewProfileHandler(repo, cfg)
	server.handlers.settings = handlers.NewSettingsHandler()
	server.handlers.history = handlers.NewHistoryHandler(repo)
	server.handlers.twofa = handlers.NewTwoFAHandler(repo)
	server.handlers.passkey = handlers.NewPasskeyHandler(repo, webAuthnService)

	// Set up routes
	server.setupRoutes()

	return server
}

// Start starts the web server
func (s *Server) Start() error {
	// Configure the HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	s.httpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start the server
	log.Printf("Starting web server on %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start web server: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the web server
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// setupRoutes configures all the routes for the server
func (s *Server) setupRoutes() {
	r := s.router

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Public routes
	r.Group(func(r chi.Router) {
		// Landing page
		r.Get("/", s.handlers.index.HandleIndex)

		// Authentication
		r.Get("/login", s.handlers.auth.HandleLoginForm)
		r.Post("/login", s.handlers.auth.HandleLogin)
		r.Get("/register", s.handlers.auth.HandleRegisterForm)
		r.Post("/register", s.handlers.auth.HandleRegister)

		// Passkey authentication
		r.Post("/login/passkey/begin", s.handlers.passkey.HandleBeginLogin)
		r.Post("/login/passkey/finish", s.handlers.passkey.HandleFinishLogin)

		// Recipient confirmation
		r.Get("/confirm/{code}", s.handlers.recipients.HandleConfirmRecipient)

		// Static files - try multiple paths
		// First try absolute path in container, then relative path
		staticDirs := []string{"/app/web/static", "./web/static"}
		var fileServer http.Handler

		// Try each path until we find one that exists
		for _, dir := range staticDirs {
			if _, err := os.Stat(dir); err == nil {
				if s.config.Debug {
					log.Printf("Using static files from: %s", dir)
				}
				fileServer = http.FileServer(http.Dir(dir))
				break
			}
		}

		// If no valid path was found, use the first one as a fallback
		if fileServer == nil {
			log.Printf("Warning: Could not find static files directory, using fallback")
			fileServer = http.FileServer(http.Dir(staticDirs[0]))
		}

		r.Handle("/static/*", http.StripPrefix("/static", fileServer))

		// Logout (accessible without authentication)
		r.Get("/logout", s.handlers.auth.HandleLogout)
	})

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Auth(s.repo))

		// Dashboard
		r.Get("/dashboard", s.handlers.dashboard.HandleDashboard)

		// Secrets management
		r.Get("/secrets", s.handlers.secrets.HandleListSecrets)
		r.Get("/secrets/new", s.handlers.secrets.HandleNewSecretForm)
		r.Post("/secrets/new", s.handlers.secrets.HandleCreateSecret)
		r.Get("/secrets/{id}", s.handlers.secrets.HandleViewSecretForm)
		r.Get("/secrets/{id}/assign", s.handlers.secrets.HandleManageRecipients)
		r.Post("/secrets/{id}/assign", s.handlers.secrets.HandleUpdateSecretRecipients)
		r.Delete("/secrets/{id}", s.handlers.secrets.HandleDeleteSecret)
		// Handle POST requests with _method=DELETE or regular updates
		r.Post("/secrets/{id}", func(w http.ResponseWriter, r *http.Request) {
			if r.FormValue("_method") == "DELETE" {
				s.handlers.secrets.HandleDeleteSecret(w, r)
				return
			}
			// If not a DELETE request, handle as update
			s.handlers.secrets.HandleUpdateSecret(w, r)
		})

		// Recipients management
		r.Get("/recipients", s.handlers.recipients.HandleListRecipients)
		r.Get("/recipients/new", s.handlers.recipients.HandleNewRecipientForm)
		r.Post("/recipients/new", s.handlers.recipients.HandleCreateRecipient)
		r.Get("/recipients/{id}", s.handlers.recipients.HandleEditRecipientForm)
		r.Post("/recipients/{id}", func(w http.ResponseWriter, r *http.Request) {
			if r.FormValue("_method") == "DELETE" {
				s.handlers.recipients.HandleDeleteRecipient(w, r)
				return
			}
			// If not a DELETE request, handle as update
			s.handlers.recipients.HandleUpdateRecipient(w, r)
		})
		r.Delete("/recipients/{id}", s.handlers.recipients.HandleDeleteRecipient)
		r.Get("/recipients/{id}/secrets", s.handlers.recipients.HandleManageSecrets)
		r.Post("/recipients/{id}/secrets", s.handlers.recipients.HandleUpdateRecipientSecrets)
		r.Get("/recipients/{id}/test", s.handlers.recipients.HandleTestContact)

		// Profile and settings
		r.Get("/profile", s.handlers.profile.HandleProfile)
		r.Post("/profile", s.handlers.profile.HandleUpdateProfile)
		r.Post("/profile/github/disconnect", s.handlers.profile.HandleDisconnectGitHub)
		r.Get("/settings", s.handlers.settings.HandleSettings)
		r.Post("/settings/notifications", s.handlers.settings.HandleUpdateNotificationSettings)
		r.Post("/settings/security", s.handlers.settings.HandleUpdateSecuritySettings)

		// Two-factor authentication
		r.Get("/2fa/setup", s.handlers.twofa.HandleSetup)
		r.Post("/2fa/verify", s.handlers.twofa.HandleVerify)
		r.Post("/2fa/disable", s.handlers.twofa.HandleDisable)

		// Passkey management
		r.Get("/profile/passkeys", s.handlers.passkey.HandlePasskeyManagement)
		r.Post("/profile/passkeys/register/begin", s.handlers.passkey.HandleBeginRegistration)
		r.Post("/profile/passkeys/register/finish", s.handlers.passkey.HandleFinishRegistration)
		r.Post("/profile/passkeys/{id}", s.handlers.passkey.HandleDeletePasskey)

		// History
		r.Get("/history", s.handlers.history.HandleHistory)

		// Check-in
		r.Post("/api/check-in", s.handlers.api.HandleCheckIn)
	})
}

// Basic route handlers

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// For now, just render a simple index page
	// Check if we're in debug mode and print the current working directory
	if s.config.Debug {
		cwd, err := os.Getwd()
		if err == nil {
			log.Printf("Current working directory: %s", cwd)
		}
	}

	// Try to find templates in multiple locations
	templatePaths := [][]string{
		{"/app/web/templates/layout.html", "/app/web/templates/index.html"},
		{"./web/templates/layout.html", "./web/templates/index.html"},
	}

	var tmpl *template.Template
	var err error
	var templateErr error

	for _, paths := range templatePaths {
		tmpl, err = template.ParseFiles(paths...)
		if err == nil {
			break
		}
		if templateErr == nil {
			templateErr = err
		}
	}

	if tmpl == nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing index template: %v", templateErr)
		return
	}
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing index template: %v", err)
		return
	}

	data := map[string]interface{}{
		"Title":           "Dead Man's Switch",
		"ActivePage":      "home",
		"IsAuthenticated": false,
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		log.Printf("Error executing index template: %v", err)
		return
	}
}

func (s *Server) handleLoginForm(w http.ResponseWriter, r *http.Request) {
	// If user is already logged in, redirect to dashboard
	if cookie, err := r.Cookie("session_token"); err == nil {
		ctx := r.Context()
		if session, err := s.repo.GetSessionByToken(ctx, cookie.Value); err == nil {
			if !session.ExpiresAt.Before(time.Now()) {
				http.Redirect(w, r, "/dashboard", http.StatusFound)
				return
			}
		}
	}

	// Prepare data for the template
	data := map[string]interface{}{
		"Title":           "Login",
		"ActivePage":      "login",
		"IsAuthenticated": false,
	}

	// Try to find templates in multiple locations
	templatePaths := [][]string{
		{"/app/web/templates/layout.html", "/app/web/templates/login.html"},
		{"./web/templates/layout.html", "./web/templates/login.html"},
	}

	var tmpl *template.Template
	var err error
	var templateErr error

	for _, paths := range templatePaths {
		tmpl, err = template.ParseFiles(paths...)
		if err == nil {
			break
		}
		if templateErr == nil {
			templateErr = err
		}
	}

	if tmpl == nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing login template: %v", templateErr)
		return
	}

	// Render the template
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		log.Printf("Error executing login template: %v", err)
		return
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Parse the form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get email and password from form
	email := r.FormValue("email")
	password := r.FormValue("password")
	rememberMe := r.FormValue("remember") == "on"

	// Validate inputs
	if email == "" || password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Find user by email
	ctx := r.Context()
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		// Use generic error message to prevent email enumeration
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Verify password using bcrypt
	if !utils.VerifyPassword(user.PasswordHash, password) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Create a new session
	sessionToken := generateSecureToken()
	expiresAt := time.Now().Add(24 * time.Hour) // Default expiry time

	if rememberMe {
		expiresAt = time.Now().Add(30 * 24 * time.Hour) // 30 days if "remember me" is checked
	}

	// Create the session in the database
	session := &models.Session{
		ID:           generateID(),
		UserID:       user.ID,
		Token:        sessionToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    expiresAt,
		LastActivity: time.Now(),
		IPAddress:    r.RemoteAddr,
		UserAgent:    r.UserAgent(),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		log.Printf("Error creating session: %v", err)
		return
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  expiresAt,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil, // Set Secure flag if connection is HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	// Create audit log entry
	auditLog := &models.AuditLog{
		ID:        generateID(),
		UserID:    user.ID,
		Action:    "login",
		Timestamp: time.Now(),
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Details:   "User login",
	}

	if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
		// Non-fatal error, just log it
		log.Printf("Error creating audit log for login: %v", err)
	}

	// Redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

func (s *Server) handleRegisterForm(w http.ResponseWriter, r *http.Request) {
	// Try to find templates in multiple locations
	templatePaths := [][]string{
		{"/app/web/templates/layout.html", "/app/web/templates/register.html"},
		{"./web/templates/layout.html", "./web/templates/register.html"},
	}

	var tmpl *template.Template
	var err error
	var templateErr error

	for _, paths := range templatePaths {
		tmpl, err = template.ParseFiles(paths...)
		if err == nil {
			break
		}
		if templateErr == nil {
			templateErr = err
		}
	}

	if tmpl == nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing register template: %v", templateErr)
		return
	}

	data := map[string]interface{}{
		"Title":           "Register",
		"ActivePage":      "register",
		"IsAuthenticated": false,
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		log.Printf("Error executing register template: %v", err)
		return
	}
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	// Parse the form
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get user data from form
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

	// Check if email already exists
	ctx := r.Context()
	_, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil {
		// Email already exists
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		log.Printf("Error hashing password: %v", err)
		return
	}

	// Create new user
	user := &models.User{
		ID:                generateID(),
		Email:             email,
		PasswordHash:      hashedPassword,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
		LastActivity:      time.Now().UTC(),
		PingFrequency:     7,  // Default: check in every 7 days
		PingDeadline:      14, // Default: 14 days after last activity to trigger switch
		PingingEnabled:    false,
		NextScheduledPing: time.Time{},
	}

	// Save user to database
	if err := s.repo.CreateUser(ctx, user); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		log.Printf("Error creating user: %v", err)
		return
	}

	// Send welcome email
	welcomeMessage := fmt.Sprintf(`
		Hello %s,

		Welcome to Dead Man's Switch!

		Your account has been created successfully.

		Best regards,
		Dead Man's Switch Team
	`, name)

	// Send welcome email
	if err := s.sendEmail(
		[]string{email},
		"Welcome to Dead Man's Switch",
		welcomeMessage,
		false,
	); err != nil {
		log.Printf("Error sending welcome email: %v", err)
		// Continue anyway, this is not critical
	}

	// Create audit log entry
	auditLog := &models.AuditLog{
		ID:        generateID(),
		UserID:    user.ID,
		Action:    "register",
		Timestamp: time.Now().UTC(),
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Details:   "User registration",
	}

	if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
		// Non-fatal error, just log it
		log.Printf("Error creating audit log for registration: %v", err)
	}

	// Redirect to login page with a success message
	http.Redirect(w, r, "/login?registered=true", http.StatusFound)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	ctx := r.Context()
	user, ok := ctx.Value("user").(*models.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Try to find templates in multiple locations
	templatePaths := [][]string{
		{"/app/web/templates/layout.html", "/app/web/templates/dashboard.html"},
		{"./web/templates/layout.html", "./web/templates/dashboard.html"},
	}

	var tmpl *template.Template
	var err error
	var templateErr error

	for _, paths := range templatePaths {
		tmpl, err = template.ParseFiles(paths...)
		if err == nil {
			break
		}
		if templateErr == nil {
			templateErr = err
		}
	}

	if tmpl == nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing dashboard template: %v", templateErr)
		return
	}

	// Get actual counts from database
	secretCount := 0
	recipientCount := 0

	// Calculate days active
	daysActive := int(time.Since(user.CreatedAt).Hours() / 24)
	if daysActive < 1 {
		daysActive = 1 // At least 1 day
	}

	// Calculate next check-in time
	nextCheckIn := user.LastActivity.AddDate(0, 0, user.PingFrequency)

	data := map[string]interface{}{
		"Title":           "Dashboard",
		"ActivePage":      "dashboard",
		"IsAuthenticated": true,
		"User": map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		"Status":      "active",
		"NextCheckIn": nextCheckIn.Format("Jan 2, 2006 15:04 MST"),
		"Stats": map[string]interface{}{
			"TotalSecrets":     secretCount,
			"ActiveRecipients": recipientCount,
			"DaysActive":       daysActive,
		},
		"Activities": []map[string]string{
			{
				"Time":        user.CreatedAt.Format("Jan 2, 2006 15:04"),
				"Description": "Account created",
			},
		},
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		log.Printf("Error executing dashboard template: %v", err)
		return
	}
}

func (s *Server) handleListSecrets(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	ctx := r.Context()
	user, ok := ctx.Value("user").(*models.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Try to find templates in multiple locations
	templatePaths := [][]string{
		{"/app/web/templates/layout.html", "/app/web/templates/secrets.html"},
		{"./web/templates/layout.html", "./web/templates/secrets.html"},
	}

	var tmpl *template.Template
	var err error
	var templateErr error

	for _, paths := range templatePaths {
		tmpl, err = template.ParseFiles(paths...)
		if err == nil {
			break
		}
		if templateErr == nil {
			templateErr = err
		}
	}

	if tmpl == nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing secrets template: %v", templateErr)
		return
	}

	// In a real implementation, we would fetch the user's secrets from the database
	// For now, we'll just show an empty list
	secrets := []map[string]interface{}{}

	data := map[string]interface{}{
		"Title":           "My Secrets",
		"ActivePage":      "secrets",
		"IsAuthenticated": true,
		"User": map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		"Data": map[string]interface{}{
			"Secrets": secrets,
		},
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		log.Printf("Error executing secrets template: %v", err)
		return
	}
}

func (s *Server) handleNewSecretForm(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	ctx := r.Context()
	user, ok := ctx.Value("user").(*models.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Try to find templates in multiple locations
	templatePaths := [][]string{
		{"/app/web/templates/layout.html", "/app/web/templates/new-secret.html"},
		{"./web/templates/layout.html", "./web/templates/new-secret.html"},
	}

	var tmpl *template.Template
	var err error
	var templateErr error

	for _, paths := range templatePaths {
		tmpl, err = template.ParseFiles(paths...)
		if err == nil {
			break
		}
		if templateErr == nil {
			templateErr = err
		}
	}

	if tmpl == nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing new-secret template: %v", templateErr)
		return
	}

	// In a real implementation, we would fetch the user's recipients from the database
	// For now, we'll just show an empty list
	recipients := []map[string]interface{}{}

	data := map[string]interface{}{
		"Title":           "Add New Secret",
		"ActivePage":      "secrets",
		"IsAuthenticated": true,
		"User": map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		"Data": map[string]interface{}{
			"Recipients": recipients,
		},
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		log.Printf("Error executing new-secret template: %v", err)
		return
	}
}

func (s *Server) handleNewSecret(w http.ResponseWriter, r *http.Request) {
	// This will be implemented later
	http.Redirect(w, r, "/secrets", http.StatusFound)
}

func (s *Server) handleListRecipients(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	ctx := r.Context()
	user, ok := ctx.Value("user").(*models.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Try to find templates in multiple locations
	templatePaths := [][]string{
		{"/app/web/templates/layout.html", "/app/web/templates/recipients.html"},
		{"./web/templates/layout.html", "./web/templates/recipients.html"},
	}

	// Customize the function map to include formatDate
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("Jan 2, 2006")
		},
	}

	var tmpl *template.Template
	var err error
	var templateErr error

	// Try each template path with the function map
	for _, paths := range templatePaths {
		tmpl, err = template.New(paths[0]).Funcs(funcMap).ParseFiles(paths...)
		if err == nil {
			break
		}
		if templateErr == nil {
			templateErr = err
		}
	}

	if tmpl == nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing recipients template: %v", templateErr)
		return
	}

	// In a real implementation, we would fetch the user's recipients from the database
	// For now, we'll just show an empty list
	recipients := []map[string]interface{}{}

	data := map[string]interface{}{
		"Title":           "Recipients",
		"ActivePage":      "recipients",
		"IsAuthenticated": true,
		"User": map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
		"Recipients": recipients,
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		log.Printf("Error executing recipients template: %v", err)
		return
	}
}

func (s *Server) handleNewRecipientForm(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from context
	ctx := r.Context()
	user, ok := ctx.Value("user").(*models.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Try to find templates in multiple locations
	templatePaths := [][]string{
		{"/app/web/templates/layout.html", "/app/web/templates/new-recipient.html"},
		{"./web/templates/layout.html", "./web/templates/new-recipient.html"},
	}

	var tmpl *template.Template
	var err error
	var templateErr error

	for _, paths := range templatePaths {
		tmpl, err = template.ParseFiles(paths...)
		if err == nil {
			break
		}
		if templateErr == nil {
			templateErr = err
		}
	}

	if tmpl == nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing new-recipient template: %v", templateErr)
		return
	}

	data := map[string]interface{}{
		"Title":           "Add Recipient",
		"ActivePage":      "recipients",
		"IsAuthenticated": true,
		"User": map[string]interface{}{
			"Email": user.Email,
			"Name":  user.Email, // Use email as name since we don't have a separate name field
		},
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		log.Printf("Error executing new-recipient template: %v", err)
		return
	}
}

func (s *Server) handleNewRecipient(w http.ResponseWriter, r *http.Request) {
	// This will be implemented later
	http.Redirect(w, r, "/recipients", http.StatusFound)
}

func (s *Server) handleCheckIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get the authenticated user from context
	user, ok := ctx.Value("user").(*models.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Update the user's last activity time
	user.LastActivity = time.Now()

	// Calculate the next scheduled ping based on the user's ping frequency
	user.NextScheduledPing = time.Now().AddDate(0, 0, user.PingFrequency)

	// Enable pinging if it wasn't already enabled
	if !user.PingingEnabled {
		user.PingingEnabled = true
	}

	// Update the user in the database
	if err := s.repo.UpdateUser(ctx, user); err != nil {
		log.Printf("Error updating user during check-in: %v", err)
		http.Error(w, "Failed to update check-in status", http.StatusInternalServerError)
		return
	}

	// Create a ping history entry
	pingHistory := &models.PingHistory{
		ID:          generateID(),
		UserID:      user.ID,
		SentAt:      time.Now(),
		Method:      "web",
		Status:      "responded",
		RespondedAt: &user.LastActivity,
	}

	if err := s.repo.CreatePingHistory(ctx, pingHistory); err != nil {
		log.Printf("Error creating ping history during check-in: %v", err)
		// Non-fatal error, continue
	}

	// Create audit log entry
	auditLog := &models.AuditLog{
		ID:        generateID(),
		UserID:    user.ID,
		Action:    "check_in",
		Timestamp: time.Now(),
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Details:   "User check-in via web",
	}

	if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
		log.Printf("Error creating audit log for check-in: %v", err)
		// Non-fatal error, continue
	}

	// Return JSON response with next check-in time
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	nextCheckIn := user.NextScheduledPing.Format("Jan 2, 2006 15:04 MST")
	fmt.Fprintf(w, `{"success":true,"message":"Check-in successful","nextCheckIn":"%s"}`, nextCheckIn)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Get the session token from cookie
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Delete the session from database
		ctx := r.Context()
		session, err := s.repo.GetSessionByToken(ctx, cookie.Value)
		if err == nil {
			// Create audit log entry
			user := ctx.Value("user").(*models.User)
			if user != nil {
				auditLog := &models.AuditLog{
					ID:        generateID(),
					UserID:    user.ID,
					Action:    "logout",
					Timestamp: time.Now(),
					IPAddress: r.RemoteAddr,
					UserAgent: r.UserAgent(),
					Details:   "User logout",
				}

				if err := s.repo.CreateAuditLog(ctx, auditLog); err != nil {
					// Non-fatal error, just log it
					log.Printf("Error creating audit log for logout: %v", err)
				}
			}

			// Delete the session
			if err := s.repo.DeleteSession(ctx, session.ID); err != nil {
				log.Printf("Error deleting session during logout: %v", err)
			}
		}
	}

	// Delete the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusFound)
}

// Helper functions

func generateSecureToken() string {
	// Generate a random 32-byte token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(tokenBytes)
}

func generateID() string {
	// Generate a UUID-like ID
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(idBytes)
}

// Adapter functions to convert between models and UI types

// adaptUserToUI converts a models.User to a format suitable for templates
func adaptUserToUI(user *models.User) map[string]interface{} {
	return map[string]interface{}{
		"ID":                user.ID,
		"Email":             user.Email,
		"Name":              user.Email, // Using email as name since name is not in the User model
		"LastActivity":      user.LastActivity,
		"CreatedAt":         user.CreatedAt,
		"PingFrequency":     user.PingFrequency,
		"PingDeadline":      user.PingDeadline,
		"PingingEnabled":    user.PingingEnabled,
		"NextScheduledPing": user.NextScheduledPing,
	}
}

// adaptSecretToUI converts a models.Secret to a format suitable for templates
func adaptSecretToUI(secret *models.Secret) map[string]interface{} {
	// For real implementation, we would decode the encrypted data to determine the type
	// For now, we'll use a placeholder approach
	secretType := determineSecretType(secret)

	return map[string]interface{}{
		"ID":           secret.ID,
		"Title":        secret.Name,
		"Type":         secretType,
		"Description":  "",                   // Not in base model, would be in metadata
		"Content":      secret.EncryptedData, // In real impl, this would be decrypted
		"CreatedAt":    secret.CreatedAt,
		"LastModified": secret.UpdatedAt,
	}
}

// determineSecretType attempts to determine the type of secret from its encrypted data
// In a real implementation, this would be stored in metadata or determined by decryption
func determineSecretType(secret *models.Secret) string {
	// Store the secret type for categorization
	// in metadata or could be determined from the decrypted data
	return "note" // Default type
}

// adaptRecipientToUI converts a models.Recipient to a format suitable for templates
func adaptRecipientToUI(recipient *models.Recipient) map[string]interface{} {
	return map[string]interface{}{
		"ID":            recipient.ID,
		"Name":          recipient.Name,
		"Email":         recipient.Email,
		"PhoneNumber":   recipient.PhoneNumber,
		"CreatedAt":     recipient.CreatedAt,
		"Relationship":  "other", // Default value, not in the base model
		"ContactMethod": determineContactMethod(recipient),
		"Verified":      true, // Default value, not in the base model
	}
}

// determineContactMethod determines the contact method based on recipient data
func determineContactMethod(recipient *models.Recipient) string {
	if recipient.PhoneNumber != "" {
		return "phone"
	}
	return "email" // Default contact method
}

// sendEmail is a helper method to send emails
func (s *Server) sendEmail(to []string, subject, body string, isHTML bool) error {
	if s.emailClient == nil {
		return fmt.Errorf("email client not configured")
	}

	// Use the simplified email sending method
	return s.emailClient.SendEmailSimple(to, subject, body, isHTML)
}
