package handlers

import (
	"testing"
	"time"

	"studigame/internal/game"
)

func TestSetBafogEnableDisable(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Nadja", "nadja@example.com")
	now := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	if char.BafogNextAt != nil {
		t.Fatalf("new character must not have BAföG applied")
	}

	on, err := SetBafog(conn, char, true, now)
	if err != nil {
		t.Fatalf("enable: %v", err)
	}
	if on.BafogNextAt == nil {
		t.Fatal("enable did not set a next payout")
	}
	want := now
	if !on.BafogNextAt.Equal(want) {
		t.Fatalf("payout anchor: got %v, want %v", on.BafogNextAt, want)
	}

	// No-op when already enabled.
	again, err := SetBafog(conn, on, true, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("re-enable: %v", err)
	}
	if again.BafogNextAt == nil || !again.BafogNextAt.Equal(want) {
		t.Fatalf("re-enable moved the pointer: %v", again.BafogNextAt)
	}

	off, err := SetBafog(conn, on, false, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("disable: %v", err)
	}
	if off.BafogNextAt != nil {
		t.Fatal("disable did not clear the next payout")
	}
}

func TestCollectBafogPaysOutWhenDue(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Timo", "timo@example.com")
	now := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	// Jutebeutel (5000) so the 30 € payout fits without overflowing.
	if _, err := conn.Exec(`UPDATE characters SET container_level = 1 WHERE id = ?`, char.ID); err != nil {
		t.Fatalf("set container level: %v", err)
	}
	char.ContainerLevel = 1

	on, err := SetBafog(conn, char, true, now)
	if err != nil {
		t.Fatalf("enable: %v", err)
	}

	// Nothing is due before the first period passes.
	summary, err := CollectBafog(conn, on, now.Add(game.BafogPeriod-1*time.Second))
	if err != nil {
		t.Fatalf("collect early: %v", err)
	}
	if summary != nil {
		t.Fatalf("early collect returned a summary: %+v", summary)
	}

	due := now.Add(game.BafogPeriod)
	summary, err = CollectBafog(conn, on, due)
	if err != nil {
		t.Fatalf("collect due: %v", err)
	}
	if summary == nil {
		t.Fatal("expected a summary when a payout is due")
	}
	if summary.Title != "BAföG eingegangen" {
		t.Fatalf("title: got %q, want BAföG eingegangen", summary.Title)
	}

	reloaded, err := CharacterForUser(conn, char.UserID, char.EditionID)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	// 10 € starting capital + 30 € payout, container not full.
	if got, want := reloaded.MoneyCents, 1000+game.BafogPayoutCents; got != want {
		t.Fatalf("money after payout: got %d, want %d", got, want)
	}
	// Pointer advanced one period along the anchored cadence.
	if got, want := reloaded.BafogNextAt, now.Add(game.BafogPeriod); got == nil || !got.Equal(want) {
		t.Fatalf("next payout after collect: got %v, want %v", got, want)
	}
}

func TestCollectBafogDisabledOrNotDueReturnsNil(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Lea", "lea@example.com")
	now := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	// Never applied.
	s, err := CollectBafog(conn, char, now.Add(10*time.Hour))
	if err != nil {
		t.Fatalf("collect without application: %v", err)
	}
	if s != nil {
		t.Fatalf("unapplied collect returned a summary: %+v", s)
	}

	// Applied but then disabled.
	on, _ := SetBafog(conn, char, true, now)
	off, _ := SetBafog(conn, on, false, now)
	s, err = CollectBafog(conn, off, now.Add(game.BafogPeriod))
	if err != nil {
		t.Fatalf("collect after cancel: %v", err)
	}
	if s != nil {
		t.Fatalf("cancelled collect returned a summary: %+v", s)
	}
}

func TestCollectBafogClampsAtMaxPeriods(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Juri", "juri@example.com")
	now := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	// Rucksack (20000) so 4 × 30 € fits without overflowing.
	if _, err := conn.Exec(`UPDATE characters SET container_level = 2 WHERE id = ?`, char.ID); err != nil {
		t.Fatalf("set container level: %v", err)
	}
	char.ContainerLevel = 2

	on, err := SetBafog(conn, char, true, now)
	if err != nil {
		t.Fatalf("enable: %v", err)
	}

	// Years later: dozens of periods due, but the clamp kicks in.
	late := on.BafogNextAt.Add(365 * 24 * time.Hour)
	summary, err := CollectBafog(conn, on, late)
	if err != nil {
		t.Fatalf("collect late: %v", err)
	}
	if summary == nil {
		t.Fatal("expected a summary after a long absence")
	}

	reloaded, _ := CharacterForUser(conn, char.UserID, char.EditionID)
	if got, want := reloaded.MoneyCents, 1000+game.MaxBafogPeriods*game.BafogPayoutCents; got != want {
		t.Fatalf("money after clamp: got %d, want %d (clamped to %d periods)",
			got, want, game.MaxBafogPeriods)
	}
	// Pointer advances only the clamped number of periods, not to now.
	wantNext := on.BafogNextAt.Add(time.Duration(game.MaxBafogPeriods) * game.BafogPeriod)
	if got := reloaded.BafogNextAt; got == nil || !got.Equal(wantNext) {
		t.Fatalf("next payout after clamp: got %v, want %v", got, wantNext)
	}
}

func TestCollectBafogOverflowGoesToCharity(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Pia", "pia@example.com")
	now := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	// Fill the container so the payout overflows.
	if _, err := conn.Exec(`UPDATE characters SET money_cents = ? WHERE id = ?`,
		game.ContainerCapacity(0), char.ID); err != nil {
		t.Fatalf("fill container: %v", err)
	}

	on, err := SetBafog(conn, char, true, now)
	if err != nil {
		t.Fatalf("enable: %v", err)
	}

	summary, err := CollectBafog(conn, on, now.Add(game.BafogPeriod))
	if err != nil {
		t.Fatalf("collect full: %v", err)
	}
	if summary == nil {
		t.Fatal("expected a summary")
	}
	if !summary.HasDonation {
		t.Fatalf("expected a donation flag when the container overflows")
	}
	if summary.Donated == "" {
		t.Fatal("expected a donated amount")
	}

	reloaded, _ := CharacterForUser(conn, char.UserID, char.EditionID)
	if got, want := reloaded.MoneyCents, game.ContainerCapacity(0); got != want {
		t.Fatalf("money must stay at capacity: got %d, want %d", got, want)
	}

	var pool int
	if err := conn.QueryRow(`SELECT cents FROM charity_pool WHERE edition_id = ?`, char.EditionID).Scan(&pool); err != nil {
		t.Fatalf("read charity pool: %v", err)
	}
	if pool <= 0 {
		t.Fatalf("charity pool empty after overflow, got %d", pool)
	}
}

func TestCollectBafogWritesLedgerRow(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Emre", "emre@example.com")
	now := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	// Jutebeutel (5000) so the 30 € payout fits without overflowing.
	if _, err := conn.Exec(`UPDATE characters SET container_level = 1 WHERE id = ?`, char.ID); err != nil {
		t.Fatalf("set container level: %v", err)
	}
	char.ContainerLevel = 1

	on, _ := SetBafog(conn, char, true, now)
	if _, err := CollectBafog(conn, on, now.Add(game.BafogPeriod)); err != nil {
		t.Fatalf("collect: %v", err)
	}

	var reason string
	var delta int
	if err := conn.QueryRow(
		`SELECT reason, delta_cents FROM ledger WHERE character_id = ? ORDER BY id DESC LIMIT 1`,
		char.ID).Scan(&reason, &delta); err != nil {
		t.Fatalf("read ledger: %v", err)
	}
	if reason != "bafog" {
		t.Fatalf("ledger reason: got %q, want bafog", reason)
	}
	if delta != game.BafogPayoutCents {
		t.Fatalf("ledger delta: got %d, want %d", delta, game.BafogPayoutCents)
	}
}
