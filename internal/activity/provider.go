package activity

import (
	"context"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
)

// Provider defines the interface for checking user activity on external platforms
type Provider interface {
	// Name returns the name of the activity provider
	Name() string

	// IsConfigured returns true if the provider is properly configured for the user
	IsConfigured(user *models.User) bool

	// CheckActivity checks if the user has been active on the platform since the given time
	// Returns true if activity was detected, false otherwise
	CheckActivity(ctx context.Context, user *models.User, since time.Time) (bool, error)

	// LastActivityTime returns the time of the user's last activity on the platform
	// Returns zero time if no activity was found or an error occurred
	LastActivityTime(ctx context.Context, user *models.User) (time.Time, error)
}

// Registry maintains a collection of activity providers
type Registry struct {
	providers []Provider
}

// NewRegistry creates a new activity provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make([]Provider, 0),
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(provider Provider) {
	r.providers = append(r.providers, provider)
}

// GetProviders returns all registered providers
func (r *Registry) GetProviders() []Provider {
	return r.providers
}

// GetConfiguredProviders returns all providers that are configured for the given user
func (r *Registry) GetConfiguredProviders(user *models.User) []Provider {
	configured := make([]Provider, 0)
	for _, provider := range r.providers {
		if provider.IsConfigured(user) {
			configured = append(configured, provider)
		}
	}
	return configured
}

// CheckAnyActivity checks if the user has been active on any platform since the given time
// Returns true if activity was detected on any platform, false otherwise
func (r *Registry) CheckAnyActivity(ctx context.Context, user *models.User, since time.Time) (bool, error) {
	for _, provider := range r.GetConfiguredProviders(user) {
		active, err := provider.CheckActivity(ctx, user, since)
		if err != nil {
			// Log the error but continue checking other providers
			continue
		}
		if active {
			return true, nil
		}
	}
	return false, nil
}

// GetLatestActivityTime returns the most recent activity time across all platforms
func (r *Registry) GetLatestActivityTime(ctx context.Context, user *models.User) time.Time {
	var latest time.Time
	for _, provider := range r.GetConfiguredProviders(user) {
		t, err := provider.LastActivityTime(ctx, user)
		if err != nil {
			// Log the error but continue checking other providers
			continue
		}
		if t.After(latest) {
			latest = t
		}
	}
	return latest
}
