package auth

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"studigame/internal/config"
	"studigame/internal/db"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestRegisterAndAuthenticate(t *testing.T) {
	conn := testDB(t)

	id, err := Register(conn, "Alice", "alice@example.com", "hunter22", "Köln")
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if _, err := Register(conn, "Alice2", "alice@example.com", "hunter22", "Köln"); err != ErrEmailTaken {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}

	// Unverified user cannot authenticate
	if _, err := Authenticate(conn, "alice@example.com", "hunter22"); err != ErrUnverifiedEmail {
		t.Fatalf("expected ErrUnverifiedEmail for new account, got %v", err)
	}

	// After verification, authentication succeeds
	var mailer fakeMailer
	if err := RequestEmailVerification(conn, &mailer, "example.com", id, "alice@example.com", time.Hour); err != nil {
		t.Fatalf("request verification: %v", err)
	}
	token := tokenFromLink(t, mailer.body)
	if err := VerifyEmail(conn, token); err != nil {
		t.Fatalf("verify email: %v", err)
	}

	got, err := Authenticate(conn, "alice@example.com", "hunter22")
	if err != nil || got != id {
		t.Fatalf("authenticate after verify: got id=%d err=%v, want id=%d", got, err, id)
	}

	if _, err := Authenticate(conn, "alice@example.com", "wrong"); err != ErrInvalidLogin {
		t.Fatalf("expected ErrInvalidLogin, got %v", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	conn := testDB(t)
	id, _ := Register(conn, "Bob", "bob@example.com", "hunter22", "Köln")
	store := NewSessionStore(conn)

	token, err := store.Create(id)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	sess, err := store.Get(token)
	if err != nil || sess.UserID != id {
		t.Fatalf("get session: sess=%+v err=%v", sess, err)
	}

	if err := store.Delete(token); err != nil {
		t.Fatalf("delete session: %v", err)
	}
	if _, err := store.Get(token); err != ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound after delete, got %v", err)
	}
}

func TestPasswordReset(t *testing.T) {
	conn := testDB(t)
	id, _ := Register(conn, "Carol", "carol@example.com", "oldpassword", "Köln")

	// Verify email first so Carol can authenticate
	var verifyMailer fakeMailer
	if err := RequestEmailVerification(conn, &verifyMailer, "example.com", id, "carol@example.com", time.Hour); err != nil {
		t.Fatalf("request verification: %v", err)
	}
	verifyToken := tokenFromLink(t, verifyMailer.body)
	if err := VerifyEmail(conn, verifyToken); err != nil {
		t.Fatalf("verify email: %v", err)
	}

	var resetMailer fakeMailer
	err := RequestPasswordReset(conn, &resetMailer, "example.invalid", "carol@example.com", time.Hour)
	if err != nil {
		t.Fatalf("request reset: %v", err)
	}
	if resetMailer.to == "" {
		t.Fatalf("expected mailer to be called")
	}

	token := tokenFromLink(t, resetMailer.body)
	if err := ResetPassword(conn, token, "newpassword"); err != nil {
		t.Fatalf("reset password: %v", err)
	}

	if _, err := Authenticate(conn, "carol@example.com", "newpassword"); err != nil {
		t.Fatalf("authenticate with new password: %v", err)
	}
	if err := ResetPassword(conn, token, "again"); err != ErrInvalidResetLink {
		t.Fatalf("expected token to be single-use, got %v", err)
	}
}

func TestValidCSRFRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	token := EnsureCSRFToken(rec, req, false)

	newRequest := func(submitted string) *http.Request {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		for _, c := range rec.Result().Cookies() {
			r.AddCookie(c)
		}
		r.PostForm = map[string][]string{"csrf_token": {submitted}}
		return r
	}

	if !ValidCSRFRequest(newRequest(token)) {
		t.Fatalf("expected matching csrf token to be valid")
	}
	if ValidCSRFRequest(newRequest("wrong")) {
		t.Fatalf("expected mismatched csrf token to be invalid")
	}
}

type fakeMailer struct{ to, subject, body string }

func (m *fakeMailer) Send(to, subject, body string) error {
	m.to, m.subject, m.body = to, subject, body
	return nil
}

func tokenFromLink(t *testing.T, body string) string {
	t.Helper()
	const marker = "token="
	i := strings.Index(body, marker)
	if i < 0 {
		t.Fatalf("no token in body: %q", body)
	}
	return body[i+len(marker):]
}

func TestNewMailer(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.Config
		wantType string
	}{
		{
			name:     "Empty SMTP host returns ConsoleMailer",
			cfg:      config.Config{SMTPHost: ""},
			wantType: "ConsoleMailer",
		},
		{
			name: "SMTP config returns SMTPMailer",
			cfg: config.Config{
				SMTPHost: "smtp.example.com",
				SMTPPort: "587",
				SMTPUser: "user",
				SMTPPass: "pass",
				MailFrom: "noreply@example.com",
			},
			wantType: "SMTPMailer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := NewMailer(tt.cfg)
			switch v := mailer.(type) {
			case ConsoleMailer:
				if tt.wantType != "ConsoleMailer" {
					t.Fatalf("got ConsoleMailer, want %s", tt.wantType)
				}
			case SMTPMailer:
				if tt.wantType != "SMTPMailer" {
					t.Fatalf("got SMTPMailer, want %s", tt.wantType)
				}
				if v.host != tt.cfg.SMTPHost {
					t.Fatalf("SMTP host mismatch: got %q, want %q", v.host, tt.cfg.SMTPHost)
				}
			default:
				t.Fatalf("unknown mailer type: %T", v)
			}
		})
	}
}

func TestEmailVerification(t *testing.T) {
	conn := testDB(t)
	id, _ := Register(conn, "Diana", "diana@example.com", "hunter22", "Köln")

	var mailer fakeMailer
	if err := RequestEmailVerification(conn, &mailer, "example.com", id, "diana@example.com", time.Hour); err != nil {
		t.Fatalf("request verification: %v", err)
	}
	if mailer.to == "" {
		t.Fatalf("expected mailer to be called")
	}

	token := tokenFromLink(t, mailer.body)
	if err := VerifyEmail(conn, token); err != nil {
		t.Fatalf("verify email: %v", err)
	}

	// Token is single-use
	if err := VerifyEmail(conn, token); err != ErrInvalidVerifyLink {
		t.Fatalf("expected token to be single-use, got %v", err)
	}

	// Now authentication should succeed
	got, err := Authenticate(conn, "diana@example.com", "hunter22")
	if err != nil || got != id {
		t.Fatalf("authenticate after verification: got id=%d err=%v, want id=%d", got, err, id)
	}
}

func TestUnverifiedStatusOnlyForCorrectPassword(t *testing.T) {
	conn := testDB(t)
	Register(conn, "Finn", "finn@example.com", "hunter22", "Köln")
	if _, err := Authenticate(conn, "finn@example.com", "wrong"); err != ErrInvalidLogin {
		t.Fatalf("wrong password on unverified account: got %v, want ErrInvalidLogin", err)
	}
	if _, err := Authenticate(conn, "nobody@example.com", "hunter22"); err != ErrInvalidLogin {
		t.Fatalf("unknown email: got %v, want ErrInvalidLogin", err)
	}
}

func TestPasswordProblem(t *testing.T) {
	cases := []struct {
		pw string
		ok bool
	}{
		{"short", false}, {"hunter22", true}, {strings.Repeat("a", 72), true},
		{strings.Repeat("a", 73), false}, {"üüüüüüüü", true}, // 8 runes, 16 bytes
	}
	for _, tc := range cases {
		if got := PasswordProblem(tc.pw) == ""; got != tc.ok {
			t.Errorf("PasswordProblem(%d bytes) ok=%v, want %v", len(tc.pw), got, tc.ok)
		}
	}
}

func TestResendVerification(t *testing.T) {
	conn := testDB(t)
	id, _ := Register(conn, "Gina", "gina@example.com", "hunter22", "Köln")
	var mailer fakeMailer

	// Unknown address: nil, no mail.
	if err := ResendVerification(conn, &mailer, "example.com", "nobody@example.com", time.Hour); err != nil || mailer.to != "" {
		t.Fatalf("unknown: err=%v to=%q", err, mailer.to)
	}
	// Known, no recent link: mail goes out and the link works.
	if err := ResendVerification(conn, &mailer, "example.com", "gina@example.com", time.Hour); err != nil || mailer.to == "" {
		t.Fatalf("resend: err=%v to=%q", err, mailer.to)
	}
	if err := VerifyEmail(conn, tokenFromLink(t, mailer.body)); err != nil {
		t.Fatalf("verify resent link: %v", err)
	}
	// Already verified: no second mail.
	mailer = fakeMailer{}
	if err := ResendVerification(conn, &mailer, "example.com", "gina@example.com", time.Hour); err != nil || mailer.to != "" {
		t.Fatalf("verified: err=%v to=%q", err, mailer.to)
	}
	_ = id
}

func TestResendVerificationThrottled(t *testing.T) {
	conn := testDB(t)
	Register(conn, "Hans", "hans@example.com", "hunter22", "Köln")
	var mailer fakeMailer
	ResendVerification(conn, &mailer, "example.com", "hans@example.com", time.Hour)
	mailer = fakeMailer{}
	ResendVerification(conn, &mailer, "example.com", "hans@example.com", time.Hour)
	if mailer.to != "" {
		t.Fatal("second resend within a minute should not mail")
	}
}

func TestExpiredResetTokenIsBurned(t *testing.T) {
	conn := testDB(t)
	id, _ := Register(conn, "Ida", "ida@example.com", "hunter22", "Köln")
	conn.Exec(`INSERT INTO password_resets (token_hash, user_id, expires_at) VALUES (?, ?, ?)`,
		hashToken("tok"), id, time.Now().UTC().Add(-time.Minute))
	if err := ResetPassword(conn, "tok", "newpassword"); err != ErrInvalidResetLink {
		t.Fatalf("got %v, want ErrInvalidResetLink", err)
	}
	var n int
	conn.QueryRow(`SELECT COUNT(*) FROM password_resets`).Scan(&n)
	if n != 0 {
		t.Fatal("expired token should be deleted")
	}
}
