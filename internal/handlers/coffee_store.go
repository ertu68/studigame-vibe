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
	if char.MoneyCents < game.CoffeeCostCents {
		return char, ErrNotEnoughMoney
	}

	until := now.Add(game.CoffeeBoostWindow)
	if char.CaffeineUntil != nil && char.CaffeineUntil.After(now) {
		until = char.CaffeineUntil.Add(game.CoffeeBoostWindow)
	}
	balance := char.MoneyCents - game.CoffeeCostCents

	tx, err := db.Begin()
	if err != nil {
		return char, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`UPDATE characters SET money_cents = ?, caffeine_until = ? WHERE id = ?`,
		balance, until.UTC().Format(time.RFC3339Nano), char.ID,
	); err != nil {
		return char, err
	}
	if _, err := tx.Exec(
		`INSERT INTO ledger (character_id, delta_cents, balance_after, reason, ref_id)
		 VALUES (?, ?, ?, 'coffee', ?)`,
		char.ID, -game.CoffeeCostCents, balance, char.ID,
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
