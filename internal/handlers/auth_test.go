package handlers_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"todo/internal/auth"
	"todo/internal/database"
	"todo/internal/handlers"
	"todo/internal/repository"
	"todo/internal/templates"
)

func TestAuth_MiddlewareAndFlow(t *testing.T) {
	db, _ := database.InitDB(":memory:")
	_ = database.Migrate(db)
	repo := repository.NewItemRepo(db)
	tmpl, _ := templates.ParseTemplates()

	cfg := auth.Config{
		Username: "admin",
		Password: "secretpassword",
	}

	mux := handlers.RegisterRoutesWithAuth(repo, tmpl, cfg)

	// 1. Unauthenticated request to / should redirect to /login
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/login" {
		t.Fatalf("expected 302 redirect to /login, got status %d location %s", rec.Code, rec.Header().Get("Location"))
	}

	// 2. Unauthenticated HTMX request should receive HX-Redirect header
	req = httptest.NewRequest(http.MethodGet, "/items/filter", nil)
	req.Header.Set("HX-Request", "true")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Header().Get("HX-Redirect") != "/login" {
		t.Fatalf("expected HX-Redirect header to /login, got %s", rec.Header().Get("HX-Redirect"))
	}

	// 3. Static assets should remain accessible unauthenticated
	req = httptest.NewRequest(http.MethodGet, "/static/css/style.css", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected static files to be public, got status %d", rec.Code)
	}

	// 4. Invalid credentials POST to /login should fail with 401
	form := url.Values{}
	form.Set("username", "admin")
	form.Set("password", "wrongpass")
	req = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong credentials, got %d", rec.Code)
	}

	// 5. Valid credentials POST to /login should set persistent cookie and redirect to /
	form.Set("password", "secretpassword")
	req = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/" {
		t.Fatalf("expected 302 redirect to / on login, got status %d location %s", rec.Code, rec.Header().Get("Location"))
	}

	cookies := rec.Result().Cookies()
	var authCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == cfg.CookieName() {
			authCookie = c
			break
		}
	}
	if authCookie == nil {
		t.Fatalf("expected auth cookie to be set")
	}
	if authCookie.MaxAge < 365*24*3600 {
		t.Fatalf("expected cookie MaxAge to be long-lived (>= 1 year), got %d", authCookie.MaxAge)
	}

	// 6. Request with valid cookie should succeed
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(authCookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with auth cookie, got %d", rec.Code)
	}

	// 7. Request with HTTP Basic Auth header should succeed
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:secretpassword"))
	req.Header.Set("Authorization", authHeader)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with HTTP Basic Auth header, got %d", rec.Code)
	}
}
