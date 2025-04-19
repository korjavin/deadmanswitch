package activity

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
)

func TestGitHubProvider_IsConfigured(t *testing.T) {
	provider := NewGitHubProvider()

	// Test with configured user
	userWithGitHub := &models.User{
		GitHubUsername: "testuser",
	}
	if !provider.IsConfigured(userWithGitHub) {
		t.Error("IsConfigured should return true for user with GitHub username")
	}

	// Test with unconfigured user
	userWithoutGitHub := &models.User{
		GitHubUsername: "",
	}
	if provider.IsConfigured(userWithoutGitHub) {
		t.Error("IsConfigured should return false for user without GitHub username")
	}
}

// mockGitHubServer creates a test server that returns GitHub events
func mockGitHubServer(_ *testing.T, events []GitHubEvent) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleMockGitHubRequests(w, r, events)
	}))
}

func handleMockGitHubRequests(w http.ResponseWriter, r *http.Request, events []GitHubEvent) {
	// Return the events as JSON regardless of the path
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(events); err != nil {
		log.Printf("Error encoding events: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func TestGitHubProvider_CheckActivity(t *testing.T) {
	// Create mock events
	events := []GitHubEvent{
		{
			Type:      "PushEvent",
			CreatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
		},
	}

	// Create a mock server
	server := mockGitHubServer(t, events)
	defer server.Close()

	// Create a custom transport that redirects all requests to our test server
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	// Create a provider that uses the test server
	provider := NewGitHubProvider()
	// Replace the client with one that uses our test server
	provider.client = &http.Client{
		Transport: transport,
	}

	// Create a user with GitHub username
	user := &models.User{
		GitHubUsername: "korjavin",
	}

	// Create a context with the test server URL
	ctx := context.WithValue(context.Background(), "github_api_url", server.URL)

	// Test with activity after the since time
	since := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC) // Before the event
	active, err := provider.CheckActivity(ctx, user, since)
	if err != nil {
		t.Errorf("CheckActivity returned an error: %v", err)
	}
	if !active {
		t.Error("CheckActivity should return true for activity after the since time")
	}

	// Test with activity before the since time
	since = time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC) // After the event
	active, err = provider.CheckActivity(ctx, user, since)
	if err != nil {
		t.Errorf("CheckActivity returned an error: %v", err)
	}
	if active {
		t.Error("CheckActivity should return false for activity before the since time")
	}
}

func TestGitHubProvider_LastActivityTime(t *testing.T) {
	// Create mock events with different timestamps
	events := []GitHubEvent{
		{
			Type:      "PushEvent",
			CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			Type:      "IssueCommentEvent",
			CreatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC), // Latest event
		},
		{
			Type:      "PullRequestEvent",
			CreatedAt: time.Date(2023, 1, 1, 6, 0, 0, 0, time.UTC),
		},
	}

	// Create a mock server
	server := mockGitHubServer(t, events)
	defer server.Close()

	// Create a custom transport that redirects all requests to our test server
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	// Create a provider that uses the test server
	provider := NewGitHubProvider()
	// Replace the client with one that uses our test server
	provider.client = &http.Client{
		Transport: transport,
	}

	// Test with a configured user
	user := &models.User{
		GitHubUsername: "korjavin",
	}
	// Create a context with the test server URL
	ctx := context.WithValue(context.Background(), "github_api_url", server.URL)

	lastActivity, err := provider.LastActivityTime(ctx, user)
	if err != nil {
		t.Errorf("LastActivityTime returned an error: %v", err)
	}

	// The latest event is the IssueCommentEvent on 2023-01-02
	expected := time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC)
	if !lastActivity.Equal(expected) {
		t.Errorf("LastActivityTime returned %v, expected %v", lastActivity, expected)
	}

	// Test with an unconfigured user
	userWithoutGitHub := &models.User{
		GitHubUsername: "",
	}
	_, err = provider.LastActivityTime(context.Background(), userWithoutGitHub)
	if err == nil {
		t.Error("LastActivityTime should return an error for user without GitHub username")
	}
}

// Helper function to parse URLs in tests
func mustParseURL(t *testing.T, rawURL string) *url.URL {
	url, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("Failed to parse URL %s: %v", rawURL, err)
	}
	return url
}
