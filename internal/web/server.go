package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/korjavin/deadmanswitch/internal/auth"
	"github.com/korjavin/deadmanswitch/internal/config"
	"github.com/korjavin/deadmanswitch/internal/email"
	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/scheduler"
	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/telegram"
	"github.com/korjavin/deadmanswitch/internal/web/handlers"
	authMiddleware "github.com/korjavin/deadmanswitch/internal/web/middleware"
	"github.com/korjavin/deadmanswitch/internal/web/templates"
	"github.com/korjavin/deadmanswitch/internal/web/utils"
)

// Server represents the web server
type Server struct {
	config      *config.Config
	repo        storage.Repository
	emailClient *email.Client
	telegramBot *telegram.Bot
	scheduler   *scheduler.Scheduler
	router      *http.ServeMux
	httpServer  *http.Server
	handlers    struct {
		index           *handlers.IndexHandler
		auth            *handlers.AuthHandler
		dashboard       *handlers.DashboardHandler
		secrets         *handlers.SecretsHandler
		recipients      *handlers.RecipientsHandler
		api             *handlers.APIHandler
		profile         *handlers.ProfileHandler
		settings        *handlers.SettingsHandler
		history         *handlers.HistoryHandler
		twofa           *handlers.TwoFAHandler
		passkey         *handlers.PasskeyHandler
		secretQuestions *handlers.SecretQuestionsHandler
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
		router:      http.NewServeMux(),
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
	server.handlers.settings = handlers.NewSettingsHandler(repo)
	server.handlers.history = handlers.NewHistoryHandler(repo)
	server.handlers.twofa = handlers.NewTwoFAHandler(repo)
	server.handlers.passkey = handlers.NewPasskeyHandler(repo, webAuthnService)
	server.handlers.secretQuestions = handlers.NewSecretQuestionsHandler(repo, templates.NewTemplateRenderer())

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

	// Public routes
	r.HandleFunc("/", s.handlers.index.HandleIndex)
	r.HandleFunc("/login", s.handleMethodRouter(
		"GET", s.handlers.auth.HandleLoginForm,
		"POST", s.handlers.auth.HandleLogin,
	))
	r.HandleFunc("/register", s.handleMethodRouter(
		"GET", s.handlers.auth.HandleRegisterForm,
		"POST", s.handlers.auth.HandleRegister,
	))
	r.HandleFunc("/login/passkey/begin", s.handlers.passkey.HandleBeginLogin)
	r.HandleFunc("/login/passkey/finish", s.handlers.passkey.HandleFinishLogin)
	r.HandleFunc("/confirm/", s.handleConfirmation)
	r.Handle("/static/", http.StripPrefix("/static/", s.setupFileServer()))
	r.HandleFunc("/logout", s.handlers.auth.HandleLogout)

	// Protected routes
	r.HandleFunc("/dashboard", authMiddleware.Auth(s.repo)(s.handlers.dashboard.HandleDashboard))
	r.HandleFunc("/secrets", authMiddleware.Auth(s.repo)(s.handlers.secrets.HandleListSecrets))
	r.HandleFunc("/secrets/new", authMiddleware.Auth(s.repo)(s.handleMethodRouter(
		"GET", s.handlers.secrets.HandleNewSecretForm,
		"POST", s.handlers.secrets.HandleCreateSecret,
	)))
	r.HandleFunc("/secrets/", authMiddleware.Auth(s.repo)(s.handleSecrets))
	r.HandleFunc("/recipients", authMiddleware.Auth(s.repo)(s.handlers.recipients.HandleListRecipients))
	r.HandleFunc("/recipients/new", authMiddleware.Auth(s.repo)(s.handleMethodRouter(
		"GET", s.handlers.recipients.HandleNewRecipientForm,
		"POST", s.handlers.recipients.HandleCreateRecipient,
	)))
	r.HandleFunc("/recipients/", authMiddleware.Auth(s.repo)(s.handleRecipients))
	r.HandleFunc("/profile", authMiddleware.Auth(s.repo)(s.handleMethodRouter(
		"GET", s.handlers.profile.HandleProfile,
		"POST", s.handlers.profile.HandleUpdateProfile,
	)))
	r.HandleFunc("/profile/github/disconnect", authMiddleware.Auth(s.repo)(s.handlers.profile.HandleDisconnectGitHub))
	r.HandleFunc("/profile/passkeys", authMiddleware.Auth(s.repo)(s.handlers.passkey.HandlePasskeyManagement))
	r.HandleFunc("/profile/passkeys/register/begin", authMiddleware.Auth(s.repo)(s.handlers.passkey.HandleBeginRegistration))
	r.HandleFunc("/profile/passkeys/register/finish", authMiddleware.Auth(s.repo)(s.handlers.passkey.HandleFinishRegistration))
	r.HandleFunc("/profile/passkeys/", authMiddleware.Auth(s.repo)(s.handlePasskeys))
	r.HandleFunc("/settings", authMiddleware.Auth(s.repo)(s.handlers.settings.HandleSettings))
	r.HandleFunc("/settings/deadmanswitch", authMiddleware.Auth(s.repo)(s.handlers.settings.HandleUpdateDeadManSwitchSettings))
	r.HandleFunc("/settings/notifications", authMiddleware.Auth(s.repo)(s.handlers.settings.HandleUpdateNotificationSettings))
	r.HandleFunc("/settings/security", authMiddleware.Auth(s.repo)(s.handlers.settings.HandleUpdateSecuritySettings))
	r.HandleFunc("/2fa/setup", authMiddleware.Auth(s.repo)(s.handlers.twofa.HandleSetup))
	r.HandleFunc("/2fa/verify", authMiddleware.Auth(s.repo)(s.handlers.twofa.HandleVerify))
	r.HandleFunc("/2fa/disable", authMiddleware.Auth(s.repo)(s.handlers.twofa.HandleDisable))
	r.HandleFunc("/history", authMiddleware.Auth(s.repo)(s.handlers.history.HandleHistory))
	r.HandleFunc("/api/check-in", authMiddleware.Auth(s.repo)(s.handlers.api.HandleCheckIn))
}

// Helper functions for routing

func (s *Server) handleMethodRouter(methods ...interface{}) http.HandlerFunc {
	handlers := make(map[string]http.HandlerFunc)
	for i := 0; i < len(methods); i += 2 {
		method := methods[i].(string)
		switch h := methods[i+1].(type) {
		case http.HandlerFunc:
			handlers[method] = h
		case func(http.ResponseWriter, *http.Request):
			handlers[method] = http.HandlerFunc(h)
		default:
			panic(fmt.Sprintf("unsupported handler type: %T", methods[i+1]))
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if handler, ok := handlers[r.Method]; ok {
			handler(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleSecrets(w http.ResponseWriter, r *http.Request) {
	id := utils.GetLastURLSegment(r)
	if id == "" {
		http.NotFound(w, r)
		return
	}

	// Check if this is an "assign" request
	if strings.HasSuffix(r.URL.Path, "/assign") {
		if r.Method == http.MethodGet {
			s.handlers.secrets.HandleManageRecipients(w, r)
		} else if r.Method == http.MethodPost {
			s.handlers.secrets.HandleUpdateSecretRecipients(w, r)
		}
		return
	}

	// Handle regular secret operations
	if r.Method == http.MethodGet {
		s.handlers.secrets.HandleViewSecretForm(w, r)
	} else if r.Method == http.MethodPost {
		if r.FormValue("_method") == "DELETE" {
			s.handlers.secrets.HandleDeleteSecret(w, r)
		} else {
			s.handlers.secrets.HandleUpdateSecret(w, r)
		}
	}
}

func (s *Server) handleRecipients(w http.ResponseWriter, r *http.Request) {
	id := utils.GetLastURLSegment(r)
	if id == "" {
		http.NotFound(w, r)
		return
	}

	// Handle test contact request
	if strings.HasSuffix(r.URL.Path, "/test") {
		s.handlers.recipients.HandleTestContact(w, r)
		return
	}

	// Handle secrets management
	if strings.HasSuffix(r.URL.Path, "/secrets") {
		if r.Method == http.MethodGet {
			s.handlers.recipients.HandleManageSecrets(w, r)
		} else if r.Method == http.MethodPost {
			s.handlers.recipients.HandleUpdateRecipientSecrets(w, r)
		}
		return
	}

	// Handle questions management
	if strings.HasSuffix(r.URL.Path, "/questions") {
		if r.Method == http.MethodGet {
			s.handlers.secretQuestions.ShowQuestionsPage(w, r)
		} else if r.Method == http.MethodPost {
			s.handlers.secretQuestions.CreateQuestions(w, r)
		}
		return
	}

	// Handle question operations
	if strings.Contains(r.URL.Path, "/questions/") {
		// Extract the question ID
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 3 {
			// Check if this is a delete request
			if strings.HasSuffix(r.URL.Path, "/delete") {
				s.handlers.secretQuestions.DeleteQuestion(w, r)
				return
			}

			// Handle update question
			s.handlers.secretQuestions.UpdateQuestion(w, r)
			return
		}
	}

	// Handle regular recipient operations
	if r.Method == http.MethodGet {
		s.handlers.recipients.HandleEditRecipientForm(w, r)
	} else if r.Method == http.MethodPost {
		if r.FormValue("_method") == "DELETE" {
			s.handlers.recipients.HandleDeleteRecipient(w, r)
		} else {
			s.handlers.recipients.HandleUpdateRecipient(w, r)
		}
	}
}

func (s *Server) handlePasskeys(w http.ResponseWriter, r *http.Request) {
	id := utils.GetLastURLSegment(r)
	if id == "" {
		http.NotFound(w, r)
		return
	}
	s.handlers.passkey.HandleDeletePasskey(w, r)
}

func (s *Server) handleConfirmation(w http.ResponseWriter, r *http.Request) {
	code := utils.GetLastURLSegment(r)
	if code == "" {
		http.NotFound(w, r)
		return
	}
	s.handlers.recipients.HandleConfirmRecipient(w, r)
}

func (s *Server) setupFileServer() http.Handler {
	// Static files - try multiple paths
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

	return fileServer
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
