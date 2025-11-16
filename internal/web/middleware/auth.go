package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
	"github.com/korjavin/deadmanswitch/internal/storage"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserContextKey is the context key for storing user data
	UserContextKey contextKey = "user"
	// SessionContextKey is the context key for storing session data
	SessionContextKey contextKey = "session"
	// RecipientIDContextKey is the context key for storing recipient ID
	RecipientIDContextKey contextKey = "recipientID"
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
			if err != nil {
				// Invalid session, redirect to login
				log.Printf("Invalid session: %v", err)
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
			if err != nil {
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
			ctx = context.WithValue(ctx, UserContextKey, user)
			ctx = context.WithValue(ctx, SessionContextKey, session)

			// Call the next handler with the updated context
			next(w, r.WithContext(ctx))
		}
	}
}

// GetUserFromContext gets the user from the request context
func GetUserFromContext(r *http.Request) (*models.User, bool) {
	user, ok := r.Context().Value(UserContextKey).(*models.User)
	return user, ok
}
