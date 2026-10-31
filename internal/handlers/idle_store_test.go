package handlers

import (
	"testing"
	"time"
)

func TestSettleRestRegeneratesWholeHours(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Mia", "mia@example.com")
	t0 := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)
	if _, err := conn.Exec(`UPDATE characters SET mood = 40, mood_at = ? WHERE id = ?`,
		t0.Format(time.RFC3339Nano), char.ID); err != nil {
		t.Fatal(err)
	}

	// 2 h 30 min of rest: +10, the half hour stays banked.
	if _, err := SettleRest(conn, char, t0.Add(150*time.Minute)); err != nil {
		t.Fatal(err)
	}
	var mood int
	var at string
	conn.QueryRow(`SELECT mood, mood_at FROM characters WHERE id = ?`, char.ID).Scan(&mood, &at)
	if mood != 50 {
		t.Fatalf("mood = %d, want 50", mood)
	}
	if want := t0.Add(2 * time.Hour).Format(time.RFC3339Nano); at != want {
		t.Fatalf("mood_at = %s, want %s", at, want)
	}

	// A second call a minute later adds nothing.
	if _, err := SettleRest(conn, char, t0.Add(151*time.Minute)); err != nil {
		t.Fatal(err)
	}
	conn.QueryRow(`SELECT mood FROM characters WHERE id = ?`, char.ID).Scan(&mood)
	if mood != 50 {
		t.Fatalf("mood after reload = %d, want 50", mood)
	}
}

func TestSettleRestNoRegenWhileBusy(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Ben", "ben@example.com")
	t0 := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)
	if _, err := conn.Exec(`UPDATE characters SET mood = 40, mood_at = ? WHERE id = ?`,
		t0.Format(time.RFC3339Nano), char.ID); err != nil {
		t.Fatal(err)
	}
	if err := StartCollect(conn, char, time.Hour, t0); err != nil {
		t.Fatal(err)
	}
	if _, err := SettleRest(conn, char, t0.Add(10*time.Hour)); err != nil {
		t.Fatal(err)
	}
	var mood int
	conn.QueryRow(`SELECT mood FROM characters WHERE id = ?`, char.ID).Scan(&mood)
	if mood != 40 {
		t.Fatalf("mood = %d, want 40 (busy)", mood)
	}
}

func TestSettleRestPaysPassiveIncomeOnce(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Eva", "eva@example.com")
	t0 := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)
	// Gitarre (id 4): 50 c/h passive_income from the seed data.
	if _, err := conn.Exec(`INSERT INTO character_items (character_id, item_id, quantity) VALUES (?, 4, 1)`, char.ID); err != nil {
		t.Fatal(err)
	}
	// First call only sets the anchor.
	if s, err := SettleRest(conn, char, t0); err != nil || s != nil {
		t.Fatalf("anchor call: %v %v", s, err)
	}
	later := t0.Add(3*time.Hour + 20*time.Minute)
	s, err := SettleRest(conn, char, later)
	if err != nil || s == nil {
		t.Fatalf("payout: %v %v", s, err)
	}
	var money int
	conn.QueryRow(`SELECT money_cents FROM characters WHERE id = ?`, char.ID).Scan(&money)
	if money != 1000+3*50 {
		t.Fatalf("money = %d, want %d", money, 1000+3*50)
	}
	if s, _ := SettleRest(conn, char, later.Add(time.Minute)); s != nil {
		t.Fatal("paid twice")
	}
}

func TestCapacityIncludesHousing(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Tim", "tim@example.com")
	base, _ := capacityOf(conn, char)
	// WG-Zimmer (id 6): +5000.
	if _, err := conn.Exec(`INSERT INTO character_items (character_id, item_id, quantity) VALUES (?, 6, 1)`, char.ID); err != nil {
		t.Fatal(err)
	}
	got, _ := capacityOf(conn, char)
	if got != base+5000 {
		t.Fatalf("capacity = %d, want %d", got, base+5000)
	}
}
