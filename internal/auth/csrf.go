package auth

import (
	"crypto/subtle"
	"net/http"
)

// CSRF uses the double-submit-cookie pattern: a random token is set as an
// HttpOnly cookie and mirrored into a hidden form field by the handler that
// renders the form. On POST, the two must match. Works the same for
// anonymous (login/register) and authenticated forms, no session storage needed.
const (
	csrfFieldName  = "csrf_token"
	csrfCookieName = "csrf_token"
)

const CSRFFieldName = csrfFieldName

// EnsureCSRFToken returns the current CSRF token for this browser, setting a
// fresh cookie on first visit.
func EnsureCSRFToken(w http.ResponseWriter, r *http.Request, secure bool) string {
	if c, err := r.Cookie(csrfCookieName); err == nil && c.Value != "" {
		return c.Value
	}
	token, err := randomToken()
	if err != nil {
		panic(err) // crypto/rand failure is unrecoverable
	}
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
	return token
}

// ValidCSRFRequest checks the submitted form field against the cookie.
func ValidCSRFRequest(r *http.Request) bool {
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	submitted := r.FormValue(csrfFieldName)
	if submitted == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(submitted)) == 1
}
