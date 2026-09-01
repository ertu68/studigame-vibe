package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"studigame/internal/game"
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
	// Retrieve the item from the database
	var item ShopItem
	err := db.QueryRow(`
		SELECT id, name, description, price_cents, category, effect_type, effect_value, max_quantity
		FROM shop_items
		WHERE id = ?
	`, itemID).Scan(&item.ID, &item.Name, &item.Description, &item.PriceCents,
		&item.Category, &item.EffectType, &item.EffectValue, &item.MaxQuantity)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return char, ErrItemNotFound
		}
		return char, err
	}

	// Check if character already owns this item and quantity limit
	var existingQuantity int
	err = db.QueryRow(`
		SELECT quantity FROM character_items 
		WHERE character_id = ? AND item_id = ?
	`, char.ID, itemID).Scan(&existingQuantity)

	// If no rows, existingQuantity remains 0 (not owned yet)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return char, err
	}

	// Check if already at max quantity
	if existingQuantity >= item.MaxQuantity {
		return char, fmt.Errorf("%w: already own maximum quantity of %s", ErrCannotPurchase, item.Name)
	}

	// Check if character has enough money
	if char.MoneyCents < item.PriceCents {
		return char, ErrNotEnoughMoney
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return char, err
	}
	defer tx.Rollback()

	// Deduct money from character
	newBalance := char.MoneyCents - item.PriceCents
	_, err = tx.Exec(`
		UPDATE characters SET money_cents = ? WHERE id = ?
	`, newBalance, char.ID)
	if err != nil {
		return char, err
	}

	// Add item to character's inventory or update quantity
	if existingQuantity > 0 {
		// Update existing item quantity
		_, err = tx.Exec(`
			UPDATE character_items 
			SET quantity = quantity + 1 
			WHERE character_id = ? AND item_id = ?
		`, char.ID, itemID)
	} else {
		// Insert new item
		_, err = tx.Exec(`
			INSERT INTO character_items (character_id, item_id, quantity)
			VALUES (?, ?, 1)
		`, char.ID, itemID)
	}

	if err != nil {
		return char, err
	}

	// Record the purchase in the ledger
	_, err = tx.Exec(`
		INSERT INTO ledger (character_id, delta_cents, balance_after, reason, ref_id)
		VALUES (?, ?, ?, 'shop_purchase', ?)
	`, char.ID, -item.PriceCents, newBalance, itemID)
	if err != nil {
		return char, err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return char, err
	}

	// Update character money and return
	char.MoneyCents = newBalance
	return char, nil
}

// GetShopItemEffects calculates the total effects from all items owned by a character.
func GetShopItemEffects(db *sql.DB, characterID int64) (passiveIncome, containerBonus, moodBonus int, err error) {
	// Get all shop items for reference
	shopItems, err := GetAllShopItems(db)
	if err != nil {
		return 0, 0, 0, err
	}

	// Get character's items
	characterItems, err := GetCharacterItems(db, characterID)
	if err != nil {
		return 0, 0, 0, err
	}

	// Convert to game models for calculation
	var gameCharItems []game.CharacterItem
	for _, item := range characterItems {
		gameCharItems = append(gameCharItems, game.CharacterItem{
			ID:          int(item.ID),
			CharacterID: int(item.CharacterID),
			ItemID:      item.ItemID,
			Quantity:    item.Quantity,
		})
	}

	var gameShopItems []game.ShopItem
	for _, item := range shopItems {
		gameShopItems = append(gameShopItems, game.ShopItem{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			PriceCents:  item.PriceCents,
			Category:    item.Category,
			EffectType:  item.EffectType,
			EffectValue: item.EffectValue,
			MaxQuantity: item.MaxQuantity,
		})
	}

	// Calculate effects using game logic
	passiveIncome = game.PassiveIncome(gameCharItems, gameShopItems)
	containerBonus = game.ContainerBonus(gameCharItems, gameShopItems)
	moodBonus = game.MoodBonus(gameCharItems, gameShopItems)

	return passiveIncome, containerBonus, moodBonus, nil
}

// ErrItemNotFound is returned when trying to purchase an item that doesn't exist.
var ErrItemNotFound = errors.New("item not found")

// ErrCannotPurchase is returned when a purchase cannot be completed.
var ErrCannotPurchase = errors.New("cannot purchase item")
