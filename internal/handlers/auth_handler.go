package handlers

import (
	"html/template"
	"net/http"
	"strings"
	"todo/internal/auth"
)

type AuthHandler struct {
	cfg  auth.Config
	tmpl *template.Template
}

func NewAuthHandler(cfg auth.Config, tmpl *template.Template) *AuthHandler {
	return &AuthHandler{cfg: cfg, tmpl: tmpl}
}

// ServeLogin renders the login form page.
func (h *AuthHandler) ServeLogin(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.IsEnabled() {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.ExecuteTemplate(w, "login.html", nil)
}

// HandleLogin processes credentials submission and issues a persistent cookie.
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user := strings.TrimSpace(r.FormValue("username"))
	pass := strings.TrimSpace(r.FormValue("password"))

	if !h.cfg.Validate(user, pass) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = h.tmpl.ExecuteTemplate(w, "login.html", map[string]string{
			"Error": "Invalid username or password",
		})
		return
	}

	http.SetCookie(w, h.cfg.NewSessionCookie())
	http.Redirect(w, r, "/", http.StatusFound)
}

// HandleLogout clears the session cookie and redirects to /login.
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, h.cfg.ClearSessionCookie())
	http.Redirect(w, r, "/login", http.StatusFound)
}

// Middleware protects endpoints by requiring a valid persistent cookie or Basic Auth header.
func (h *AuthHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.cfg.IsEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		path := r.URL.Path
		// Public paths: /login, /logout, /static/*
		if path == "/login" || path == "/logout" || strings.HasPrefix(path, "/static/") {
			next.ServeHTTP(w, r)
			return
		}

		// 1. Check HTTP Basic Auth header
		user, pass, ok := r.BasicAuth()
		if ok && h.cfg.Validate(user, pass) {
			next.ServeHTTP(w, r)
			return
		}

		// 2. Check persistent session cookie
		cookie, err := r.Cookie(h.cfg.CookieName())
		if err == nil && cookie.Value == h.cfg.Token() {
			next.ServeHTTP(w, r)
			return
		}

		// Unauthenticated handling
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Redirect", "/login")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		http.Redirect(w, r, "/login", http.StatusFound)
	})
}
