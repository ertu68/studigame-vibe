// Package handlers wires HTTP requests to the auth/db logic and renders templates.
package handlers

import (
	"database/sql"
	"log/slog"
	"math/rand"
	"net/http"
	"time"

	"pennergame/internal/auth"
	"pennergame/internal/config"
	"pennergame/internal/game"
)

type Deps struct {
	DB       *sql.DB
	Sessions *auth.SessionStore
	Mailer   auth.Mailer
	Config   config.Config

	// Now and Roll are injected so tests can pin time and luck.
	Now  func() time.Time
	Roll func() float64
}

func (d *Deps) now() time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now()
}

func (d *Deps) roll() float64 {
	if d.Roll != nil {
		return d.Roll()
	}
	return rand.Float64()
}

// addCharacterStatus fills the values the status bar renders.
func (d *Deps) addCharacterStatus(data map[string]any, char Character) {
	data["Money"] = formatCents(char.MoneyCents)
	data["MoneyLevel"] = char.ContainerLevel
	data["MoneyFillPercent"] = fillPercent(char.MoneyCents, game.ContainerCapacity(char.ContainerLevel))
	data["Mood"] = char.Mood
}

type user struct {
	ID   int64
	Name string
	City string
}

// currentUser loads the logged-in user for a request, or nil if anonymous.
func (d *Deps) currentUser(r *http.Request) *user {
	rawToken := auth.ReadSessionCookie(r)
	sess, err := d.Sessions.Get(rawToken)
	if err != nil {
		if err != auth.ErrSessionNotFound {
			slog.Warn("Session retrieval error", "error", err)
		}
		return nil
	}
	var u user
	err = d.DB.QueryRow(`SELECT id, name, city FROM users WHERE id = ?`, sess.UserID).Scan(&u.ID, &u.Name, &u.City)
	if err != nil {
		if err != sql.ErrNoRows {
			slog.Error("Failed to get user", "error", err, "userID", sess.UserID)
		}
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
			slog.Info("Redirecting anonymous user to login", "path", r.URL.Path)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		slog.Info("Authenticated access", "userID", u.ID, "path", r.URL.Path)
		next(u, w, r)
	}
}

// RequireCSRF rejects POSTs whose csrf_token field doesn't match the cookie.
func RequireCSRF(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.ValidCSRFRequest(r) {
			slog.Warn("CSRF validation failed", "method", r.Method, "path", r.URL.Path)
			http.Error(w, "ungueltiges CSRF-Token", http.StatusForbidden)
			return
		}
		slog.Debug("CSRF validation passed", "method", r.Method, "path", r.URL.Path)
		next(w, r)
	}
}
