package handlers

import (
	"context"

	"github.com/korjavin/deadmanswitch/internal/models"
)

// setupTestContext sets up a mock context for testing
func setupTestContext(user *models.User) context.Context {
	return context.WithValue(context.Background(), "user", user)
}
