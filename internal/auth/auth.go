package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

type Config struct {
	Username string
	Password string
}

// IsEnabled returns true if authentication credentials are configured.
func (c Config) IsEnabled() bool {
	return c.Username != "" && c.Password != ""
}

// Validate checks whether the provided username and password match configuration.
func (c Config) Validate(user, pass string) bool {
	if !c.IsEnabled() {
		return true
	}
	return user == c.Username && pass == c.Password
}

// Token generates a deterministic SHA-256 hash of the credentials.
func (c Config) Token() string {
	h := sha256.Sum256([]byte(c.Username + ":" + c.Password))
	return hex.EncodeToString(h[:])
}

// CookieName returns the cookie name used for session authentication.
func (c Config) CookieName() string {
	return "todo_auth_token"
}

// NewSessionCookie returns a persistent cookie with ~10 years lifespan (effectively no expiry).
func (c Config) NewSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     c.CookieName(),
		Value:    c.Token(),
		Path:     "/",
		MaxAge:   315360000, // 10 years in seconds
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// ClearSessionCookie returns a cookie configured to delete the session.
func (c Config) ClearSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     c.CookieName(),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}
