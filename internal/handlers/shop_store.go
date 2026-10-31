package handlers

import (
	"database/sql"
	"errors"
	"fmt"
)

// ShopItem represents an item available in the shop (DB model).
type ShopItem struct {
	ID          int
	Name        string
	Description string
	PriceCents  int
	Category    string
	EffectType  string
	EffectValue int
	MaxQuantity int
}

// CharacterItem represents an item owned by a character (DB model).
type CharacterItem struct {
	ID          int64
	CharacterID int64
	ItemID      int
	Quantity    int
}

// GetAllShopItems retrieves all items available in the shop.
func GetAllShopItems(db *sql.DB) ([]ShopItem, error) {
	rows, err := db.Query(`
		SELECT id, name, description, price_cents, category, effect_type, effect_value, max_quantity
		FROM shop_items
		ORDER BY category, price_cents
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ShopItem
	for rows.Next() {
		var item ShopItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.PriceCents,
			&item.Category, &item.EffectType, &item.EffectValue, &item.MaxQuantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// GetCharacterItems retrieves all items owned by a character.
func GetCharacterItems(db *sql.DB, characterID int64) ([]CharacterItem, error) {
	rows, err := db.Query(`
		SELECT id, character_id, item_id, quantity
		FROM character_items
		WHERE character_id = ?
	`, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CharacterItem
	for rows.Next() {
		var item CharacterItem
		if err := rows.Scan(&item.ID, &item.CharacterID, &item.ItemID, &item.Quantity); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// PurchaseItem attempts to purchase an item for a character.
// It deducts the price from the character's money and adds the item to their inventory.
func PurchaseItem(db *sql.DB, char Character, itemID int) (Character, error) {
	var item ShopItem
	err := db.QueryRow(`
		SELECT id, name, description, price_cents, category, effect_type, effect_value, max_quantity
		FROM shop_items WHERE id = ?`, itemID,
	).Scan(&item.ID, &item.Name, &item.Description, &item.PriceCents,
		&item.Category, &item.EffectType, &item.EffectValue, &item.MaxQuantity)
	if errors.Is(err, sql.ErrNoRows) {
		return char, ErrItemNotFound
	}
	if err != nil {
		return char, err
	}

	tx, err := db.Begin()
	if err != nil {
		return char, err
	}
	defer tx.Rollback()

	// Quantity is checked inside the transaction so a double click cannot
	// exceed max_quantity.
	var owned int
	err = tx.QueryRow(
		`SELECT quantity FROM character_items WHERE character_id = ? AND item_id = ?`,
		char.ID, itemID,
	).Scan(&owned)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return char, err
	}
	if owned >= item.MaxQuantity {
		return char, fmt.Errorf("%w: already own maximum quantity of %s", ErrCannotPurchase, item.Name)
	}

	balance, err := postMoney(tx, char.ID, -item.PriceCents, "shop_purchase", itemID)
	if err != nil {
		return char, err
	}
	if _, err := tx.Exec(
		`INSERT INTO character_items (character_id, item_id, quantity) VALUES (?, ?, 1)
		 ON CONFLICT(character_id, item_id) DO UPDATE SET quantity = quantity + 1`,
		char.ID, itemID,
	); err != nil {
		return char, err
	}
	if err := tx.Commit(); err != nil {
		return char, err
	}

	char.MoneyCents = balance
	return char, nil
}

// ErrItemNotFound is returned when trying to purchase an item that doesn't exist.
var ErrItemNotFound = errors.New("item not found")

// ErrCannotPurchase is returned when a purchase cannot be completed.
var ErrCannotPurchase = errors.New("cannot purchase item")
