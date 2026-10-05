package handlers

import (
	"html/template"
	"io/fs"
	"net/http"
	"todo/internal/auth"
	"todo/internal/repository"
	"todo/web"
)

// RegisterRoutes registers all routes without authentication (for dev or backward compatibility).
func RegisterRoutes(repo *repository.ItemRepo, tmpl *template.Template) http.Handler {
	return RegisterRoutesWithAuth(repo, tmpl, auth.Config{})
}

// RegisterRoutesWithAuth registers routes and wraps them in authentication middleware.
func RegisterRoutesWithAuth(repo *repository.ItemRepo, tmpl *template.Template, cfg auth.Config) http.Handler {
	mux := http.NewServeMux()

	viewsHandler := NewViewsHandler(repo, tmpl)
	itemHandler := NewItemHandler(repo, tmpl)
	authHandler := NewAuthHandler(cfg, tmpl)

	// Auth endpoints
	mux.HandleFunc("GET /login", authHandler.ServeLogin)
	mux.HandleFunc("POST /login", authHandler.HandleLogin)
	mux.HandleFunc("POST /logout", authHandler.HandleLogout)

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

	return authHandler.Middleware(mux)
}
