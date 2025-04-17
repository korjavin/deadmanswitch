package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/korjavin/deadmanswitch/internal/config"
	"github.com/korjavin/deadmanswitch/internal/email"
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

		// Verification
		r.Get("/verify/{code}", s.handleVerify)

		// Recipient access to shared secrets
		r.Get("/access/{code}", s.handleAccessForm)
		r.Post("/access/{code}", s.handleAccess)
	})

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(s.authMiddleware)

		// Dashboard
		r.Get("/dashboard", s.handleDashboard)

		// User settings
		r.Get("/settings", s.handleSettingsForm)
		r.Post("/settings", s.handleSettings)
		r.Post("/logout", s.handleLogout)

		// Secrets management
		r.Get("/secrets", s.handleListSecrets)
		r.Get("/secrets/new", s.handleNewSecretForm)
		r.Post("/secrets/new", s.handleNewSecret)
		r.Get("/secrets/{id}", s.handleViewSecret)
		r.Post("/secrets/{id}", s.handleUpdateSecret)
		r.Delete("/secrets/{id}", s.handleDeleteSecret)

		// Recipients management
		r.Get("/recipients", s.handleListRecipients)
		r.Get("/recipients/new", s.handleNewRecipientForm)
		r.Post("/recipients/new", s.handleNewRecipient)
		r.Get("/recipients/{id}", s.handleViewRecipient)
		r.Post("/recipients/{id}", s.handleUpdateRecipient)
		r.Delete("/recipients/{id}", s.handleDeleteRecipient)

		// Assignments (connecting secrets to recipients)
		r.Post("/assign", s.handleCreateAssignment)
		r.Delete("/assign/{id}", s.handleDeleteAssignment)

		// Status and history
		r.Get("/history", s.handleViewHistory)
		r.Post("/ping", s.handleManualPing)

		// Telegram integration
		r.Get("/telegram", s.handleTelegramConnectionForm)
		r.Post("/telegram", s.handleTelegramConnection)
	})

	// API routes (for frontend JavaScript)
	r.Route("/api", func(r chi.Router) {
		r.Use(s.authMiddleware)

		r.Get("/user", s.handleGetUser)
		r.Get("/secrets", s.handleGetSecrets)
		r.Get("/recipients", s.handleGetRecipients)
		r.Get("/assignments", s.handleGetAssignments)
		r.Get("/ping-history", s.handleGetPingHistory)
	})

	// Static files
	fileServer := http.FileServer(http.Dir("./web/static"))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))
}

// authMiddleware checks if the user is authenticated
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the session token from the cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Get the session from the database
		ctx := r.Context()
		session, err := s.repo.GetSessionByToken(ctx, cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Check if the session is expired
		if session.ExpiresAt.Before(time.Now()) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Update the session's last activity
		if err := s.repo.UpdateSessionActivity(ctx, session.ID); err != nil {
			log.Printf("Error updating session activity: %v", err)
		}

		// Get the user
		user, err := s.repo.GetUserByID(ctx, session.UserID)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// Update the user's last activity time
		user.LastActivity = time.Now().UTC()
		if err := s.repo.UpdateUser(ctx, user); err != nil {
			log.Printf("Error updating user activity: %v", err)
		}

		// Store user and session in the request context
		ctx = context.WithValue(ctx, "user", user)
		ctx = context.WithValue(ctx, "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Route handlers (stubs for now)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to Dead Man's Switch")
}

func (s *Server) handleLoginForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Login form")
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Login handler")
}

func (s *Server) handleRegisterForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Register form")
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Register handler")
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	fmt.Fprintf(w, "Verify handler for code: %s", code)
}

func (s *Server) handleAccessForm(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	fmt.Fprintf(w, "Access form for code: %s", code)
}

func (s *Server) handleAccess(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	fmt.Fprintf(w, "Access handler for code: %s", code)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Dashboard")
}

func (s *Server) handleSettingsForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Settings form")
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Settings handler")
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Logout handler")
}

func (s *Server) handleListSecrets(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "List secrets")
}

func (s *Server) handleNewSecretForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "New secret form")
}

func (s *Server) handleNewSecret(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "New secret handler")
}

func (s *Server) handleViewSecret(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fmt.Fprintf(w, "View secret handler for ID: %s", id)
}

func (s *Server) handleUpdateSecret(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fmt.Fprintf(w, "Update secret handler for ID: %s", id)
}

func (s *Server) handleDeleteSecret(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fmt.Fprintf(w, "Delete secret handler for ID: %s", id)
}

func (s *Server) handleListRecipients(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "List recipients")
}

func (s *Server) handleNewRecipientForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "New recipient form")
}

func (s *Server) handleNewRecipient(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "New recipient handler")
}

func (s *Server) handleViewRecipient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fmt.Fprintf(w, "View recipient handler for ID: %s", id)
}

func (s *Server) handleUpdateRecipient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fmt.Fprintf(w, "Update recipient handler for ID: %s", id)
}

func (s *Server) handleDeleteRecipient(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fmt.Fprintf(w, "Delete recipient handler for ID: %s", id)
}

func (s *Server) handleCreateAssignment(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Create assignment handler")
}

func (s *Server) handleDeleteAssignment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	fmt.Fprintf(w, "Delete assignment handler for ID: %s", id)
}

func (s *Server) handleViewHistory(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "View history handler")
}

func (s *Server) handleManualPing(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Manual ping handler")
}

func (s *Server) handleTelegramConnectionForm(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Telegram connection form")
}

func (s *Server) handleTelegramConnection(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Telegram connection handler")
}

// API handlers

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Get user API handler")
}

func (s *Server) handleGetSecrets(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Get secrets API handler")
}

func (s *Server) handleGetRecipients(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Get recipients API handler")
}

func (s *Server) handleGetAssignments(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Get assignments API handler")
}

func (s *Server) handleGetPingHistory(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Get ping history API handler")
}
