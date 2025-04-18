package handlers

import (
	"log"
	"net/http"

	"github.com/korjavin/deadmanswitch/internal/web/templates"
)

// IndexHandler handles index-related requests
type IndexHandler struct{}

// NewIndexHandler creates a new IndexHandler
func NewIndexHandler() *IndexHandler {
	return &IndexHandler{}
}

// HandleIndex handles the index page
func (h *IndexHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	// Check if the user is already logged in
	cookie, err := r.Cookie("session_token")
	isAuthenticated := err == nil && cookie.Value != ""

	data := templates.TemplateData{
		Title:           "Dead Man's Switch",
		ActivePage:      "home",
		IsAuthenticated: isAuthenticated,
		Data:            make(map[string]interface{}),
	}

	if err := templates.RenderTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		log.Printf("Error rendering index template: %v", err)
	}
}
