package handlers

import (
	"database/sql"
	"errors"
	"testing"

	"studigame/internal/game"
)

func TestUpgradeContainerHappyPath(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Tom", "tom@example.com")
	// Level 0 -> 1 costs 2000 cents; give the character exactly that.
	if _, err := conn.Exec(`UPDATE characters SET money_cents = 2000 WHERE id = ?`, char.ID); err != nil {
		t.Fatalf("prepare character: %v", err)
	}
	char.MoneyCents = 2000

	got, err := UpgradeContainer(conn, char)
	if err != nil {
		t.Fatalf("upgrade: %v", err)
	}
	if got.ContainerLevel != 1 {
		t.Errorf("level: got %d, want 1", got.ContainerLevel)
	}
	if got.MoneyCents != 0 {
		t.Errorf("money: got %d, want 0", got.MoneyCents)
	}

	reloaded, err := CharacterForUser(conn, char.UserID, char.EditionID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.ContainerLevel != 1 || reloaded.MoneyCents != 0 {
		t.Fatalf("stored wrong state: level=%d money=%d", reloaded.ContainerLevel, reloaded.MoneyCents)
	}

	var delta, balance int
	var reason string
	err = conn.QueryRow(
		`SELECT delta_cents, balance_after, reason FROM ledger
		 WHERE character_id = ? AND reason = 'container_upgrade'`, char.ID,
	).Scan(&delta, &balance, &reason)
	if err == sql.ErrNoRows {
		t.Fatal("no container_upgrade ledger entry")
	}
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	if delta != -2000 || balance != 0 || reason != "container_upgrade" {
		t.Fatalf("ledger row wrong: delta=%d balance=%d reason=%q", delta, balance, reason)
	}
}

func TestUpgradeContainerNotEnoughMoney(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Ada", "ada@example.com") // 1000 < 2000

	if _, err := UpgradeContainer(conn, char); !errors.Is(err, ErrNotEnoughMoney) {
		t.Fatalf("expected ErrNotEnoughMoney, got %v", err)
	}

	reloaded, err := CharacterForUser(conn, char.UserID, char.EditionID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.ContainerLevel != 0 || reloaded.MoneyCents != 1000 {
		t.Fatalf("failed upgrade changed state: level=%d money=%d",
			reloaded.ContainerLevel, reloaded.MoneyCents)
	}

	var entries int
	if err := conn.QueryRow(
		`SELECT count(*) FROM ledger WHERE reason = 'container_upgrade'`, char.ID,
	).Scan(&entries); err != nil {
		t.Fatalf("count ledger: %v", err)
	}
	if entries != 0 {
		t.Fatalf("failed upgrade booked money: %d entries", entries)
	}
}

func TestUpgradeContainerAtMaxLevel(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Sam", "sam@example.com")
	if _, err := conn.Exec(
		`UPDATE characters SET container_level = ? , money_cents = 999999 WHERE id = ?`,
		game.MaxContainerLevel, char.ID,
	); err != nil {
		t.Fatalf("prepare character: %v", err)
	}
	char.ContainerLevel = game.MaxContainerLevel
	char.MoneyCents = 999999

	if _, err := UpgradeContainer(conn, char); !errors.Is(err, ErrContainerMax) {
		t.Fatalf("expected ErrContainerMax, got %v", err)
	}

	reloaded, err := CharacterForUser(conn, char.UserID, char.EditionID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.ContainerLevel != game.MaxContainerLevel {
		t.Fatalf("level changed at max: %d", reloaded.ContainerLevel)
	}
	if reloaded.MoneyCents != 999999 {
		t.Fatalf("money was charged at max: %d", reloaded.MoneyCents)
	}
}

// Buying the next tier must not be possible with the exact balance short of it.
func TestUpgradeContainerRefusesOneCentShort(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Luz", "luz@example.com")
	if _, err := conn.Exec(`UPDATE characters SET money_cents = 1999 WHERE id = ?`, char.ID); err != nil {
		t.Fatalf("prepare character: %v", err)
	}
	char.MoneyCents = 1999

	if _, err := UpgradeContainer(conn, char); !errors.Is(err, ErrNotEnoughMoney) {
		t.Fatalf("expected ErrNotEnoughMoney, got %v", err)
	}
}
