package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/korjavin/deadmanswitch/internal/web/handlers"
)

// RegisterPublicRoutes registers all public routes
func (r *Router) RegisterPublicRoutes(
	indexHandler *handlers.IndexHandler,
	authHandler *handlers.AuthHandler,
	passkeyHandler *handlers.PasskeyHandler,
	recipientsHandler *handlers.RecipientsHandler,
	staticHandler http.Handler,
) {
	routes := []Route{
		// Home page
		GET("home", "/", http.HandlerFunc(indexHandler.HandleIndex)),

		// Authentication
		GET("login", "/login", http.HandlerFunc(authHandler.HandleLoginForm)),
		POST("login-submit", "/login", http.HandlerFunc(authHandler.HandleLogin)),
		GET("register", "/register", http.HandlerFunc(authHandler.HandleRegisterForm)),
		POST("register-submit", "/register", http.HandlerFunc(authHandler.HandleRegister)),
		GET("logout", "/logout", http.HandlerFunc(authHandler.HandleLogout)),

		// Passkey authentication
		GET("passkey-login-begin", "/login/passkey/begin", http.HandlerFunc(passkeyHandler.HandleBeginLogin)),
		POST("passkey-login-finish", "/login/passkey/finish", http.HandlerFunc(passkeyHandler.HandleFinishLogin)),

		// Recipient confirmation
		GET("confirm-recipient", "/confirm/{code}", http.HandlerFunc(recipientsHandler.HandleConfirmRecipient)),

		// Static files
		NewRoute("static", "/static/{file:.*}", []string{"GET"}, staticHandler),
	}

	r.RegisterRoutes(routes)
}

// RegisterProtectedRoutes registers all protected routes
func (r *Router) RegisterProtectedRoutes(
	dashboardHandler *handlers.DashboardHandler,
	secretsHandler *handlers.SecretsHandler,
	recipientsHandler *handlers.RecipientsHandler,
	apiHandler *handlers.APIHandler,
	profileHandler *handlers.ProfileHandler,
	settingsHandler *handlers.SettingsHandler,
	historyHandler *handlers.HistoryHandler,
	twofaHandler *handlers.TwoFAHandler,
	passkeyHandler *handlers.PasskeyHandler,
	secretQuestionsHandler *handlers.SecretQuestionsHandler,
) {
	// Create a subrouter with auth middleware
	authMiddleware := r.AuthMiddleware()

	// Dashboard routes
	r.mux.Handle("/dashboard", authMiddleware(http.HandlerFunc(dashboardHandler.HandleDashboard))).
		Methods("GET").
		Name("dashboard")

	// Secrets routes
	r.registerSecretsRoutes(secretsHandler, authMiddleware)

	// Recipients routes
	r.registerRecipientsRoutes(recipientsHandler, secretQuestionsHandler, authMiddleware)

	// Profile routes
	r.registerProfileRoutes(profileHandler, passkeyHandler, authMiddleware)

	// Settings routes
	r.registerSettingsRoutes(settingsHandler, authMiddleware)

	// 2FA routes
	r.register2FARoutes(twofaHandler, authMiddleware)

	// History routes
	r.mux.Handle("/history", authMiddleware(http.HandlerFunc(historyHandler.HandleHistory))).
		Methods("GET").
		Name("history")

	// API routes
	r.mux.Handle("/api/check-in", authMiddleware(http.HandlerFunc(apiHandler.HandleCheckIn))).
		Methods("POST").
		Name("api-check-in")
}

// registerSecretsRoutes registers all secrets-related routes
func (r *Router) registerSecretsRoutes(secretsHandler *handlers.SecretsHandler, authMiddleware mux.MiddlewareFunc) {
	// List secrets
	r.mux.Handle("/secrets", authMiddleware(http.HandlerFunc(secretsHandler.HandleListSecrets))).
		Methods("GET").
		Name("secrets-list")

	// New secret form
	r.mux.Handle("/secrets/new", authMiddleware(http.HandlerFunc(secretsHandler.HandleNewSecretForm))).
		Methods("GET").
		Name("secrets-new-form")

	// Create secret
	r.mux.Handle("/secrets/new", authMiddleware(http.HandlerFunc(secretsHandler.HandleCreateSecret))).
		Methods("POST").
		Name("secrets-create")

	// View secret
	r.mux.Handle("/secrets/{id}", authMiddleware(http.HandlerFunc(secretsHandler.HandleViewSecretForm))).
		Methods("GET").
		Name("secrets-view")

	// Update secret
	r.mux.Handle("/secrets/{id}", authMiddleware(http.HandlerFunc(secretsHandler.HandleUpdateSecret))).
		Methods("POST").
		Name("secrets-update")

	// Delete secret (using POST with _method=DELETE)
	r.mux.Handle("/secrets/{id}", authMiddleware(http.HandlerFunc(secretsHandler.HandleDeleteSecret))).
		Methods("POST").
		HeadersRegexp("Content-Type", "application/x-www-form-urlencoded").
		MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
			return r.FormValue("_method") == "DELETE"
		}).
		Name("secrets-delete")

	// Manage recipients for a secret
	r.mux.Handle("/secrets/{id}/assign", authMiddleware(http.HandlerFunc(secretsHandler.HandleManageRecipients))).
		Methods("GET").
		Name("secrets-manage-recipients")

	// Update secret recipients
	r.mux.Handle("/secrets/{id}/assign", authMiddleware(http.HandlerFunc(secretsHandler.HandleUpdateSecretRecipients))).
		Methods("POST").
		Name("secrets-update-recipients")
}

// registerRecipientsRoutes registers all recipients-related routes
func (r *Router) registerRecipientsRoutes(
	recipientsHandler *handlers.RecipientsHandler,
	secretQuestionsHandler *handlers.SecretQuestionsHandler,
	authMiddleware mux.MiddlewareFunc,
) {
	// List recipients
	r.mux.Handle("/recipients", authMiddleware(http.HandlerFunc(recipientsHandler.HandleListRecipients))).
		Methods("GET").
		Name("recipients-list")

	// New recipient form
	r.mux.Handle("/recipients/new", authMiddleware(http.HandlerFunc(recipientsHandler.HandleNewRecipientForm))).
		Methods("GET").
		Name("recipients-new-form")

	// Create recipient
	r.mux.Handle("/recipients/new", authMiddleware(http.HandlerFunc(recipientsHandler.HandleCreateRecipient))).
		Methods("POST").
		Name("recipients-create")

	// View/edit recipient
	r.mux.Handle("/recipients/{id}", authMiddleware(http.HandlerFunc(recipientsHandler.HandleEditRecipientForm))).
		Methods("GET").
		Name("recipients-edit-form")

	// Update recipient
	r.mux.Handle("/recipients/{id}", authMiddleware(http.HandlerFunc(recipientsHandler.HandleUpdateRecipient))).
		Methods("POST").
		Name("recipients-update")

	// Delete recipient (using POST with _method=DELETE)
	r.mux.Handle("/recipients/{id}", authMiddleware(http.HandlerFunc(recipientsHandler.HandleDeleteRecipient))).
		Methods("POST").
		HeadersRegexp("Content-Type", "application/x-www-form-urlencoded").
		MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
			return r.FormValue("_method") == "DELETE"
		}).
		Name("recipients-delete")

	// Test contact with recipient
	r.mux.Handle("/recipients/{id}/test", authMiddleware(http.HandlerFunc(recipientsHandler.HandleTestContact))).
		Methods("GET").
		Name("recipients-test-contact")

	// Manage secrets for a recipient
	r.mux.Handle("/recipients/{id}/secrets", authMiddleware(http.HandlerFunc(recipientsHandler.HandleManageSecrets))).
		Methods("GET").
		Name("recipients-manage-secrets")

	// Update recipient secrets
	r.mux.Handle("/recipients/{id}/secrets", authMiddleware(http.HandlerFunc(recipientsHandler.HandleUpdateRecipientSecrets))).
		Methods("POST").
		Name("recipients-update-secrets")

	// Secret questions management
	r.mux.Handle("/recipients/{id}/questions", authMiddleware(http.HandlerFunc(secretQuestionsHandler.ShowQuestionsPage))).
		Methods("GET").
		Name("recipients-questions")

	r.mux.Handle("/recipients/{id}/questions", authMiddleware(http.HandlerFunc(secretQuestionsHandler.CreateQuestions))).
		Methods("POST").
		Name("recipients-create-questions")

	r.mux.Handle("/recipients/{id}/questions/{questionId}", authMiddleware(http.HandlerFunc(secretQuestionsHandler.UpdateQuestion))).
		Methods("POST").
		Name("recipients-update-question")

	r.mux.Handle("/recipients/{id}/questions/{questionId}/delete", authMiddleware(http.HandlerFunc(secretQuestionsHandler.DeleteQuestion))).
		Methods("POST").
		Name("recipients-delete-question")
}

// registerProfileRoutes registers all profile-related routes
func (r *Router) registerProfileRoutes(
	profileHandler *handlers.ProfileHandler,
	passkeyHandler *handlers.PasskeyHandler,
	authMiddleware mux.MiddlewareFunc,
) {
	// Profile page
	r.mux.Handle("/profile", authMiddleware(http.HandlerFunc(profileHandler.HandleProfile))).
		Methods("GET").
		Name("profile")

	// Update profile
	r.mux.Handle("/profile", authMiddleware(http.HandlerFunc(profileHandler.HandleUpdateProfile))).
		Methods("POST").
		Name("profile-update")

	// Disconnect GitHub
	r.mux.Handle("/profile/github/disconnect", authMiddleware(http.HandlerFunc(profileHandler.HandleDisconnectGitHub))).
		Methods("GET").
		Name("profile-github-disconnect")

	// Passkey management
	r.mux.Handle("/profile/passkeys", authMiddleware(http.HandlerFunc(passkeyHandler.HandlePasskeyManagement))).
		Methods("GET").
		Name("profile-passkeys")

	// Begin passkey registration
	r.mux.Handle("/profile/passkeys/register/begin", authMiddleware(http.HandlerFunc(passkeyHandler.HandleBeginRegistration))).
		Methods("GET").
		Name("profile-passkeys-register-begin")

	// Finish passkey registration
	r.mux.Handle("/profile/passkeys/register/finish", authMiddleware(http.HandlerFunc(passkeyHandler.HandleFinishRegistration))).
		Methods("POST").
		Name("profile-passkeys-register-finish")

	// Delete passkey
	r.mux.Handle("/profile/passkeys/{id}", authMiddleware(http.HandlerFunc(passkeyHandler.HandleDeletePasskey))).
		Methods("GET").
		Name("profile-passkeys-delete")
}

// registerSettingsRoutes registers all settings-related routes
func (r *Router) registerSettingsRoutes(settingsHandler *handlers.SettingsHandler, authMiddleware mux.MiddlewareFunc) {
	// Settings page
	r.mux.Handle("/settings", authMiddleware(http.HandlerFunc(settingsHandler.HandleSettings))).
		Methods("GET").
		Name("settings")

	// Update deadman switch settings
	r.mux.Handle("/settings/deadmanswitch", authMiddleware(http.HandlerFunc(settingsHandler.HandleUpdateDeadManSwitchSettings))).
		Methods("POST").
		Name("settings-deadmanswitch-update")

	// Update notification settings
	r.mux.Handle("/settings/notifications", authMiddleware(http.HandlerFunc(settingsHandler.HandleUpdateNotificationSettings))).
		Methods("POST").
		Name("settings-notifications-update")

	// Update security settings
	r.mux.Handle("/settings/security", authMiddleware(http.HandlerFunc(settingsHandler.HandleUpdateSecuritySettings))).
		Methods("POST").
		Name("settings-security-update")
}

// register2FARoutes registers all 2FA-related routes
func (r *Router) register2FARoutes(twofaHandler *handlers.TwoFAHandler, authMiddleware mux.MiddlewareFunc) {
	// 2FA setup
	r.mux.Handle("/2fa/setup", authMiddleware(http.HandlerFunc(twofaHandler.HandleSetup))).
		Methods("GET").
		Name("2fa-setup")

	// 2FA verify
	r.mux.Handle("/2fa/verify", authMiddleware(http.HandlerFunc(twofaHandler.HandleVerify))).
		Methods("POST").
		Name("2fa-verify")

	// 2FA disable
	r.mux.Handle("/2fa/disable", authMiddleware(http.HandlerFunc(twofaHandler.HandleDisable))).
		Methods("POST").
		Name("2fa-disable")
}
