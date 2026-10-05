package templates

import (
	"fmt"
	"html/template"
	"io"
	"time"
	"todo/web"
)

// FuncMap defines helper functions available across all HTML templates.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("Monday, January 2, 2006")
		},
		"formatTime": func(t time.Time) string {
			return t.Format("15:04")
		},
		"formatDateTimeInput": func(t time.Time) string {
			return t.Format("2006-01-02T15:04")
		},
		"formatRelative": func(t time.Time) string {
			now := time.Now()
			diff := t.Sub(now)
			if diff < 0 {
				hours := int(-diff.Hours())
				if hours < 24 {
					return fmt.Sprintf("Overdue by %dh", hours+1)
				}
				days := hours / 24
				return fmt.Sprintf("Overdue by %dd", days)
			}
			hours := int(diff.Hours())
			if hours < 24 {
				return fmt.Sprintf("in %dh", hours+1)
			}
			days := hours / 24
			return fmt.Sprintf("in %dd", days)
		},
	}
}

// ParseTemplates parses all embedded HTML templates.
func ParseTemplates() (*template.Template, error) {
	tmpl, err := template.New("").Funcs(FuncMap()).ParseFS(
		web.FS,
		"templates/*.html",
		"templates/components/*.html",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}
	return tmpl, nil
}

// Render executes a specific template by name.
func Render(w io.Writer, tmpl *template.Template, name string, data any) error {
	return tmpl.ExecuteTemplate(w, name, data)
}
