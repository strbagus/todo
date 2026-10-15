package handlers_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"todo/internal/database"
	"todo/internal/handlers"
	"todo/internal/models"
	"todo/internal/repository"
	"todo/internal/templates"
)

func TestItemHandler_Endpoints(t *testing.T) {
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

	mux := handlers.RegisterRoutes(repo, tmpl)

	// 1. Validation failure returns 422
	form := url.Values{}
	form.Set("title", "") // empty title
	req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for empty title, got %d", rec.Code)
	}

	// 2. Valid creation returns 200/201 with item card
	now := time.Now().Format("2006-01-02T15:04")
	form.Set("title", "Finish Thesis Proposal")
	form.Set("type", string(models.TypeTask))
	form.Set("start_at", now)
	form.Set("priority", string(models.PriorityHigh))
	req = httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
		t.Fatalf("expected 200/201 for valid creation, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Finish Thesis Proposal") {
		t.Errorf("response does not contain created item title")
	}

	// 3. Toggle completed
	items, err := repo.List(context.Background(), "", "")
	if err != nil || len(items) == 0 {
		t.Fatalf("expected at least 1 item in db, err: %v", err)
	}
	targetID := items[0].ID
	toggleURL := fmt.Sprintf("/items/%d/toggle", targetID)
	req = httptest.NewRequest(http.MethodPost, toggleURL, nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on toggle, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "completed") {
		t.Errorf("expected card to have completed class")
	}

	// 4. Edit modal fetch
	editURL := fmt.Sprintf("/items/%d/edit", targetID)
	req = httptest.NewRequest(http.MethodGet, editURL, nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on edit modal, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Edit Item") {
		t.Errorf("expected edit modal to contain 'Edit Item'")
	}

	// 5. Update item
	updateForm := url.Values{}
	updateForm.Set("title", "Finish Thesis Proposal (Final)")
	updateForm.Set("type", string(models.TypeTask))
	updateForm.Set("start_at", now)
	updateForm.Set("priority", string(models.PriorityHigh))
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/items/%d", targetID), strings.NewReader(updateForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Finish Thesis Proposal (Final)") {
		t.Errorf("expected updated title in response card")
	}

	// 6. Filter endpoint
	req = httptest.NewRequest(http.MethodGet, "/items/filter?type=task", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on filter, got %d", rec.Code)
	}

	// 7. Delete item
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/items/%d", targetID), nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 on delete, got %d", rec.Code)
	}
}
