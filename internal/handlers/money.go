package handlers

import (
	"database/sql"
	"fmt"
)

// moneyOf reads the current balance inside a transaction. Handlers carry a
// Character loaded before the transaction; never book from its MoneyCents.
func moneyOf(tx *sql.Tx, characterID int64) (int, error) {
	var cents int
	err := tx.QueryRow(`SELECT money_cents FROM characters WHERE id = ?`, characterID).Scan(&cents)
	return cents, err
}

// postMoney is the one way money moves: a relative UPDATE (never a stale
// absolute value, so overlapping requests cannot overwrite each other) plus the
// ledger row, in the caller's transaction. A debit that would go below zero
// returns ErrNotEnoughMoney and books nothing. Returns the new balance.
func postMoney(tx *sql.Tx, characterID int64, delta int, reason string, refID any) (int, error) {
	var balance int
	err := tx.QueryRow(
		`UPDATE characters SET money_cents = money_cents + ?1
		 WHERE id = ?2 AND money_cents + ?1 >= 0 RETURNING money_cents`,
		delta, characterID,
	).Scan(&balance)
	if err == sql.ErrNoRows {
		return 0, ErrNotEnoughMoney
	}
	if err != nil {
		return 0, fmt.Errorf("post money: %w", err)
	}
	if _, err := tx.Exec(
		`INSERT INTO ledger (character_id, delta_cents, balance_after, reason, ref_id)
		 VALUES (?, ?, ?, ?, ?)`,
		characterID, delta, balance, reason, refID,
	); err != nil {
		return 0, fmt.Errorf("post ledger: %w", err)
	}
	return balance, nil
}
