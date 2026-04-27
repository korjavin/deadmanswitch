package templates

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/korjavin/deadmanswitch/internal/models"
)

// TemplatePaths contains possible paths where templates might be located
var TemplatePaths = []string{
	"/app/web/templates",
	"./web/templates",
}

// TemplateData represents the data passed to templates
type TemplateData struct {
	Title            string
	ActivePage       string
	IsAuthenticated  bool
	User             map[string]interface{}
	Data             map[string]interface{}
	Flash            map[string]string
	HeartbeatVariant string // "ok" | "warn" | "crit"; empty means layout falls back to "ok"
	HeartbeatLabel   string // short suffix shown in the topbar pill (e.g. "OK", "OVERDUE", "URGENT")
}

// SetHeartbeat fills the topbar HeartbeatVariant/Label from the user's
// LastActivity + ping cadence. Call this from every authenticated handler
// before RenderTemplate so the topbar pulse reflects real status instead
// of a hardcoded "OK".
func (d *TemplateData) SetHeartbeat(user *models.User) {
	if user == nil {
		return
	}
	d.HeartbeatVariant, d.HeartbeatLabel = HeartbeatStatus(user, time.Now())
}

// HeartbeatStatus collapses the user's ping schedule into the soft
// "ok" / "warn" / "crit" variant + matching uppercase label used by the
// topbar pulse. Mirrors the dashboard's status derivation.
func HeartbeatStatus(user *models.User, now time.Time) (variant, label string) {
	nextCheckIn := user.LastActivity.AddDate(0, 0, user.PingFrequency)
	deadline := user.LastActivity.AddDate(0, 0, user.PingDeadline)
	if !now.Before(deadline) {
		return "crit", "URGENT"
	}
	if !now.Before(nextCheckIn) {
		return "warn", "OVERDUE"
	}
	return "ok", "OK"
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
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
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
