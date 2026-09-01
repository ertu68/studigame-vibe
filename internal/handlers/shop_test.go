package handlers

import (
	"testing"
)

func TestShopItems(t *testing.T) {
	db := testDB(t)

	// Debug: Check if the table exists
	var tableCount int
	err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='shop_items'`).Scan(&tableCount)
	if err != nil {
		t.Fatalf("check table existence: %v", err)
	}
	if tableCount == 0 {
		t.Fatal("shop_items table does not exist")
	}

	// Insert test shop items with IDs that don't conflict with the migration
	_, err = db.Exec(`
		INSERT INTO shop_items (id, name, description, price_cents, category, effect_type, effect_value, max_quantity)
		VALUES 
		(100, 'Test Item 1', 'Description 1', 100, 'equipment', 'passive_income', 10, 1),
		(101, 'Test Item 2', 'Description 2', 200, 'instrument', 'passive_income', 20, 1),
		(102, 'Test Item 3', 'Description 3', 300, 'housing', 'container_bonus', 1000, 1)
	`)
	if err != nil {
		t.Fatalf("insert shop items: %v", err)
	}

	// Test GetAllShopItems
	items, err := GetAllShopItems(db)
	if err != nil {
		t.Fatalf("GetAllShopItems: %v", err)
	}
	// We expect at least 3 items (the ones we just inserted) plus whatever was inserted by the migration
	if len(items) < 3 {
		t.Fatalf("expected at least 3 items, got %d", len(items))
	}

	// Since GetAllShopItems orders by category and then price, we need to find our items
	var foundItems []ShopItem
	for _, item := range items {
		if item.ID >= 100 && item.ID <= 102 {
			foundItems = append(foundItems, item)
		}
	}

	if len(foundItems) != 3 {
		t.Fatalf("expected to find 3 test items, got %d", len(foundItems))
	}

	// Check that our items are ordered by category and price
	// Since we're inserting them in order, they should appear in the correct positions
	// But we need to account for other items that might be in the database
	foundByID := make(map[int]ShopItem)
	for _, item := range foundItems {
		foundByID[item.ID] = item
	}

	item1 := foundByID[100]
	item2 := foundByID[101]
	item3 := foundByID[102]

	if item1.Name != "Test Item 1" || item1.Category != "equipment" {
		t.Errorf("first item: expected Test Item 1/equipment, got %s/%s", item1.Name, item1.Category)
	}
	if item2.Name != "Test Item 2" || item2.Category != "instrument" {
		t.Errorf("second item: expected Test Item 2/instrument, got %s/%s", item2.Name, item2.Category)
	}
	if item3.Name != "Test Item 3" || item3.Category != "housing" {
		t.Errorf("third item: expected Test Item 3/housing, got %s/%s", item3.Name, item3.Category)
	}
}

func TestCharacterItems(t *testing.T) {
	db := testDB(t)

	// Insert test shop items with IDs that don't conflict with the migration
	_, err := db.Exec(`
		INSERT INTO shop_items (id, name, description, price_cents, category, effect_type, effect_value, max_quantity)
		VALUES (103, 'Test Item', 'Description', 100, 'equipment', 'passive_income', 10, 1)
	`)
	if err != nil {
		t.Fatalf("insert shop items: %v", err)
	}

	// Insert test character
	char := newCharacter(t, db, "Test User", "test@example.com")

	// Initially, character should have no items
	items, err := GetCharacterItems(db, char.ID)
	if err != nil {
		t.Fatalf("GetCharacterItems: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}

	// Purchase an item
	char.MoneyCents = 500
	updatedChar, err := PurchaseItem(db, char, 103)
	if err != nil {
		t.Fatalf("PurchaseItem: %v", err)
	}
	if updatedChar.MoneyCents != 400 {
		t.Errorf("expected 400 money cents, got %d", updatedChar.MoneyCents)
	}

	// Character should now have the item
	items, err = GetCharacterItems(db, char.ID)
	if err != nil {
		t.Fatalf("GetCharacterItems after purchase: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ItemID != 103 {
		t.Errorf("expected item ID 103, got %d", items[0].ItemID)
	}
	if items[0].Quantity != 1 {
		t.Errorf("expected quantity 1, got %d", items[0].Quantity)
	}
}

func TestPurchaseItem(t *testing.T) {
	db := testDB(t)

	// Insert test shop items with IDs that don't conflict with the migration
	_, err := db.Exec(`
		INSERT INTO shop_items (id, name, description, price_cents, category, effect_type, effect_value, max_quantity)
		VALUES 
		(104, 'Expensive Item', 'Description', 1000, 'equipment', 'passive_income', 10, 1),
		(105, 'Cheap Item', 'Description', 100, 'instrument', 'passive_income', 20, 2)
	`)
	if err != nil {
		t.Fatalf("insert shop items: %v", err)
	}

	char := newCharacter(t, db, "Test User", "test@example.com")
	char.MoneyCents = 500

	// Try to purchase an item that's too expensive
	_, err = PurchaseItem(db, char, 104)
	if err != ErrNotEnoughMoney {
		t.Errorf("expected ErrNotEnoughMoney, got %v", err)
	}

	// Purchase a cheap item
	updatedChar, err := PurchaseItem(db, char, 105)
	if err != nil {
		t.Fatalf("PurchaseItem cheap item: %v", err)
	}
	if updatedChar.MoneyCents != 400 {
		t.Errorf("expected 400 money cents, got %d", updatedChar.MoneyCents)
	}

	// Try to purchase the same item again (should work since max_quantity is 2)
	updatedChar, err = PurchaseItem(db, updatedChar, 105)
	if err != nil {
		t.Fatalf("PurchaseItem second cheap item: %v", err)
	}
	if updatedChar.MoneyCents != 300 {
		t.Errorf("expected 300 money cents, got %d", updatedChar.MoneyCents)
	}

	// Try to purchase the same item again (should fail since we've reached max_quantity)
	_, err = PurchaseItem(db, updatedChar, 105)
	if err == nil {
		t.Error("expected error when exceeding max_quantity")
	}
}

func TestGetShopItemEffects(t *testing.T) {
	db := testDB(t)

	// Insert test shop items with IDs that don't conflict with the migration
	_, err := db.Exec(`
		INSERT INTO shop_items (id, name, description, price_cents, category, effect_type, effect_value, max_quantity)
		VALUES 
		(106, 'Income Item', 'Description', 100, 'equipment', 'passive_income', 10, 1),
		(107, 'Container Item', 'Description', 200, 'instrument', 'container_bonus', 1000, 1),
		(108, 'Mood Item', 'Description', 300, 'housing', 'mood_bonus', 5, 1)
	`)
	if err != nil {
		t.Fatalf("insert shop items: %v", err)
	}

	char := newCharacter(t, db, "Test User", "test@example.com")

	// Initially, character should have no effects
	passiveIncome, containerBonus, moodBonus, err := GetShopItemEffects(db, char.ID)
	if err != nil {
		t.Fatalf("GetShopItemEffects: %v", err)
	}
	if passiveIncome != 0 {
		t.Errorf("expected 0 passive income, got %d", passiveIncome)
	}
	if containerBonus != 0 {
		t.Errorf("expected 0 container bonus, got %d", containerBonus)
	}
	if moodBonus != 0 {
		t.Errorf("expected 0 mood bonus, got %d", moodBonus)
	}

	// Purchase items
	char.MoneyCents = 1000
	for _, itemID := range []int{106, 107, 108} {
		char, err = PurchaseItem(db, char, itemID)
		if err != nil {
			t.Fatalf("PurchaseItem %d: %v", itemID, err)
		}
	}

	// Character should now have effects
	passiveIncome, containerBonus, moodBonus, err = GetShopItemEffects(db, char.ID)
	if err != nil {
		t.Fatalf("GetShopItemEffects after purchase: %v", err)
	}
	if passiveIncome != 10 {
		t.Errorf("expected 10 passive income, got %d", passiveIncome)
	}
	if containerBonus != 1000 {
		t.Errorf("expected 1000 container bonus, got %d", containerBonus)
	}
	if moodBonus != 5 {
		t.Errorf("expected 5 mood bonus, got %d", moodBonus)
	}
}
