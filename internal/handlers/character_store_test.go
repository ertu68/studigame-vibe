package handlers

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"studigame/internal/auth"
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

// subjectID resolves a Studienfach name to its id in the single test edition.
func subjectID(t *testing.T, conn *sql.DB, name string) int64 {
	t.Helper()
	var id int64
	err := conn.QueryRow(`SELECT s.id FROM subjects s
		JOIN editions e ON e.id = s.edition_id
		WHERE e.slug = 'studistadt' AND s.name = ?`, name).Scan(&id)
	if err != nil {
		t.Fatalf("resolve subject %q: %v", name, err)
	}
	return id
}

func TestCreateCharacter(t *testing.T) {
	conn := testDB(t)
	userID, err := auth.Register(conn, "Ida", "ida@example.com", "hunter22", "Köln")
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	char, err := CreateCharacter(conn, userID, "studistadt", subjectID(t, conn, "Informatik"), "Schnorrer")
	if err != nil {
		t.Fatalf("create character: %v", err)
	}
	if char.MoneyCents != 1000 {
		t.Fatalf("start capital: got %d, want 1000", char.MoneyCents)
	}
	if char.SubjectName != "Informatik" {
		t.Fatalf("starting subject: got %q, want Informatik", char.SubjectName)
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
	first, _ := auth.Register(conn, "Jan", "jan@example.com", "hunter22", "Köln")
	second, _ := auth.Register(conn, "Kim", "kim@example.com", "hunter22", "Köln")
	info := subjectID(t, conn, "Informatik")

	if _, err := CreateCharacter(conn, first, "studistadt", info, "Schnorrer"); err != nil {
		t.Fatalf("create first character: %v", err)
	}

	// Same user, same edition — one character per edition.
	if _, err := CreateCharacter(conn, first, "studistadt", info, "Zweitfigur"); !errors.Is(err, ErrCharacterExists) {
		t.Fatalf("expected ErrCharacterExists, got %v", err)
	}

	// Different user, name already taken in that edition.
	if _, err := CreateCharacter(conn, second, "studistadt", info, "Schnorrer"); !errors.Is(err, ErrNameTaken) {
		t.Fatalf("expected ErrNameTaken, got %v", err)
	}

	// The same name in another edition is fine — editions are separate worlds.
	if _, err := conn.Exec(
		`INSERT INTO editions (slug, name, active) VALUES ('campus2', 'Campus 2', TRUE)`); err != nil {
		t.Fatalf("add second edition: %v", err)
	}
	if _, err := conn.Exec(
		`INSERT INTO subjects (edition_id, name)
		 SELECT id, 'Informatik' FROM editions WHERE slug = 'campus2'`); err != nil {
		t.Fatalf("add second subject: %v", err)
	}
	var otherID int64
	if err := conn.QueryRow(
		`SELECT s.id FROM subjects s JOIN editions e ON e.id = s.edition_id
		 WHERE e.slug = 'campus2' LIMIT 1`).Scan(&otherID); err != nil {
		t.Fatalf("resolve other subject: %v", err)
	}
	if _, err := CreateCharacter(conn, first, "campus2", otherID, "Schnorrer"); err != nil {
		t.Fatalf("same name in other edition: %v", err)
	}
}

func TestCreateCharacterRejectsUnknownSubject(t *testing.T) {
	conn := testDB(t)
	userID, _ := auth.Register(conn, "Lars", "lars@example.com", "hunter22", "Köln")

	if _, err := CreateCharacter(conn, userID, "studistadt", 9999, "Frueh"); !errors.Is(err, ErrSubjectNotFound) {
		t.Fatalf("expected ErrSubjectNotFound for unknown subject, got %v", err)
	}
}

func TestCreateCharacterRejectsClosedEdition(t *testing.T) {
	conn := testDB(t)
	userID, _ := auth.Register(conn, "Mia", "mia@example.com", "hunter22", "Köln")

	if _, err := CreateCharacter(conn, userID, "campus2", subjectID(t, conn, "Informatik"), "Frueh"); !errors.Is(err, ErrEditionNotFound) {
		t.Fatalf("expected ErrEditionNotFound for unknown edition, got %v", err)
	}
}

func TestActiveEditions(t *testing.T) {
	conn := testDB(t)
	editions, err := ActiveEditions(conn)
	if err != nil {
		t.Fatalf("active editions: %v", err)
	}
	if len(editions) != 1 || editions[0].Slug != "studistadt" {
		t.Fatalf("expected exactly the studistadt edition, got %+v", editions)
	}
}

func TestActiveSubjects(t *testing.T) {
	conn := testDB(t)
	edition, err := ActiveEditions(conn)
	if err != nil {
		t.Fatalf("active editions: %v", err)
	}
	subjects, err := ActiveSubjects(conn, edition[0].ID)
	if err != nil {
		t.Fatalf("active subjects: %v", err)
	}
	want := map[string]bool{"Informatik": true, "BWL": true, "Jura": true, "Geisteswissenschaften": true}
	if len(subjects) != len(want) {
		t.Fatalf("subjects: got %d, want %d", len(subjects), len(want))
	}
	for _, s := range subjects {
		if !want[s.Name] {
			t.Errorf("unexpected subject %q", s.Name)
		}
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
