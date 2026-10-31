package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

var (
	ErrEmailTaken        = errors.New("email already registered")
	ErrInvalidLogin      = errors.New("invalid email or password")
	ErrInvalidResetLink  = errors.New("invalid or expired reset link")
	ErrUnverifiedEmail   = errors.New("please verify your email first")
	ErrInvalidVerifyLink = errors.New("invalid or expired verification link")
)

// UniCities is the hardcoded list of German university cities offered at
// registration. Display only for now, no gameplay effect.
var UniCities = []string{
	"Aachen", "Augsburg", "Bamberg", "Bayreuth", "Berlin", "Bielefeld",
	"Bochum", "Bonn", "Braunschweig", "Bremen", "Chemnitz", "Cottbus",
	"Darmstadt", "Dortmund", "Dresden", "Duisburg", "Düsseldorf", "Erfurt",
	"Erlangen", "Essen", "Frankfurt am Main", "Freiburg", "Gießen",
	"Göttingen", "Greifswald", "Halle", "Hamburg", "Hannover", "Heidelberg",
	"Ilmenau", "Jena", "Kaiserslautern", "Karlsruhe", "Kassel", "Kiel",
	"Koblenz", "Köln", "Konstanz", "Leipzig", "Lübeck", "Magdeburg",
	"Mainz", "Mannheim", "Marburg", "München", "Münster", "Oldenburg",
	"Osnabrück", "Paderborn", "Passau", "Potsdam", "Regensburg", "Rostock",
	"Saarbrücken", "Siegen", "Stuttgart", "Trier", "Tübingen", "Ulm",
	"Wuppertal", "Würzburg",
}

func Register(db *sql.DB, name, email, password, city string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	res, err := db.Exec(
		`INSERT INTO users (name, email, password_hash, city) VALUES (?, ?, ?, ?)`,
		name, email, string(hash), city,
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

// dummyHash lets Authenticate spend the same bcrypt time for unknown emails
// as for known ones, so response time does not reveal which accounts exist.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy"), bcrypt.DefaultCost)

// MaxPasswordBytes is bcrypt's input limit; longer passwords are rejected
// instead of silently truncated.
const MaxPasswordBytes = 72

// PasswordProblem returns a German message if the password is unusable, else "".
func PasswordProblem(password string) string {
	switch {
	case utf8.RuneCountInString(password) < 8:
		return "Passwort muss mind. 8 Zeichen haben"
	case len(password) > MaxPasswordBytes:
		return "Passwort ist zu lang (max. 72 Bytes)"
	}
	return ""
}

// Authenticate checks the password first; only someone who knows it learns
// that the account is still unverified. Unknown email and wrong password are
// indistinguishable, in answer and in time.
func Authenticate(db *sql.DB, email, password string) (int64, error) {
	var id int64
	var hash string
	var verified bool
	err := db.QueryRow(`SELECT id, password_hash, verified FROM users WHERE email = ?`, email).Scan(&id, &hash, &verified)
	if err == sql.ErrNoRows {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return 0, ErrInvalidLogin
	}
	if err != nil {
		return 0, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return 0, ErrInvalidLogin
	}
	if !verified {
		return 0, ErrUnverifiedEmail
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
		hashToken(rawToken), userID, time.Now().UTC().Add(ttl),
	)
	if err != nil {
		return err
	}

	link := fmt.Sprintf("https://%s/reset-password/confirm?token=%s", siteDomain, rawToken)
	return mailer.Send(email, "Passwort zuruecksetzen", "Link: "+link)
}

// ResetPassword spends the token and sets the password in one transaction;
// the DELETE ... RETURNING is the claim, so a token works exactly once.
func ResetPassword(db *sql.DB, token, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userID int64
	var expiresAt time.Time
	err = tx.QueryRow(
		`DELETE FROM password_resets WHERE token_hash = ? RETURNING user_id, expires_at`, hashToken(token),
	).Scan(&userID, &expiresAt)
	if err == sql.ErrNoRows {
		return ErrInvalidResetLink
	}
	if err != nil {
		return err
	}
	if time.Now().After(expiresAt) {
		tx.Commit() // expired token stays deleted
		return ErrInvalidResetLink
	}
	if _, err := tx.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, string(hash), userID); err != nil {
		return err
	}
	// D4: Invalidate all sessions for this user
	if _, err := tx.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID); err != nil {
		return err
	}
	return tx.Commit()
}

// RequestEmailVerification sends a verification link to a newly registered user.
func RequestEmailVerification(db *sql.DB, mailer Mailer, siteDomain string, userID int64, email string, ttl time.Duration) error {
	rawToken, err := randomToken()
	if err != nil {
		return err
	}
	_, err = db.Exec(
		`INSERT INTO verification_tokens (token_hash, user_id, expires_at) VALUES (?, ?, ?)`,
		hashToken(rawToken), userID, time.Now().UTC().Add(ttl),
	)
	if err != nil {
		return err
	}

	link := fmt.Sprintf("https://%s/verify?token=%s", siteDomain, rawToken)
	return mailer.Send(email, "Bestätige deine E-Mail", "Link: "+link)
}

// ResendVerification sends a fresh link to an unverified account. Like
// RequestPasswordReset it returns nil whether or not the account exists. A
// link sent within the last minute is not repeated (no mail bombing).
func ResendVerification(db *sql.DB, mailer Mailer, siteDomain, email string, ttl time.Duration) error {
	var userID int64
	var recent bool
	err := db.QueryRow(
		`SELECT id, EXISTS(SELECT 1 FROM verification_tokens WHERE user_id = users.id
		                   AND created_at > datetime('now', '-1 minute'))
		 FROM users WHERE email = ? AND verified = FALSE`, email,
	).Scan(&userID, &recent)
	if err == sql.ErrNoRows || recent {
		return nil
	}
	if err != nil {
		return err
	}
	return RequestEmailVerification(db, mailer, siteDomain, userID, email, ttl)
}

// VerifyEmail marks a user as verified if the token is valid. The token is
// spent and the user updated in one transaction.
func VerifyEmail(db *sql.DB, token string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userID int64
	var expiresAt time.Time
	err = tx.QueryRow(
		`DELETE FROM verification_tokens WHERE token_hash = ? RETURNING user_id, expires_at`, hashToken(token),
	).Scan(&userID, &expiresAt)
	if err == sql.ErrNoRows {
		return ErrInvalidVerifyLink
	}
	if err != nil {
		return err
	}
	if time.Now().After(expiresAt) {
		tx.Commit() // expired token stays deleted
		return ErrInvalidVerifyLink
	}
	if _, err := tx.Exec(`UPDATE users SET verified = TRUE WHERE id = ?`, userID); err != nil {
		return err
	}
	return tx.Commit()
}
