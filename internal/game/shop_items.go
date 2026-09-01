// Package game implements all game rules that do not depend on infrastructure:
// no database access, no wall clock, no random source (these are injected).
package game

// ShopItem represents an item available in the shop.
type ShopItem struct {
	ID          int
	Name        string
	Description string
	PriceCents  int
	Category    string // "equipment", "instrument", "housing"
	EffectType  string // "passive_income", "mood_bonus", "container_bonus"
	EffectValue int    // Value of the effect (e.g., income per hour in cents)
	MaxQuantity int    // How many of this item a character can own
}

// CharacterItem represents an item owned by a character.
type CharacterItem struct {
	ID          int
	CharacterID int
	ItemID      int
	Quantity    int
}

// PassiveIncome returns the total passive income in cents per hour from owned items.
func PassiveIncome(items []CharacterItem, shopItems []ShopItem) int {
	income := 0
	itemMap := make(map[int]ShopItem)

	// Create a map for quick lookup
	for _, item := range shopItems {
		itemMap[item.ID] = item
	}

	// Calculate total passive income
	for _, charItem := range items {
		if shopItem, exists := itemMap[charItem.ItemID]; exists {
			if shopItem.EffectType == "passive_income" {
				income += shopItem.EffectValue * charItem.Quantity
			}
		}
	}

	return income
}

// ContainerBonus returns the total container capacity bonus from owned items.
func ContainerBonus(items []CharacterItem, shopItems []ShopItem) int {
	bonus := 0
	itemMap := make(map[int]ShopItem)

	// Create a map for quick lookup
	for _, item := range shopItems {
		itemMap[item.ID] = item
	}

	// Calculate total container bonus
	for _, charItem := range items {
		if shopItem, exists := itemMap[charItem.ItemID]; exists {
			if shopItem.EffectType == "container_bonus" {
				bonus += shopItem.EffectValue * charItem.Quantity
			}
		}
	}

	return bonus
}

// MoodBonus returns the total mood bonus from owned items.
func MoodBonus(items []CharacterItem, shopItems []ShopItem) int {
	bonus := 0
	itemMap := make(map[int]ShopItem)

	// Create a map for quick lookup
	for _, item := range shopItems {
		itemMap[item.ID] = item
	}

	// Calculate total mood bonus
	for _, charItem := range items {
		if shopItem, exists := itemMap[charItem.ItemID]; exists {
			if shopItem.EffectType == "mood_bonus" {
				bonus += shopItem.EffectValue * charItem.Quantity
			}
		}
	}

	return bonus
}

// CanPurchase checks if a character can purchase an item.
func CanPurchase(charMoneyCents int, item ShopItem, ownedItems []CharacterItem) bool {
	// Check if character has enough money
	if charMoneyCents < item.PriceCents {
		return false
	}

	// Check if character already owns maximum quantity
	for _, owned := range ownedItems {
		if owned.ItemID == item.ID {
			if owned.Quantity >= item.MaxQuantity {
				return false
			}
			break
		}
	}

	return true
}
