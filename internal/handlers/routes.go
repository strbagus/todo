package handlers

import (
	"html/template"
	"io/fs"
	"net/http"
	"todo/internal/repository"
	"todo/web"
)

// RegisterRoutes registers all dashboard and HTMX routes using Go 1.22+ ServeMux patterns.
func RegisterRoutes(repo *repository.ItemRepo, tmpl *template.Template) *http.ServeMux {
	mux := http.NewServeMux()

	viewsHandler := NewViewsHandler(repo, tmpl)
	itemHandler := NewItemHandler(repo, tmpl)

	// Dashboard page
	mux.HandleFunc("GET /", viewsHandler.ServeDashboard)

	// HTMX item interactions
	mux.HandleFunc("GET /items/filter", itemHandler.FilterItems)
	mux.HandleFunc("POST /items", itemHandler.CreateItem)
	mux.HandleFunc("POST /items/{id}/toggle", itemHandler.ToggleItem)
	mux.HandleFunc("GET /items/{id}/edit", itemHandler.GetEditModal)
	mux.HandleFunc("PUT /items/{id}", itemHandler.UpdateItem)
	mux.HandleFunc("DELETE /items/{id}", itemHandler.DeleteItem)

	// Static file serving from embedded filesystem
	staticSub, err := fs.Sub(web.FS, "static")
	if err == nil {
		mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))
	}

	return mux
}
