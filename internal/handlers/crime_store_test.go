package handlers

import (
	"errors"
	"testing"
	"time"

	"studigame/internal/game"
)

func TestCrimeCleanPayout(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Nora", "nora@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	// Empty pockets, 1 h clean run (roll 0.5 > 0.30): the full 15 € land.
	if _, err := conn.Exec(`UPDATE characters SET money_cents = 0 WHERE id = ?`, char.ID); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	char.MoneyCents = 0

	if err := StartCrime(conn, char, time.Hour, start); err != nil {
		t.Fatalf("start: %v", err)
	}
	summary, err := ResolveDueAction(conn, char, start.Add(time.Hour), 0.5)
	if err != nil || summary == nil {
		t.Fatalf("resolve: summary=%v err=%v", summary, err)
	}

	reloaded, err := CharacterForUser(conn, char.UserID, char.EditionID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.MoneyCents != game.CrimeRatePerHourCents {
		t.Fatalf("money: got %d, want %d", reloaded.MoneyCents, game.CrimeRatePerHourCents)
	}
}

func TestCrimeCaughtPayout(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Finn", "finn@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	if _, err := conn.Exec(`UPDATE characters SET money_cents = 0 WHERE id = ?`, char.ID); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	char.MoneyCents = 0

	if err := StartCrime(conn, char, time.Hour, start); err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, err := ResolveDueAction(conn, char, start.Add(time.Hour), 0.1); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	reloaded, err := CharacterForUser(conn, char.UserID, char.EditionID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	want := int(float64(game.CrimeRatePerHourCents) * game.CrimeCatchPayoutFactor)
	if reloaded.MoneyCents != want {
		t.Fatalf("caught money: got %d, want %d", reloaded.MoneyCents, want)
	}
}

func TestStartCrimeRejectsSecondAction(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Pat", "pat@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	if err := StartJob(conn, char, time.Hour, start); err != nil {
		t.Fatalf("start job: %v", err)
	}
	if err := StartCrime(conn, char, time.Hour, start); !errors.Is(err, ErrActionRunning) {
		t.Fatalf("expected ErrActionRunning over a running job, got %v", err)
	}
}

func TestStartCrimeRejectsUnknownDuration(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Liv", "liv@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	if err := StartCrime(conn, char, 42*time.Minute, start); !errors.Is(err, ErrInvalidShift) {
		t.Fatalf("expected ErrInvalidShift, got %v", err)
	}
}
