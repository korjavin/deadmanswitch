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
	"github.com/korjavin/deadmanswitch/internal/web/router"
	"github.com/korjavin/deadmanswitch/internal/web/templates"
)

// ServerNew represents the web server with the new router
type ServerNew struct {
	config      *config.Config
	repo        storage.Repository
	emailClient *email.Client
	telegramBot *telegram.Bot
	scheduler   *scheduler.Scheduler
	router      *router.Router
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

// NewServerWithRouter creates a new web server with the new router
func NewServerWithRouter(
	cfg *config.Config,
	repo storage.Repository,
	emailClient *email.Client,
	telegramBot *telegram.Bot,
	scheduler *scheduler.Scheduler,
) *ServerNew {
	server := &ServerNew{
		config:      cfg,
		repo:        repo,
		emailClient: emailClient,
		telegramBot: telegramBot,
		scheduler:   scheduler,
		router:      router.New(repo),
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
func (s *ServerNew) Start() error {
	// Configure the HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083" // Use a different port to avoid conflicts
	}
	s.httpServer = &http.Server{
		Addr:         ":" + port,
		Handler:      s.router.Handler(),
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
func (s *ServerNew) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// setupRoutes configures all the routes for the server
func (s *ServerNew) setupRoutes() {
	// Set up static file handler
	staticHandler := s.setupFileServer()

	// Register public routes
	s.router.RegisterPublicRoutes(
		s.handlers.index,
		s.handlers.auth,
		s.handlers.passkey,
		s.handlers.recipients,
		http.StripPrefix("/static/", staticHandler),
	)

	// Register protected routes
	s.router.RegisterProtectedRoutes(
		s.handlers.dashboard,
		s.handlers.secrets,
		s.handlers.recipients,
		s.handlers.api,
		s.handlers.profile,
		s.handlers.settings,
		s.handlers.history,
		s.handlers.twofa,
		s.handlers.passkey,
		s.handlers.secretQuestions,
	)
}

func (s *ServerNew) setupFileServer() http.Handler {
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
func (s *ServerNew) sendEmail(to []string, subject, body string, isHTML bool) error {
	if s.emailClient == nil {
		return fmt.Errorf("email client not configured")
	}

	// Use the simplified email sending method
	return s.emailClient.SendEmailSimple(to, subject, body, isHTML)
}
