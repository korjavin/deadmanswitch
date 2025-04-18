package templates

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// TemplatePaths contains possible paths where templates might be located
var TemplatePaths = []string{
	"/app/web/templates",
	"./web/templates",
}

// TemplateData represents the data passed to templates
type TemplateData struct {
	Title           string
	ActivePage      string
	IsAuthenticated bool
	User            map[string]interface{}
	Data            map[string]interface{}
	Flash           map[string]string
}

// FuncMap returns a template.FuncMap with common template functions
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("Jan 2, 2006")
		},
		"formatDateTime": func(t time.Time) string {
			return t.Format("Jan 2, 2006 15:04")
		},
		"add": func(a, b int) int {
			return a + b
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}
}

// RenderTemplate renders a template with the given data
func RenderTemplate(w http.ResponseWriter, templateName string, data TemplateData) error {
	// Try to find templates in multiple locations
	var tmpl *template.Template
	var err error
	var templateErr error

	// Try each template path
	for _, basePath := range TemplatePaths {
		// Check if the directory exists
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			continue
		}

		layoutPath := filepath.Join(basePath, "layout.html")
		contentPath := filepath.Join(basePath, templateName)

		// Check if both files exist
		if _, err := os.Stat(layoutPath); os.IsNotExist(err) {
			continue
		}
		if _, err := os.Stat(contentPath); os.IsNotExist(err) {
			continue
		}

		// Parse the templates with the function map
		tmpl, err = template.New(filepath.Base(layoutPath)).Funcs(FuncMap()).ParseFiles(layoutPath, contentPath)
		if err == nil {
			break
		}
		if templateErr == nil {
			templateErr = err
		}
	}

	if tmpl == nil {
		log.Printf("Error parsing template %s: %v", templateName, templateErr)
		return fmt.Errorf("template error: %w", templateErr)
	}

	// Execute the template
	return tmpl.ExecuteTemplate(w, "layout", data)
}
