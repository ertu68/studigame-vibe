package handlers

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestJobHappyPath(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Jan", "jan@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	if err := StartJob(conn, char, time.Hour, start); err != nil {
		t.Fatalf("start job: %v", err)
	}

	// Job has no luck: the roll must be ignored (pass a nonsense roll).
	summary, err := ResolveDueAction(conn, char, start.Add(time.Hour), 0.5)
	if err != nil || summary == nil {
		t.Fatalf("resolve: summary=%v err=%v", summary, err)
	}
	if summary.Title != "Werkstudenten-Job erledigt" {
		t.Errorf("title: got %q", summary.Title)
	}
	if summary.Detail != "8,00 € verdient" {
		t.Errorf("detail: got %q, want %q", summary.Detail, "8,00 € verdient")
	}

	reloaded, err := CharacterForUser(conn, char.UserID, char.EditionID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.MoneyCents != char.MoneyCents+800 {
		t.Fatalf("money: got %d, want %d", reloaded.MoneyCents, char.MoneyCents+800)
	}
	if reloaded.Mood != 95 {
		t.Fatalf("mood: got %d, want 95", reloaded.Mood)
	}

	var reason string
	var delta int
	err = conn.QueryRow(
		`SELECT reason, delta_cents FROM ledger WHERE character_id = ? AND ref_id IS NOT NULL`, char.ID,
	).Scan(&reason, &delta)
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	if reason != "job" || delta != 800 {
		t.Fatalf("ledger row: reason=%q delta=%d, want job/800", reason, delta)
	}
}

// Job earnings use the same container+charity booking path as collecting.
func TestJobOverflowGoesToCharity(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Kim", "kim@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	// Level 0 holds 20 €, the character already carries 15 €; a one-hour job
	// earns 8 €, so only 5 € fit and 3 € are donated.
	if _, err := conn.Exec(`UPDATE characters SET money_cents = 1500 WHERE id = ?`, char.ID); err != nil {
		t.Fatalf("prepare character: %v", err)
	}
	char.MoneyCents = 1500

	if err := StartJob(conn, char, time.Hour, start); err != nil {
		t.Fatalf("start job: %v", err)
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
		t.Fatalf("reload: %v", err)
	}
	if reloaded.MoneyCents != 2000 {
		t.Fatalf("money: got %d, want the 2000 cap", reloaded.MoneyCents)
	}

	pool, err := CharityPool(conn, char.EditionID)
	if err != nil {
		t.Fatalf("charity pool: %v", err)
	}
	if pool != 300 {
		t.Fatalf("charity pool: got %d, want 300", pool)
	}
}

func TestStartJobRejectsSecondAction(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Pat", "pat@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	if err := StartJob(conn, char, time.Hour, start); err != nil {
		t.Fatalf("start job: %v", err)
	}
	// A collect on top of a running job must be refused.
	if err := StartCollect(conn, char, time.Hour, start); !errors.Is(err, ErrActionRunning) {
		t.Fatalf("expected ErrActionRunning for collect over job, got %v", err)
	}
}

func TestStartJobRejectsUnknownDuration(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Liv", "liv@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	if err := StartJob(conn, char, 42*time.Minute, start); !errors.Is(err, ErrInvalidShift) {
		t.Fatalf("expected ErrInvalidShift, got %v", err)
	}
	if _, err := FindOpenAction(conn, char.ID); !errors.Is(err, ErrNoOpenAction) {
		t.Fatalf("rejected shift was stored anyway: %v", err)
	}
}

func TestJobResolvesOnlyOnce(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Mio", "mio@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)
	done := start.Add(time.Hour)

	if _, err := conn.Exec(`UPDATE characters SET money_cents = 0 WHERE id = ?`, char.ID); err != nil {
		t.Fatalf("prepare character: %v", err)
	}
	char.MoneyCents = 0

	if err := StartJob(conn, char, time.Hour, start); err != nil {
		t.Fatalf("start job: %v", err)
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
		t.Fatalf("job paid out twice: %+v", second)
	}

	var reason string
	err = conn.QueryRow(
		`SELECT reason FROM ledger WHERE character_id = ?`, char.ID,
	).Scan(&reason)
	if err == sql.ErrNoRows {
		t.Fatal("no ledger entry for the job")
	}
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
}
