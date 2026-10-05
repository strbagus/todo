package templates_test

import (
	"bytes"
	"testing"
	"todo/internal/templates"
)

func TestTemplateParsing(t *testing.T) {
	tmpl, err := templates.ParseTemplates()
	if err != nil {
		t.Fatalf("ParseTemplates failed: %v", err)
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "cheatsheet_modal.html", nil)
	if err != nil {
		t.Fatalf("ExecuteTemplate cheatsheet_modal failed: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Vim Shortcuts")) {
		t.Errorf("expected cheatsheet modal to contain 'Vim Shortcuts'")
	}
}
