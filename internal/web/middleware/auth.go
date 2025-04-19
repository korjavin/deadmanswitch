// Package middleware provides HTTP middleware for the web server
// of the Dead Man's Switch application. It includes middleware for
// authentication, logging, and other cross-cutting concerns.
package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
)

// Define proper context key types
type contextKey string

const (
	userContextKey    contextKey = "user"
	sessionContextKey contextKey = "session"
)

// Auth is a middleware that checks if the user is authenticated
func Auth(repo storage.Repository) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get the session cookie
			cookie, err := r.Cookie("session_token")
			if err != nil {
				// No session cookie, redirect to login
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Get the session token
			sessionToken := cookie.Value

			// Get the session from the database
			ctx := r.Context()
			session, err := repo.GetSessionByToken(ctx, sessionToken)
			if err != nil || session == nil {
				// Invalid or missing session, redirect to login
				log.Printf("Invalid or missing session: %v", err)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Check if the session has expired
			if session.ExpiresAt.Before(time.Now()) {
				// Session expired, redirect to login
				log.Printf("Session expired")
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Get the user from the session
			user, err := repo.GetUserByID(ctx, session.UserID)
			if err != nil || user == nil {
				// User not found, redirect to login
				log.Printf("User not found: %v", err)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Update the user's last activity time
			user.LastActivity = time.Now()
			if err := repo.UpdateUser(ctx, user); err != nil {
				log.Printf("Error updating user last activity: %v", err)
				// Continue anyway, this is not critical
			}

			// Update session activity
			if err := repo.UpdateSessionActivity(ctx, session.ID); err != nil {
				log.Printf("Error updating session activity: %v", err)
				// Continue anyway, this is not critical
			}

			// Add the user and session to the request context
			log.Printf("Setting context for user ID: %d, session ID: %s", user.ID, session.ID)
			ctx = context.WithValue(ctx, userContextKey, user)
			ctx = context.WithValue(ctx, sessionContextKey, session)

			// Call the next handler with the updated context
			log.Printf("Passing request to next handler with updated context")
			next(w, r.WithContext(ctx))
		}
	}
}

// GetUserFromContext gets the user from the request context
func GetUserFromContext(r *http.Request) (*models.User, bool) {
	user, ok := r.Context().Value(userContextKey).(*models.User)
	return user, ok
}

// GetSessionFromContext gets the session from the request context
func GetSessionFromContext(r *http.Request) (*models.Session, bool) {
	session, ok := r.Context().Value(sessionContextKey).(*models.Session)
	return session, ok
}
