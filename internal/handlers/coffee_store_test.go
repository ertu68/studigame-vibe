package handlers

import (
	"testing"
	"time"

	"studigame/internal/game"
)

func TestBuyCoffeeDebitsAndStartsBoost(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Nadja", "nadja@example.com")
	now := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	start := char.MoneyCents
	bought, err := BuyCoffee(conn, char, now)
	if err != nil {
		t.Fatalf("buy: %v", err)
	}
	if got := bought.MoneyCents; got != start-game.CoffeeCostCents {
		t.Fatalf("money: got %d, want %d", got, start-game.CoffeeCostCents)
	}
	want := now.Add(game.CoffeeBoostWindow)
	if bought.CaffeineUntil == nil || !bought.CaffeineUntil.Equal(want) {
		t.Fatalf("window: got %v, want %v", bought.CaffeineUntil, want)
	}
	if !CoffeeBoostActive(bought, now.Add(time.Hour)) {
		t.Fatal("boost should be active inside the window")
	}
}

func TestBuyCoffeeStacksActiveBoost(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Nadja", "nadja@example.com")
	now := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	first, err := BuyCoffee(conn, char, now)
	if err != nil {
		t.Fatalf("first buy: %v", err)
	}
	second, err := BuyCoffee(conn, first, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("second buy: %v", err)
	}
	want := now.Add(2 * game.CoffeeBoostWindow)
	if second.CaffeineUntil == nil || !second.CaffeineUntil.Equal(want) {
		t.Fatalf("stacked window: got %v, want %v", second.CaffeineUntil, want)
	}
}

func TestBuyCoffeeFailsWithoutMoney(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Nadja", "nadja@example.com")
	now := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	if _, err := conn.Exec(`UPDATE characters SET money_cents = 0 WHERE id = ?`, char.ID); err != nil {
		t.Fatalf("zero money: %v", err)
	}
	char.MoneyCents = 0

	got, err := BuyCoffee(conn, char, now)
	if err != ErrNotEnoughMoney {
		t.Fatalf("expected ErrNotEnoughMoney, got %v", err)
	}
	if got.MoneyCents != 0 || got.CaffeineUntil != nil {
		t.Fatal("failed buy must not change money or boost")
	}
}

func TestCoffeeBoostActiveExpiry(t *testing.T) {
	conn := testDB(t)
	char := newCharacter(t, conn, "Nadja", "nadja@example.com")
	now := time.Date(2026, 8, 26, 8, 0, 0, 0, time.UTC)

	bought, err := BuyCoffee(conn, char, now)
	if err != nil {
		t.Fatalf("buy: %v", err)
	}
	if CoffeeBoostActive(bought, *bought.CaffeineUntil) {
		t.Fatal("boost must be off exactly at the window end")
	}
	if CoffeeBoostActive(bought, bought.CaffeineUntil.Add(time.Minute)) {
		t.Fatal("boost must be off after the window")
	}
	if CoffeeBoostActive(char, now) {
		t.Fatal("a character without coffee must have no boost")
	}
}
