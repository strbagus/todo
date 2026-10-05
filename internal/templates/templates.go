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
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("invalid dict call: odd number of arguments")
			}
			m := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				m[key] = values[i+1]
			}
			return m, nil
		},
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
