package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
)

// GitHubEvent represents a GitHub event from the API
type GitHubEvent struct {
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// GitHubProvider implements the Provider interface for GitHub
type GitHubProvider struct {
	client *http.Client
}

// NewGitHubProvider creates a new GitHub activity provider
func NewGitHubProvider() *GitHubProvider {
	return &GitHubProvider{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Name returns the name of the provider
func (p *GitHubProvider) Name() string {
	return "GitHub"
}

// IsConfigured returns true if the user has a GitHub username configured
func (p *GitHubProvider) IsConfigured(user *models.User) bool {
	return user.GitHubUsername != ""
}

// CheckActivity checks if the user has been active on GitHub since the given time
func (p *GitHubProvider) CheckActivity(ctx context.Context, user *models.User, since time.Time) (bool, error) {
	if !p.IsConfigured(user) {
		return false, fmt.Errorf("github username not configured for user")
	}

	lastActivity, err := p.LastActivityTime(ctx, user)
	if err != nil {
		return false, err
	}

	return lastActivity.After(since), nil
}

// LastActivityTime returns the time of the user's last activity on GitHub
func (p *GitHubProvider) LastActivityTime(ctx context.Context, user *models.User) (time.Time, error) {
	if !p.IsConfigured(user) {
		return time.Time{}, fmt.Errorf("github username not configured for user")
	}

	// GitHub API URL for public events
	url := fmt.Sprintf("https://api.github.com/users/%s/events/public", user.GitHubUsername)

	// For testing purposes, allow overriding the URL
	if testURL := ctx.Value("github_api_url"); testURL != nil {
		if testURLStr, ok := testURL.(string); ok && testURLStr != "" {
			url = fmt.Sprintf("%s/users/%s/events/public", testURLStr, user.GitHubUsername)
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent header as required by GitHub API
	req.Header.Set("User-Agent", "DeadMansSwitch-App")

	resp, err := p.client.Do(req)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return time.Time{}, fmt.Errorf("github API returned status code %d", resp.StatusCode)
	}

	var events []GitHubEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return time.Time{}, fmt.Errorf("failed to decode response: %w", err)
	}

	// Find the most recent event
	var latestTime time.Time
	for _, event := range events {
		if event.CreatedAt.After(latestTime) {
			latestTime = event.CreatedAt
		}
	}

	return latestTime, nil
}
