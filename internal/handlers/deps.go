// Package handlers wires HTTP requests to the auth/db logic and renders templates.
package handlers

import (
	"database/sql"
	"net/http"

	"pennergame/internal/auth"
	"pennergame/internal/config"
)

type Deps struct {
	DB       *sql.DB
	Sessions *auth.SessionStore
	Mailer   auth.Mailer
	Config   config.Config
}

type user struct {
	ID   int64
	Name string
}

// currentUser loads the logged-in user for a request, or nil if anonymous.
func (d *Deps) currentUser(r *http.Request) *user {
	rawToken := auth.ReadSessionCookie(r)
	sess, err := d.Sessions.Get(rawToken)
	if err != nil {
		return nil
	}
	var u user
	err = d.DB.QueryRow(`SELECT id, name FROM users WHERE id = ?`, sess.UserID).Scan(&u.ID, &u.Name)
	if err != nil {
		return nil
	}
	return &u
}

// baseData seeds template data every page needs: site name, current user, CSRF token.
func (d *Deps) baseData(w http.ResponseWriter, r *http.Request) map[string]any {
	return map[string]any{
		"SiteName":  d.Config.SiteName,
		"User":      d.currentUser(r),
		"CSRFToken": auth.EnsureCSRFToken(w, r, d.Config.Production),
	}
}

// RequireAuth redirects anonymous visitors to /login before running the handler.
func (d *Deps) RequireAuth(next func(*user, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := d.currentUser(r)
		if u == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(u, w, r)
	}
}

// RequireCSRF rejects POSTs whose csrf_token field doesn't match the cookie.
func RequireCSRF(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.ValidCSRFRequest(r) {
			http.Error(w, "ungueltiges CSRF-Token", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
