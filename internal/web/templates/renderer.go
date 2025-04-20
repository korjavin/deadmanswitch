package templates

import (
	"log"
	"net/http"
)

// TemplateRenderer is a helper for rendering templates
type TemplateRenderer struct {
	// Add any configuration options here if needed
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer() *TemplateRenderer {
	return &TemplateRenderer{}
}

// Render renders a template with the given data
func (tr *TemplateRenderer) Render(w http.ResponseWriter, templateName string, data map[string]interface{}) error {
	// Create template data
	templateData := TemplateData{
		Title:           templateName, // Default title
		IsAuthenticated: true,         // Assume authenticated for now
		Data:            data,
	}

	// If user is provided in data, add it to the template data
	if user, ok := data["User"]; ok {
		templateData.User = map[string]interface{}{
			"User": user,
		}
	}

	// Render the template
	if err := RenderTemplate(w, templateName, templateData); err != nil {
		log.Printf("Error rendering template %s: %v", templateName, err)
		return err
	}

	return nil
}
