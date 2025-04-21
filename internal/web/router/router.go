package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/korjavin/deadmanswitch/internal/storage"
	"github.com/korjavin/deadmanswitch/internal/web/middleware"
)

// Router represents the application router
type Router struct {
	mux  *mux.Router
	repo storage.Repository
}

// New creates a new router
func New(repo storage.Repository) *Router {
	return &Router{
		mux:  mux.NewRouter(),
		repo: repo,
	}
}

// Handler returns the HTTP handler for the router
func (r *Router) Handler() http.Handler {
	return r.mux
}

// RegisterRoutes registers all application routes
func (r *Router) RegisterRoutes(routes []Route) {
	for _, route := range routes {
		r.registerRoute(route)
	}
}

// registerRoute registers a single route
func (r *Router) registerRoute(route Route) {
	var handler http.Handler = route.Handler

	// Apply middleware in reverse order (last middleware is executed first)
	for i := len(route.Middleware) - 1; i >= 0; i-- {
		handler = route.Middleware[i](handler)
	}

	r.mux.
		Methods(route.Methods...).
		Path(route.Path).
		Name(route.Name).
		Handler(handler)
}

// Group creates a subrouter for a group of routes with common path prefix and middleware
func (r *Router) Group(pathPrefix string, middleware ...mux.MiddlewareFunc) *mux.Router {
	subrouter := r.mux.PathPrefix(pathPrefix).Subrouter()
	subrouter.Use(middleware...)
	return subrouter
}

// AuthMiddleware returns the authentication middleware
func (r *Router) AuthMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Get the session cookie
			cookie, err := req.Cookie("session_token")
			if err != nil {
				// No session cookie, redirect to login
				http.Redirect(w, req, "/login", http.StatusSeeOther)
				return
			}

			// Get the session token
			sessionToken := cookie.Value

			// Get the session from the database
			ctx := req.Context()
			session, err := r.repo.GetSessionByToken(ctx, sessionToken)
			if err != nil {
				// Invalid session, redirect to login
				http.Redirect(w, req, "/login", http.StatusSeeOther)
				return
			}

			// Check if the session has expired
			if session.ExpiresAt.Before(middleware.GetCurrentTime()) {
				// Session expired, redirect to login
				http.Redirect(w, req, "/login", http.StatusSeeOther)
				return
			}

			// Get the user from the session
			user, err := r.repo.GetUserByID(ctx, session.UserID)
			if err != nil {
				// User not found, redirect to login
				http.Redirect(w, req, "/login", http.StatusSeeOther)
				return
			}

			// Update the user's last activity time
			user.LastActivity = middleware.GetCurrentTime()
			if err := r.repo.UpdateUser(ctx, user); err != nil {
				// Continue anyway, this is not critical
			}

			// Update session activity
			if err := r.repo.UpdateSessionActivity(ctx, session.ID); err != nil {
				// Continue anyway, this is not critical
			}

			// Add the user and session to the request context
			ctx = middleware.SetUserInContext(ctx, user)
			ctx = middleware.SetSessionInContext(ctx, session)

			// Call the next handler with the updated context
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}
