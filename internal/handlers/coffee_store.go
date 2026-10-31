package handlers

import (
	"database/sql"
	"time"

	"studigame/internal/game"
)

// BuyCoffee buys one coffee from the Campus-Kiosk: debits the price and
// extends the concentration boost (the "caffeine_until" window). Buying while
// a boost is active stacks it, per the migration comment. Booking and the
// boost pointer move in one transaction so money and effect can never
// disagree. Returns the refreshed character.
func BuyCoffee(db *sql.DB, char Character, now time.Time) (Character, error) {
	tx, err := db.Begin()
	if err != nil {
		return char, err
	}
	defer tx.Rollback()

	// Read the boost inside the transaction so two quick coffees both stack.
	var current sql.NullString
	if err := tx.QueryRow(`SELECT caffeine_until FROM characters WHERE id = ?`, char.ID).Scan(&current); err != nil {
		return char, err
	}
	until := now.Add(game.CoffeeBoostWindow)
	if t := parseNextAt(current); t != nil && t.After(now) {
		until = t.Add(game.CoffeeBoostWindow)
	}

	balance, err := postMoney(tx, char.ID, -game.CoffeeCostCents, "coffee", char.ID)
	if err != nil {
		return char, err
	}
	if _, err := tx.Exec(
		`UPDATE characters SET caffeine_until = ? WHERE id = ?`,
		until.UTC().Format(time.RFC3339Nano), char.ID,
	); err != nil {
		return char, err
	}
	if err := tx.Commit(); err != nil {
		return char, err
	}

	char.MoneyCents = balance
	char.CaffeineUntil = &until
	return char, nil
}

// CoffeeBoostActive reports whether a student is currently caffeinated.
func CoffeeBoostActive(char Character, now time.Time) bool {
	return char.CaffeineUntil != nil && game.CoffeeBoostActive(*char.CaffeineUntil, now)
}
