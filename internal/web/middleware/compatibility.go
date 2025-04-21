package middleware

import (
	"net/http"

	"github.com/korjavin/deadmanswitch/internal/models"
)

// GetUserFromContext is kept for backward compatibility
// Use GetUserFromRequest instead for new code
func GetUserFromContext(r *http.Request) (*models.User, bool) {
	return GetUserFromRequest(r)
}
