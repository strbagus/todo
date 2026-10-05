package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
	"todo/internal/models"
	"todo/internal/repository"
)

type ItemHandler struct {
	repo *repository.ItemRepo
	tmpl *template.Template
}

func NewItemHandler(repo *repository.ItemRepo, tmpl *template.Template) *ItemHandler {
	return &ItemHandler{repo: repo, tmpl: tmpl}
}

// CreateItem parses form values, validates, and persists a new item.
func (h *ItemHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`<div id="form-error-msg" style="color: #f87171; font-size: 0.85rem; margin-bottom: 0.75rem;">Title is required</div>`))
		return
	}

	startAtStr := r.FormValue("start_at")
	startAt, err := parseFormTime(startAtStr)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`<div id="form-error-msg" style="color: #f87171; font-size: 0.85rem; margin-bottom: 0.75rem;">Invalid start/due date format</div>`))
		return
	}

	itemType := models.ItemType(r.FormValue("type"))
	priority := models.Priority(r.FormValue("priority"))
	if priority == "" {
		priority = models.PriorityMedium
	}

	item := models.Item{
		Title:         title,
		Description:   strings.TrimSpace(r.FormValue("description")),
		Type:          itemType,
		StartAt:       startAt,
		ContextTag:    strings.TrimSpace(r.FormValue("context_tag")),
		LocationOrURL: strings.TrimSpace(r.FormValue("location_or_url")),
		Priority:      priority,
	}

	endAtStr := strings.TrimSpace(r.FormValue("end_at"))
	if endAtStr != "" {
		endAt, err := parseFormTime(endAtStr)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`<div id="form-error-msg" style="color: #f87171; font-size: 0.85rem; margin-bottom: 0.75rem;">Invalid end date format</div>`))
			return
		}
		item.EndAt = &endAt
	}

	if err := item.Validate(); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(fmt.Sprintf(`<div id="form-error-msg" style="color: #f87171; font-size: 0.85rem; margin-bottom: 0.75rem;">%s</div>`, template.HTMLEscapeString(err.Error()))))
		return
	}

	if err := h.repo.Create(r.Context(), &item); err != nil {
		http.Error(w, "Failed to create item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if caller expects timeline replacement
	target := r.Header.Get("HX-Target")
	if target == "timeline-container" {
		items, err := h.repo.List(r.Context(), "", "")
		if err != nil {
			http.Error(w, "Failed to fetch items: "+err.Error(), http.StatusInternalServerError)
			return
		}
		data := DashboardData{
			TodayDate: time.Now(),
			Groups:    models.GroupItems(items, time.Now()),
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = h.tmpl.ExecuteTemplate(w, "timeline_content.html", data)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.ExecuteTemplate(w, "item_card.html", item)
}

// ToggleItem toggles the completion status of an item.
func (h *ItemHandler) ToggleItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	updated, err := h.repo.ToggleCompleted(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to toggle item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.ExecuteTemplate(w, "item_card.html", updated)
}

// GetEditModal renders the edit modal partial populated with item details.
func (h *ItemHandler) GetEditModal(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Item not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.ExecuteTemplate(w, "edit_modal.html", item)
}

// UpdateItem updates an existing item.
func (h *ItemHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`<div id="edit-form-error-msg" style="color: #f87171; font-size: 0.85rem; margin-bottom: 0.75rem;">Title is required</div>`))
		return
	}

	startAtStr := r.FormValue("start_at")
	startAt, err := parseFormTime(startAtStr)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`<div id="edit-form-error-msg" style="color: #f87171; font-size: 0.85rem; margin-bottom: 0.75rem;">Invalid start/due date format</div>`))
		return
	}

	itemType := models.ItemType(r.FormValue("type"))
	priority := models.Priority(r.FormValue("priority"))
	if priority == "" {
		priority = models.PriorityMedium
	}

	item := models.Item{
		ID:            id,
		Title:         title,
		Description:   strings.TrimSpace(r.FormValue("description")),
		Type:          itemType,
		StartAt:       startAt,
		ContextTag:    strings.TrimSpace(r.FormValue("context_tag")),
		LocationOrURL: strings.TrimSpace(r.FormValue("location_or_url")),
		Priority:      priority,
	}

	endAtStr := strings.TrimSpace(r.FormValue("end_at"))
	if endAtStr != "" {
		endAt, err := parseFormTime(endAtStr)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`<div id="edit-form-error-msg" style="color: #f87171; font-size: 0.85rem; margin-bottom: 0.75rem;">Invalid end date format</div>`))
			return
		}
		item.EndAt = &endAt
	}

	if err := item.Validate(); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(fmt.Sprintf(`<div id="edit-form-error-msg" style="color: #f87171; font-size: 0.85rem; margin-bottom: 0.75rem;">%s</div>`, template.HTMLEscapeString(err.Error()))))
		return
	}

	if err := h.repo.Update(r.Context(), &item); err != nil {
		http.Error(w, "Failed to update item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch updated item to get latest state
	fresh, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		fresh = &item
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.ExecuteTemplate(w, "item_card.html", fresh)
}

// DeleteItem deletes an item from SQLite.
func (h *ItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// FilterItems handles filtering and search, returning timeline_content.html partial.
func (h *ItemHandler) FilterItems(w http.ResponseWriter, r *http.Request) {
	filterType := models.ItemType(r.URL.Query().Get("type"))
	search := r.URL.Query().Get("search")

	items, err := h.repo.List(r.Context(), filterType, search)
	if err != nil {
		http.Error(w, "Failed to filter items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := DashboardData{
		TodayDate: time.Now(),
		Groups:    models.GroupItems(items, time.Now()),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.ExecuteTemplate(w, "timeline_content.html", data)
}

func parseFormTime(s string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time: %s", s)
}
