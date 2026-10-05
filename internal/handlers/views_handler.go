package handlers

import (
	"html/template"
	"net/http"
	"time"
	"todo/internal/models"
	"todo/internal/repository"
)

type ViewsHandler struct {
	repo *repository.ItemRepo
	tmpl *template.Template
}

func NewViewsHandler(repo *repository.ItemRepo, tmpl *template.Template) *ViewsHandler {
	return &ViewsHandler{repo: repo, tmpl: tmpl}
}

type DashboardData struct {
	TodayDate time.Time
	Groups    models.TimelineGroups
}

func (h *ViewsHandler) ServeDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	items, err := h.repo.List(ctx, "", "")
	if err != nil {
		http.Error(w, "Failed to load items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	groups := models.GroupItems(items, now)

	data := DashboardData{
		TodayDate: now,
		Groups:    groups,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		http.Error(w, "Failed to render dashboard: "+err.Error(), http.StatusInternalServerError)
	}
}
