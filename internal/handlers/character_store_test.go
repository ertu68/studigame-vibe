package handlers

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"pennergame/internal/auth"
	"pennergame/internal/db"
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

func TestCreateCharacter(t *testing.T) {
	conn := testDB(t)
	userID, err := auth.Register(conn, "Ida", "ida@example.com", "hunter22")
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	char, err := CreateCharacter(conn, userID, "koeln", "Schnorrer")
	if err != nil {
		t.Fatalf("create character: %v", err)
	}
	if char.MoneyCents != 1000 {
		t.Fatalf("start capital: got %d, want 1000", char.MoneyCents)
	}
	if char.DistrictName != "Altstadt" {
		t.Fatalf("starting district: got %q, want Altstadt", char.DistrictName)
	}

	// Start capital must be traceable in the ledger.
	var delta, balance int
	err = conn.QueryRow(
		`SELECT delta_cents, balance_after FROM ledger WHERE character_id = ?`, char.ID,
	).Scan(&delta, &balance)
	if err != nil {
		t.Fatalf("ledger lookup: %v", err)
	}
	if delta != 1000 || balance != 1000 {
		t.Fatalf("ledger entry: got delta=%d balance=%d, want 1000/1000", delta, balance)
	}

	loaded, err := CharacterForUser(conn, userID, char.EditionID)
	if err != nil {
		t.Fatalf("load character: %v", err)
	}
	if loaded.Name != "Schnorrer" {
		t.Fatalf("loaded name: got %q, want Schnorrer", loaded.Name)
	}
}

func TestCreateCharacterRejectsDuplicates(t *testing.T) {
	conn := testDB(t)
	first, _ := auth.Register(conn, "Jan", "jan@example.com", "hunter22")
	second, _ := auth.Register(conn, "Kim", "kim@example.com", "hunter22")

	if _, err := CreateCharacter(conn, first, "koeln", "Schnorrer"); err != nil {
		t.Fatalf("create first character: %v", err)
	}

	// Same user, same edition — one character per edition.
	if _, err := CreateCharacter(conn, first, "koeln", "Zweitfigur"); !errors.Is(err, ErrCharacterExists) {
		t.Fatalf("expected ErrCharacterExists, got %v", err)
	}

	// Different user, name already taken in that edition.
	if _, err := CreateCharacter(conn, second, "koeln", "Schnorrer"); !errors.Is(err, ErrNameTaken) {
		t.Fatalf("expected ErrNameTaken, got %v", err)
	}

	// The same name in another edition is fine — editions are separate worlds.
	if _, err := CreateCharacter(conn, first, "dortmund", "Schnorrer"); err != nil {
		t.Fatalf("same name in other edition: %v", err)
	}
}

func TestCreateCharacterRejectsClosedEdition(t *testing.T) {
	conn := testDB(t)
	userID, _ := auth.Register(conn, "Lars", "lars@example.com", "hunter22")

	if _, err := CreateCharacter(conn, userID, "duisburg", "Frueh"); !errors.Is(err, ErrEditionNotFound) {
		t.Fatalf("expected ErrEditionNotFound for closed edition, got %v", err)
	}
}

func TestActiveEditions(t *testing.T) {
	conn := testDB(t)
	editions, err := ActiveEditions(conn)
	if err != nil {
		t.Fatalf("active editions: %v", err)
	}

	open := make(map[string]bool, len(editions))
	for _, e := range editions {
		open[e.Slug] = true
	}
	for _, slug := range []string{"koeln", "dortmund", "duesseldorf"} {
		if !open[slug] {
			t.Errorf("edition %q should be open", slug)
		}
	}
	if open["duisburg"] {
		t.Error("edition \"duisburg\" should still be closed")
	}
}

func TestFillPercent(t *testing.T) {
	tests := []struct {
		cents, capacity, want int
	}{
		{0, 500, 0},
		{250, 500, 50},
		{500, 500, 100},
		{900, 500, 100}, // capped, overflow is charity's problem
		{100, 0, 0},     // no container yet
	}
	for _, tt := range tests {
		if got := fillPercent(tt.cents, tt.capacity); got != tt.want {
			t.Errorf("fillPercent(%d, %d) = %d, want %d", tt.cents, tt.capacity, got, tt.want)
		}
	}
}

func TestFormatCents(t *testing.T) {
	tests := []struct {
		cents int
		want  string
	}{
		{0, "0,00 €"},
		{5, "0,05 €"},
		{1000, "10,00 €"},
		{123456, "1234,56 €"},
	}
	for _, tt := range tests {
		if got := formatCents(tt.cents); got != tt.want {
			t.Errorf("formatCents(%d) = %q, want %q", tt.cents, got, tt.want)
		}
	}
}
