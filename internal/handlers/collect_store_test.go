package handlers

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"pennergame/internal/auth"
)

// newCharacter creates a verified user with a character in Köln.
func newCharacter(t *testing.T, conn *sql.DB, name, email string) Character {
	t.Helper()
	userID, err := auth.Register(conn, name, email, "hunter22")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	char, err := CreateCharacter(conn, userID, "koeln", name)
	if err != nil {
		t.Fatalf("create character: %v", err)
	}
	return char
}

func TestCollectHappyPath(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Mara", "mara@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	if err := StartCollect(conn, char, 1.0, time.Hour, start); err != nil {
		t.Fatalf("start collect: %v", err)
	}

	// Nothing is due while the shift runs.
	summary, err := ResolveDueAction(conn, char, start.Add(30*time.Minute), 0.5)
	if err != nil {
		t.Fatalf("resolve too early: %v", err)
	}
	if summary != nil {
		t.Fatalf("shift resolved before it ended: %+v", summary)
	}

	done := start.Add(time.Hour)
	summary, err = ResolveDueAction(conn, char, done, 0.5)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if summary == nil {
		t.Fatal("expected a summary for a finished shift")
	}
	if summary.Bottles != 40 {
		t.Fatalf("bottles: got %d, want 40", summary.Bottles)
	}

	reloaded, err := CharacterForUser(conn, char.UserID, char.EditionID)
	if err != nil {
		t.Fatalf("reload character: %v", err)
	}
	// 40 bottles at 25 cents on top of the 10 € starting capital.
	if reloaded.MoneyCents != char.MoneyCents+1000 {
		t.Fatalf("money: got %d, want %d", reloaded.MoneyCents, char.MoneyCents+1000)
	}
	if reloaded.Mood != 95 {
		t.Fatalf("mood: got %d, want 95", reloaded.Mood)
	}
}

// A finished shift must pay out exactly once, however often it is resolved.
func TestCollectResolvesOnlyOnce(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Nils", "nils@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)
	done := start.Add(time.Hour)

	// Give the character room so the payout is actually booked.
	if _, err := conn.Exec(`UPDATE characters SET money_cents = 0 WHERE id = ?`, char.ID); err != nil {
		t.Fatalf("prepare character: %v", err)
	}
	char.MoneyCents = 0

	if err := StartCollect(conn, char, 1.0, time.Hour, start); err != nil {
		t.Fatalf("start collect: %v", err)
	}
	first, err := ResolveDueAction(conn, char, done, 0.5)
	if err != nil || first == nil {
		t.Fatalf("first resolve: summary=%v err=%v", first, err)
	}

	second, err := ResolveDueAction(conn, char, done, 0.5)
	if err != nil {
		t.Fatalf("second resolve: %v", err)
	}
	if second != nil {
		t.Fatalf("shift paid out twice: %+v", second)
	}

	var entries int
	if err := conn.QueryRow(
		`SELECT count(*) FROM ledger WHERE character_id = ? AND reason = 'collect'`, char.ID,
	).Scan(&entries); err != nil {
		t.Fatalf("count ledger: %v", err)
	}
	if entries != 1 {
		t.Fatalf("ledger entries: got %d, want 1", entries)
	}
}

func TestStartCollectRejectsSecondShift(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Ove", "ove@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	if err := StartCollect(conn, char, 1.0, time.Hour, start); err != nil {
		t.Fatalf("start collect: %v", err)
	}
	if err := StartCollect(conn, char, 1.0, time.Hour, start); !errors.Is(err, ErrActionRunning) {
		t.Fatalf("expected ErrActionRunning, got %v", err)
	}
}

func TestStartCollectRejectsUnknownDuration(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Pia", "pia@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	if err := StartCollect(conn, char, 1.0, 42*time.Minute, start); !errors.Is(err, ErrInvalidShift) {
		t.Fatalf("expected ErrInvalidShift, got %v", err)
	}
	if _, err := FindOpenAction(conn, char.ID); !errors.Is(err, ErrNoOpenAction) {
		t.Fatalf("rejected shift was stored anyway: %v", err)
	}
}

// Everything above the container capacity goes to charity, nothing is lost.
func TestCollectOverflowGoesToCharity(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Rico", "rico@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	// Level 0 holds 20 €, the character already carries 19,50 €.
	if _, err := conn.Exec(`UPDATE characters SET money_cents = 1950 WHERE id = ?`, char.ID); err != nil {
		t.Fatalf("prepare character: %v", err)
	}
	char.MoneyCents = 1950

	if err := StartCollect(conn, char, 1.0, time.Hour, start); err != nil {
		t.Fatalf("start collect: %v", err)
	}
	summary, err := ResolveDueAction(conn, char, start.Add(time.Hour), 0.5)
	if err != nil || summary == nil {
		t.Fatalf("resolve: summary=%v err=%v", summary, err)
	}
	if !summary.HasDonation {
		t.Fatal("expected the overflow to be donated")
	}

	reloaded, err := CharacterForUser(conn, char.UserID, char.EditionID)
	if err != nil {
		t.Fatalf("reload character: %v", err)
	}
	if reloaded.MoneyCents != 2000 {
		t.Fatalf("money: got %d, want the 2000 cap", reloaded.MoneyCents)
	}

	pool, err := CharityPool(conn, char.EditionID)
	if err != nil {
		t.Fatalf("charity pool: %v", err)
	}
	// 40 bottles at 25 cents is 1000; 50 fit, 950 are donated.
	if pool != 950 {
		t.Fatalf("charity pool: got %d, want 950", pool)
	}
}

func TestOpenActionRemaining(t *testing.T) {
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)
	a := OpenAction{StartedAt: start, EndsAt: start.Add(time.Hour)}

	if got := a.Remaining(start.Add(20 * time.Minute)); got != 40*time.Minute {
		t.Errorf("remaining: got %v, want 40m", got)
	}
	if got := a.Remaining(start.Add(2 * time.Hour)); got != 0 {
		t.Errorf("remaining after the end: got %v, want 0", got)
	}
	if a.Done(start.Add(time.Minute)) {
		t.Error("running shift reported as done")
	}
	if !a.Done(start.Add(time.Hour)) {
		t.Error("finished shift not reported as done")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{15 * time.Minute, "15 min"},
		{time.Hour, "1 h"},
		{4 * time.Hour, "4 h"},
		{90 * time.Minute, "1 h 30 min"},
		{45 * time.Second, "45 s"},
		// Truncated, never rounded up: a countdown must not show more than is left.
		{119 * time.Second, "1 min"},
		{0, "fertig"},
		{-time.Minute, "fertig"},
	}
	for _, tt := range tests {
		if got := formatDuration(tt.d); got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
