package web

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/korjavin/deadmanswitch/internal/config"
	"github.com/korjavin/deadmanswitch/internal/email"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/scheduler"
	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/telegram"
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
}

// UserWithProfile extends the User model with additional profile information
// used in the web interface
type UserWithProfile struct {
	*models.User
	Name              string
	VerificationToken string
	Verified          bool
	// Any other fields needed by the templates
}

// SecretWithMetadata extends the Secret model with additional metadata
// used in the web interface
type SecretWithMetadata struct {
	*models.Secret
	Type         string
	Description  string
	Metadata     map[string]string
	LastModified time.Time
	// Any other fields needed by the templates
}

// RecipientWithDetails extends the Recipient model with additional details
// used in the web interface
type RecipientWithDetails struct {
	*models.Recipient
	Relationship     string
	ContactMethod    string
	TelegramUsername string
	Notes            string
	Verified         bool
	// Any other fields needed by the templates
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

	// Set up routes
	server.setupRoutes()

	return server
}

// Start starts the web server
func (s *Server) Start() error {
	// Configure the HTTP server
	s.httpServer = &http.Server{
		Addr:         ":8080", // TODO: Make configurable
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
		r.Get("/", s.handleIndex)

		// Authentication
		r.Get("/login", s.handleLoginForm)
		r.Post("/login", s.handleLogin)
		r.Get("/register", s.handleRegisterForm)
		r.Post("/register", s.handleRegister)

		// Static files
		fileServer := http.FileServer(http.Dir("./web/static"))
		r.Handle("/static/*", http.StripPrefix("/static", fileServer))
	})

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(s.authMiddleware)

		// Dashboard
		r.Get("/dashboard", s.handleDashboard)

		// Secrets management
		r.Get("/secrets", s.handleListSecrets)
		r.Get("/secrets/new", s.handleNewSecretForm)
		r.Post("/secrets/new", s.handleNewSecret)

		// Recipients management
		r.Get("/recipients", s.handleListRecipients)
		r.Get("/recipients/new", s.handleNewRecipientForm)
		r.Post("/recipients/new", s.handleNewRecipient)

		// Check-in
		r.Post("/api/check-in", s.handleCheckIn)

		// Logout
		r.Get("/logout", s.handleLogout)
	})
}

// authMiddleware checks if the user is authenticated
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the session token from cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			// No session token, redirect to login
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Validate the session token
		ctx := r.Context()
		session, err := s.repo.GetSessionByToken(ctx, cookie.Value)
		if err != nil {
			// Invalid session token, redirect to login
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Check if session has expired
		if session.ExpiresAt.Before(time.Now()) {
			// Session expired, delete it and redirect to login
			_ = s.repo.DeleteSession(ctx, session.ID) // Ignore error, just try to clean up
			http.Redirect(w, r, "/login?expired=true", http.StatusFound)
			return
		}

		// Get the user associated with this session
		user, err := s.repo.GetUserByID(ctx, session.UserID)
		if err != nil {
			// User not found, delete session and redirect to login
			_ = s.repo.DeleteSession(ctx, session.ID)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Update last activity time
		user.LastActivity = time.Now()
		if err := s.repo.UpdateUser(ctx, user); err != nil {
			log.Printf("Error updating user last activity: %v", err)
			// Continue anyway, this is not critical
		}

		// Update session activity
		if err := s.repo.UpdateSessionActivity(ctx, session.ID); err != nil {
			log.Printf("Error updating session activity: %v", err)
			// Continue anyway, this is not critical
		}

		// Add user and session to context for handlers to use
		ctx = context.WithValue(ctx, "user", user)
		ctx = context.WithValue(ctx, "session", session)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Basic route handlers

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	// For now, just render a simple index page
	tmpl, err := template.ParseFiles(
		"./web/templates/layout.html",
		"./web/templates/index.html",
	)
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

	// Parse the login page template
	tmpl, err := template.ParseFiles(
		"./web/templates/layout.html",
		"./web/templates/login.html",
	)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing login template: %v", err)
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

	// Verify password (using a secure password comparison)
	// Note: In a real implementation, you'd use a library like bcrypt to compare hashed passwords
	// This would need to be adapted based on how passwords are actually stored
	if !verifyPassword(user.PasswordHash, password) {
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

// verifyPassword verifies a password against a hashed password
// This is a placeholder; in a real implementation you would use bcrypt or similar
func verifyPassword(hashedPassword []byte, password string) bool {
	// TODO: Replace with proper bcrypt or similar password checking
	// Example of proper implementation:
	// return bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)) == nil

	// Simulated check - DO NOT USE IN PRODUCTION
	// This is just for testing/development
	return string(hashedPassword) == password
}

func (s *Server) handleRegisterForm(w http.ResponseWriter, r *http.Request) {
	// For now, just render the register page
	tmpl, err := template.ParseFiles(
		"./web/templates/layout.html",
		"./web/templates/register.html",
	)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing register template: %v", err)
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
	termsAgreed := r.FormValue("terms") == "on"

	// Validate inputs
	if email == "" || name == "" || password == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	if password != confirmPassword {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	if !termsAgreed {
		http.Error(w, "You must agree to the terms and conditions", http.StatusBadRequest)
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

	// Create new user
	// In a real implementation, the password would be hashed with bcrypt or similar
	user := &models.User{
		ID:                generateID(),
		Email:             email,
		PasswordHash:      []byte(password), // Should be hashed in production
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

	// Using the correct method signature for the email client
	emailOpts := &email.MessageOptions{
		To:      []string{email},
		Subject: "Welcome to Dead Man's Switch",
		Body:    welcomeMessage,
	}
	err = s.emailClient.SendEmail(emailOpts)

	if err != nil {
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
	// Simple dashboard implementation
	tmpl, err := template.ParseFiles(
		"./web/templates/layout.html",
		"./web/templates/dashboard.html",
	)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing dashboard template: %v", err)
		return
	}

	data := map[string]interface{}{
		"Title":           "Dashboard",
		"ActivePage":      "dashboard",
		"IsAuthenticated": true,
		"User": map[string]interface{}{
			"Email": "user@example.com",
		},
		"Status":      "active",
		"NextCheckIn": time.Now().Add(7 * 24 * time.Hour).Format("Jan 2, 2006 15:04 MST"),
		"Stats": map[string]interface{}{
			"TotalSecrets":     3,
			"ActiveRecipients": 2,
			"DaysActive":       30,
		},
		"Activities": []map[string]string{
			{
				"Time":        time.Now().Add(-24 * time.Hour).Format("Jan 2, 2006 15:04"),
				"Description": "Check-in via web",
			},
			{
				"Time":        time.Now().Add(-8 * 24 * time.Hour).Format("Jan 2, 2006 15:04"),
				"Description": "Check-in via email",
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
	// Placeholder for listing secrets
	tmpl, err := template.ParseFiles(
		"./web/templates/layout.html",
		"./web/templates/secrets.html",
	)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing secrets template: %v", err)
		return
	}

	data := map[string]interface{}{
		"Title":           "My Secrets",
		"ActivePage":      "secrets",
		"IsAuthenticated": true,
		"User": map[string]interface{}{
			"Email": "user@example.com",
		},
		// Example secrets data
		"Secrets": []map[string]interface{}{
			{
				"ID":          "1",
				"Title":       "Bank Account",
				"Type":        "login",
				"Description": "My main bank account credentials",
				"Username":    "myusername",
				"Recipients": []map[string]interface{}{
					{"Name": "John Doe", "Email": "john@example.com"},
				},
			},
		},
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		log.Printf("Error executing secrets template: %v", err)
		return
	}
}

func (s *Server) handleNewSecretForm(w http.ResponseWriter, r *http.Request) {
	// Placeholder for new secret form
	tmpl, err := template.ParseFiles(
		"./web/templates/layout.html",
		"./web/templates/new-secret.html",
	)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing new-secret template: %v", err)
		return
	}

	data := map[string]interface{}{
		"Title":           "Add New Secret",
		"ActivePage":      "secrets",
		"IsAuthenticated": true,
		"User": map[string]interface{}{
			"Email": "user@example.com",
		},
		// Example recipients data
		"Recipients": []map[string]interface{}{
			{"ID": "1", "Name": "John Doe", "Email": "john@example.com"},
			{"ID": "2", "Name": "Jane Smith", "Email": "jane@example.com"},
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
	// Placeholder for listing recipients
	tmpl, err := template.ParseFiles(
		"./web/templates/layout.html",
		"./web/templates/recipients.html",
	)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing recipients template: %v", err)
		return
	}

	// Customize the function map to include formatDate
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("Jan 2, 2006")
		},
	}

	// Parse with function map
	tmpl, err = template.New("layout.html").Funcs(funcMap).ParseFiles(
		"./web/templates/layout.html",
		"./web/templates/recipients.html",
	)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing recipients template with funcMap: %v", err)
		return
	}

	data := map[string]interface{}{
		"Title":           "Recipients",
		"ActivePage":      "recipients",
		"IsAuthenticated": true,
		"User": map[string]interface{}{
			"Email": "user@example.com",
		},
		// Example recipients data
		"Recipients": []map[string]interface{}{
			{
				"ID":              "1",
				"Name":            "John Doe",
				"Email":           "john@example.com",
				"Relationship":    "family",
				"ContactMethod":   "email",
				"CreatedAt":       time.Now().AddDate(0, -1, 0),
				"AssignedSecrets": []string{"1", "2"},
			},
			{
				"ID":              "2",
				"Name":            "Jane Smith",
				"Email":           "jane@example.com",
				"Relationship":    "friend",
				"ContactMethod":   "phone",
				"PhoneNumber":     "+1234567890",
				"CreatedAt":       time.Now().AddDate(0, -2, 0),
				"AssignedSecrets": []string{"1"},
			},
		},
	}

	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		log.Printf("Error executing recipients template: %v", err)
		return
	}
}

func (s *Server) handleNewRecipientForm(w http.ResponseWriter, r *http.Request) {
	// Placeholder for new recipient form
	tmpl, err := template.ParseFiles(
		"./web/templates/layout.html",
		"./web/templates/new-recipient.html",
	)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error parsing new-recipient template: %v", err)
		return
	}

	data := map[string]interface{}{
		"Title":           "Add Recipient",
		"ActivePage":      "recipients",
		"IsAuthenticated": true,
		"User": map[string]interface{}{
			"Email": "user@example.com",
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
	// This is a placeholder; in a real implementation, the type would be stored
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
