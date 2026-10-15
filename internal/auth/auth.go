package auth

import (
	"crypto/sha256"
	"crypto/subtle"
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

// Validate checks whether the provided username and password match configuration (constant time).
func (c Config) Validate(user, pass string) bool {
	if !c.IsEnabled() {
		return true
	}
	userMatch := subtle.ConstantTimeCompare([]byte(user), []byte(c.Username)) == 1
	passMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(c.Password)) == 1
	return userMatch && passMatch
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

// NewSessionCookie returns a persistent cookie compatible with cross-origin extension iframes.
// SameSite=None and Secure=true allow the cookie inside extension iframes over HTTPS (production),
// while falling back to SameSite=Lax for non-HTTPS (local development HTTP).
func (c Config) NewSessionCookie(isSecure bool) *http.Cookie {
	sameSite := http.SameSiteNoneMode
	if !isSecure {
		sameSite = http.SameSiteLaxMode
	}

	return &http.Cookie{
		Name:     c.CookieName(),
		Value:    c.Token(),
		Path:     "/",
		MaxAge:   315360000, // 10 years in seconds
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: sameSite,
	}
}

// ClearSessionCookie returns a cookie configured to delete the session.
func (c Config) ClearSessionCookie(isSecure bool) *http.Cookie {
	sameSite := http.SameSiteNoneMode
	if !isSecure {
		sameSite = http.SameSiteLaxMode
	}

	return &http.Cookie{
		Name:     c.CookieName(),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: sameSite,
	}
}
