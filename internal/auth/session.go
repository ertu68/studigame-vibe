// Package auth implements cookie sessions, CSRF tokens, and password
// registration/login/reset — all backed by plain SQL, no external session lib.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"time"
)

const (
	sessionCookieName = "session"
	sessionTTL        = 30 * 24 * time.Hour
)

var ErrSessionNotFound = errors.New("session not found or expired")

type Session struct {
	UserID int64
}

type SessionStore struct {
	db *sql.DB
}

func NewSessionStore(db *sql.DB) *SessionStore {
	return &SessionStore{db: db}
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Create starts a new session for userID and returns the raw token to set as a cookie.
func (s *SessionStore) Create(userID int64) (rawToken string, err error) {
	rawToken, err = randomToken()
	if err != nil {
		return "", err
	}
	_, err = s.db.Exec(
		`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES (?, ?, ?)`,
		hashToken(rawToken), userID, time.Now().UTC().Add(sessionTTL),
	)
	if err != nil {
		return "", err
	}
	return rawToken, nil
}

func (s *SessionStore) Get(rawToken string) (*Session, error) {
	if rawToken == "" {
		return nil, ErrSessionNotFound
	}
	var sess Session
	var expiresAt time.Time
	err := s.db.QueryRow(
		`SELECT user_id, expires_at FROM sessions WHERE token_hash = ?`,
		hashToken(rawToken),
	).Scan(&sess.UserID, &expiresAt)
	if err == sql.ErrNoRows {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	if time.Now().After(expiresAt) {
		return nil, ErrSessionNotFound
	}
	return &sess, nil
}

func (s *SessionStore) Delete(rawToken string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token_hash = ?`, hashToken(rawToken))
	return err
}

// secure should be true in production (TLS terminated at Caddy) and false for
// plain-http local development, otherwise browsers drop the cookie.
func SetSessionCookie(w http.ResponseWriter, rawToken string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    rawToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
}

func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func ReadSessionCookie(r *http.Request) string {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return ""
	}
	return c.Value
}
