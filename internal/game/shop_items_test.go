package game

import (
	"testing"
)

func TestPassiveIncome(t *testing.T) {
	shopItems := []ShopItem{
		{ID: 1, Name: "Guitar", EffectType: "passive_income", EffectValue: 50},
		{ID: 2, Name: "Keyboard", EffectType: "passive_income", EffectValue: 75},
		{ID: 3, Name: "Book", EffectType: "mood_bonus", EffectValue: 5},
	}

	characterItems := []CharacterItem{
		{ItemID: 1, Quantity: 1},
		{ItemID: 2, Quantity: 2},
		{ItemID: 3, Quantity: 1},
	}

	expected := 50*1 + 75*2 // 50 + 150 = 200
	actual := PassiveIncome(characterItems, shopItems)

	if actual != expected {
		t.Errorf("PassiveIncome() = %d; want %d", actual, expected)
	}
}

func TestContainerBonus(t *testing.T) {
	shopItems := []ShopItem{
		{ID: 1, Name: "Backpack", EffectType: "container_bonus", EffectValue: 1000},
		{ID: 2, Name: "Bike", EffectType: "container_bonus", EffectValue: 2000},
		{ID: 3, Name: "Book", EffectType: "mood_bonus", EffectValue: 5},
	}

	characterItems := []CharacterItem{
		{ItemID: 1, Quantity: 1},
		{ItemID: 2, Quantity: 1},
		{ItemID: 3, Quantity: 1},
	}

	expected := 1000*1 + 2000*1 // 1000 + 2000 = 3000
	actual := ContainerBonus(characterItems, shopItems)

	if actual != expected {
		t.Errorf("ContainerBonus() = %d; want %d", actual, expected)
	}
}

func TestMoodBonus(t *testing.T) {
	shopItems := []ShopItem{
		{ID: 1, Name: "Book", EffectType: "mood_bonus", EffectValue: 5},
		{ID: 2, Name: "Plant", EffectType: "mood_bonus", EffectValue: 3},
		{ID: 3, Name: "Guitar", EffectType: "passive_income", EffectValue: 50},
	}

	characterItems := []CharacterItem{
		{ItemID: 1, Quantity: 2},
		{ItemID: 2, Quantity: 1},
		{ItemID: 3, Quantity: 1},
	}

	expected := 5*2 + 3*1 // 10 + 3 = 13
	actual := MoodBonus(characterItems, shopItems)

	if actual != expected {
		t.Errorf("MoodBonus() = %d; want %d", actual, expected)
	}
}

func TestCanPurchase(t *testing.T) {
	item := ShopItem{
		ID:          1,
		Name:        "Test Item",
		PriceCents:  1000,
		MaxQuantity: 1,
	}

	ownedItems := []CharacterItem{
		{ItemID: 1, Quantity: 0},
	}

	// Test case 1: Not enough money
	if CanPurchase(500, item, ownedItems) {
		t.Error("Expected CanPurchase to return false when character doesn't have enough money")
	}

	// Test case 2: Enough money, not owned yet
	if !CanPurchase(1500, item, ownedItems) {
		t.Error("Expected CanPurchase to return true when character has enough money and hasn't purchased the item")
	}

	// Test case 3: Already owns maximum quantity
	ownedItems[0].Quantity = 1
	if CanPurchase(1500, item, ownedItems) {
		t.Error("Expected CanPurchase to return false when character already owns maximum quantity")
	}

	// Test case 4: Item with higher max quantity
	item.MaxQuantity = 3
	ownedItems[0].Quantity = 2
	if !CanPurchase(1500, item, ownedItems) {
		t.Error("Expected CanPurchase to return true when character can still purchase more of the item")
	}
}
