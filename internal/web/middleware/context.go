package middleware

import (
	"context"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
)

// Context keys
type contextKey string

const (
	userContextKey    contextKey = "user"
	sessionContextKey contextKey = "session"
)

// GetCurrentTime returns the current time
// This function is extracted to make testing easier
var GetCurrentTime = func() time.Time {
	return time.Now()
}

// SetUserInContext sets the user in the context
func SetUserInContext(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// GetUserFromContext gets the user from the context
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(userContextKey).(*models.User)
	return user, ok
}

// SetSessionInContext sets the session in the context
func SetSessionInContext(ctx context.Context, session *models.Session) context.Context {
	return context.WithValue(ctx, sessionContextKey, session)
}

// GetSessionFromContext gets the session from the context
func GetSessionFromContext(ctx context.Context) (*models.Session, bool) {
	session, ok := ctx.Value(sessionContextKey).(*models.Session)
	return session, ok
}
