package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"todo/internal/database"
	"todo/internal/handlers"
	"todo/internal/models"
	"todo/internal/repository"
	"todo/internal/templates"
)

func TestServeDashboard(t *testing.T) {
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
	repo := repository.NewItemRepo(db)
	tmpl, err := templates.ParseTemplates()
	if err != nil {
		t.Fatalf("ParseTemplates failed: %v", err)
	}

	// Seed item
	err = repo.Create(context.Background(), &models.Item{
		Title:    "Operating Systems Lab",
		Type:     models.TypeCollegeTask,
		StartAt:  time.Now().Add(2 * time.Hour),
		Priority: models.PriorityHigh,
	})
	if err != nil {
		t.Fatalf("Create item failed: %v", err)
	}

	handler := handlers.NewViewsHandler(repo, tmpl)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeDashboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Operating Systems Lab") {
		t.Errorf("expected dashboard body to contain 'Operating Systems Lab'")
	}
	if !strings.Contains(body, "Today") {
		t.Errorf("expected dashboard body to contain timeline group 'Today'")
	}
}
