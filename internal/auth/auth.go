package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var (
	ErrEmailTaken       = errors.New("email already registered")
	ErrInvalidLogin     = errors.New("invalid email or password")
	ErrInvalidResetLink = errors.New("invalid or expired reset link")
)

func Register(db *sql.DB, name, email, password string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	res, err := db.Exec(
		`INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)`,
		name, email, string(hash),
	)
	if err != nil {
		// D5: Check specifically for UNIQUE constraint violation
		if sqliteErr, ok := err.(*sqlite.Error); ok {
			if sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
				return 0, ErrEmailTaken
			}
		}
		// For all other errors, return the actual error
		return 0, err
	}
	return res.LastInsertId()
}

func Authenticate(db *sql.DB, email, password string) (int64, error) {
	var id int64
	var hash string
	err := db.QueryRow(`SELECT id, password_hash FROM users WHERE email = ?`, email).Scan(&id, &hash)
	if err == sql.ErrNoRows {
		return 0, ErrInvalidLogin
	}
	if err != nil {
		return 0, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return 0, ErrInvalidLogin
	}
	return id, nil
}

// RequestPasswordReset always returns nil on a well-formed email, whether or
// not an account exists, so the endpoint can't be used to enumerate users.
func RequestPasswordReset(db *sql.DB, mailer Mailer, siteDomain, email string, ttl time.Duration) error {
	var userID int64
	err := db.QueryRow(`SELECT id FROM users WHERE email = ?`, email).Scan(&userID)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	rawToken, err := randomToken()
	if err != nil {
		return err
	}
	_, err = db.Exec(
		`INSERT INTO password_resets (token_hash, user_id, expires_at) VALUES (?, ?, ?)`,
		hashToken(rawToken), userID, time.Now().Add(ttl),
	)
	if err != nil {
		return err
	}

	link := fmt.Sprintf("https://%s/reset-password/confirm?token=%s", siteDomain, rawToken)
	return mailer.Send(email, "Passwort zuruecksetzen", "Link: "+link)
}

func ResetPassword(db *sql.DB, token, newPassword string) error {
	var userID int64
	var expiresAt time.Time
	tokenHash := hashToken(token)
	err := db.QueryRow(
		`SELECT user_id, expires_at FROM password_resets WHERE token_hash = ?`, tokenHash,
	).Scan(&userID, &expiresAt)
	if err == sql.ErrNoRows {
		return ErrInvalidResetLink
	}
	if err != nil {
		return err
	}
	if time.Now().After(expiresAt) {
		return ErrInvalidResetLink
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := db.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, string(hash), userID); err != nil {
		return err
	}
	// D4: Invalidate all sessions for this user
	_, err = db.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)
	if err != nil {
		return err
	}
	_, err = db.Exec(`DELETE FROM password_resets WHERE token_hash = ?`, tokenHash)
	return err
}
