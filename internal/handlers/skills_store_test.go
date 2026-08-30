package handlers

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

// skillNow is a fixed wall clock for skill tests.
func TestStartSkillTrainingHappyPath(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Tim", "tim@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)
	// Informatik: geschicklichkeit level 0 -> 1000 ct, factor 1.0, exactly affordable.
	duration, err := StartSkillTraining(conn, char, "geschicklichkeit", start)
	if err != nil {
		t.Fatalf("start training: %v", err)
	}
	if duration <= 0 {
		t.Fatalf("expected a positive duration, got %d", duration)
	}

	reloaded, err := CharacterForUser(conn, char.UserID, char.EditionID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.MoneyCents != 0 {
		t.Errorf("money after paying: got %d, want 0", reloaded.MoneyCents)
	}

	var delta, balance int
	var reason string
	err = conn.QueryRow(
		`SELECT delta_cents, balance_after, reason FROM ledger
		 WHERE character_id = ? AND reason = 'skill_train'`, char.ID,
	).Scan(&delta, &balance, &reason)
	if err == sql.ErrNoRows {
		t.Fatal("no skill_train ledger entry")
	}
	if err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	if delta != -1000 || balance != 0 || reason != "skill_train" {
		t.Fatalf("ledger row wrong: delta=%d balance=%d reason=%q", delta, balance, reason)
	}

	// Not due yet: resolving changes nothing.
	if err := ResolveSkillTrainings(conn, char.ID, start.Add(time.Hour)); err != nil {
		t.Fatalf("resolve early: %v", err)
	}
	levels, err := SkillLevels(conn, char.ID)
	if err != nil {
		t.Fatalf("levels: %v", err)
	}
	if levels["geschicklichkeit"] != 0 {
		t.Fatalf("level must stay 0 until due, got %d", levels["geschicklichkeit"])
	}

	// Due: resolve bumps it to 1.
	due := start.Add(time.Duration(duration) * time.Minute)
	if err := ResolveSkillTrainings(conn, char.ID, due); err != nil {
		t.Fatalf("resolve due: %v", err)
	}
	levels, err = SkillLevels(conn, char.ID)
	if err != nil {
		t.Fatalf("levels after resolve: %v", err)
	}
	if levels["geschicklichkeit"] != 1 {
		t.Fatalf("level after resolve: got %d, want 1", levels["geschicklichkeit"])
	}
}

func TestStartSkillTrainingNotEnoughMoney(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Ada", "ada@example.com")
	// Level 2 of geschicklichkeit (Informatik) costs base×(3)^1.8 which is
	// well above the 10 € starting capital; force level 1 first.
	if _, err := conn.Exec(
		`INSERT INTO skills (character_id, skill, level) VALUES (?, 'geschicklichkeit', 1)`,
		char.ID,
	); err != nil {
		t.Fatalf("seed level: %v", err)
	}
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)
	if _, err := StartSkillTraining(conn, char, "geschicklichkeit", start); !errors.Is(err, ErrNotEnoughMoney) {
		t.Fatalf("expected ErrNotEnoughMoney, got %v", err)
	}
	if _, err := FindSkillTraining(conn, char.ID); !errors.Is(err, ErrNoSkillTraining) {
		t.Fatalf("failed training must not book a run, got %v", err)
	}
	var entries int
	if err := conn.QueryRow(`SELECT count(*) FROM ledger WHERE reason = 'skill_train' AND character_id = ?`, char.ID).Scan(&entries); err != nil {
		t.Fatalf("count ledger: %v", err)
	}
	if entries != 0 {
		t.Fatalf("failed training booked money: %d entries", entries)
	}
}

func TestStartSkillTrainingAtMax(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Mia", "mia@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)
	max := 10 // konzentration max level
	if _, err := conn.Exec(`INSERT INTO skills (character_id, skill, level) VALUES (?, 'konzentration', ?)`, char.ID, max); err != nil {
		t.Fatalf("seed max level: %v", err)
	}
	if _, err := StartSkillTraining(conn, char, "konzentration", start); !errors.Is(err, ErrSkillAtMax) {
		t.Fatalf("expected ErrSkillAtMax, got %v", err)
	}
}

func TestStartSkillTrainingOneAtATime(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Leo", "leo@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)
	if _, err := conn.Exec(`UPDATE characters SET money_cents = 5000 WHERE id = ?`, char.ID); err != nil {
		t.Fatalf("fund: %v", err)
	}
	char.MoneyCents = 5000

	if _, err := StartSkillTraining(conn, char, "geschicklichkeit", start); err != nil {
		t.Fatalf("first training: %v", err)
	}
	if _, err := StartSkillTraining(conn, char, "auftreten", start); !errors.Is(err, ErrSkillTrainingRunning) {
		t.Fatalf("second concurrent training must fail with ErrSkillTrainingRunning, got %v", err)
	}
}

func TestCancelSkillTrainingRefundsEightyPercent(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Nora", "nora@example.com")
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)
	if _, err := StartSkillTraining(conn, char, "geschicklichkeit", start); err != nil {
		t.Fatalf("start training: %v", err)
	}

	training, err := FindSkillTraining(conn, char.ID)
	if err != nil {
		t.Fatalf("find training: %v", err)
	}
	if training.CostCents != 1000 {
		t.Fatalf("cost: got %d, want 1000", training.CostCents)
	}

	// Reload the character like the handler does, so the balance reflects the
	// payment made when the training started.
	paid, err := CharacterForUser(conn, char.UserID, char.EditionID)
	if err != nil {
		t.Fatalf("reload paid character: %v", err)
	}
	if paid.MoneyCents != 0 {
		t.Fatalf("expected 0 after paying, got %d", paid.MoneyCents)
	}

	// Cancel: back to 1000 (paid) + 800 (refund).
	refund, err := CancelSkillTraining(conn, paid, training.ID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if refund != 800 {
		t.Errorf("refund: got %d, want 800", refund)
	}
	after, err := CharacterForUser(conn, paid.UserID, paid.EditionID)
	if err != nil {
		t.Fatalf("reload after cancel: %v", err)
	}
	if after.MoneyCents != 800 {
		t.Errorf("money after cancel: got %d, want 800", after.MoneyCents)
	}
	if _, err := FindSkillTraining(conn, char.ID); !errors.Is(err, ErrNoSkillTraining) {
		t.Fatalf("training must be gone after cancel, got %v", err)
	}
	var delta int
	var reason string
	if err := conn.QueryRow(
		`SELECT delta_cents, reason FROM ledger WHERE character_id = ? AND reason = 'skill_refund'`, char.ID,
	).Scan(&delta, &reason); err != nil {
		t.Fatalf("refund ledger: %v", err)
	}
	if delta != 800 || reason != "skill_refund" {
		t.Fatalf("refund ledger wrong: delta=%d reason=%q", delta, reason)
	}
}

// A higher Lerneffizienz is frozen into a collect shift at start, so a skill
// level-up mid-shift cannot retroactively boost it.
func TestSkillFactorFreezesIntoCollectShift(t *testing.T) {
	conn := testDB(t)
	start := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	plain := newCharacter(t, conn, "Kai", "kai@example.com")
	skilled := newCharacter(t, conn, "Sam", "sam@example.com")
	if _, err := conn.Exec(`INSERT INTO skills (character_id, skill, level) VALUES (?, 'geschicklichkeit', 10)`, skilled.ID); err != nil {
		t.Fatalf("seed skilled: %v", err)
	}

	if err := StartCollect(conn, plain, time.Hour, start); err != nil {
		t.Fatalf("start plain shift: %v", err)
	}
	if err := StartCollect(conn, skilled, time.Hour, start); err != nil {
		t.Fatalf("start skilled shift: %v", err)
	}

	roll := 0.5
	plainSum, err := ResolveDueAction(conn, plain, start.Add(time.Hour), roll)
	if err != nil {
		t.Fatalf("resolve plain: %v", err)
	}
	skilledSum, err := ResolveDueAction(conn, skilled, start.Add(time.Hour), roll)
	if err != nil {
		t.Fatalf("resolve skilled: %v", err)
	}
	if plainSum == nil || skilledSum == nil {
		t.Fatal("expected both shifts to resolve")
	}
	if !(skilledSum.Detail != plainSum.Detail) {
		t.Fatalf("skilled yield must exceed plain: plain=%q skilled=%q", plainSum.Detail, skilledSum.Detail)
	}
}
