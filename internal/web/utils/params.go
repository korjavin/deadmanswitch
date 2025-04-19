// Package utils provides utility functions for the web interface
// of the Dead Man's Switch application, including parameter parsing,
// validation, and helper functions for common operations.
package utils

import (
	"net/http"
	"strings"
)

// GetURLParam extracts a URL parameter from the request path
// Example: for path /users/{id}, calling GetURLParam(r, "id") will return the value
func GetURLParam(r *http.Request, param string) string {
	// Get the full URL path
	path := r.URL.Path

	// Split the URL into segments
	segments := strings.Split(path, "/")

	// Find the parameter position by matching {param}
	for i, segment := range segments {
		if segment == "{"+param+"}" {
			// If we find the parameter pattern and there's a next segment, return it
			if i+1 < len(segments) {
				return segments[i+1]
			}
			return ""
		}
	}

	return ""
}

// GetLastURLSegment returns the last segment of the URL path
// Useful for routes like /profile/passkeys/{id}
func GetLastURLSegment(r *http.Request) string {
	path := strings.TrimSuffix(r.URL.Path, "/")
	segments := strings.Split(path, "/")
	if len(segments) > 0 {
		return segments[len(segments)-1]
	}
	return ""
}
