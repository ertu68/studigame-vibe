package handlers

import (
	"testing"
	"time"
)

func TestKlausurBooksECTSOnCharacterNotLedger(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Lea", "lea@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	if err := StartKlausur(conn, char, time.Hour, start); err != nil {
		t.Fatal(err)
	}
	if s, err := ResolveDueAction(conn, char, start.Add(time.Hour), 0.5); err != nil || s == nil {
		t.Fatalf("resolve: %v %v", s, err)
	}
	got, err := CharacterForUser(conn, char.UserID, char.EditionID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ECTSCents <= 0 {
		t.Fatalf("ects = %d, want > 0", got.ECTSCents)
	}
	if got.MoneyCents != char.MoneyCents {
		t.Fatalf("money changed: %d -> %d", char.MoneyCents, got.MoneyCents)
	}
	var n int
	conn.QueryRow(`SELECT COUNT(*) FROM ledger WHERE character_id = ? AND reason LIKE '%klausur%'`, char.ID).Scan(&n)
	if n != 0 {
		t.Fatalf("ledger has %d klausur rows", n)
	}
}
