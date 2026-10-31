package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"studigame/internal/web"
)

// Shop shows the shop interface with all available items.
func (d *Deps) Shop(u *user, w http.ResponseWriter, r *http.Request) {
	char, err := AnyCharacterForUser(d.DB, u.ID)
	if errors.Is(err, ErrCharacterNotFound) {
		http.Redirect(w, r, "/character/new", http.StatusSeeOther)
		return
	}
	if err != nil {
		slog.Error("Failed to load character", "error", err, "userID", u.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	// Get all shop items
	items, err := GetAllShopItems(d.DB)
	if err != nil {
		slog.Error("Failed to load shop items", "error", err)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	// Get character's owned items
	characterItems, err := GetCharacterItems(d.DB, char.ID)
	if err != nil {
		slog.Error("Failed to load character items", "error", err, "characterID", char.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	// Create a map of owned items for easy lookup
	ownedItems := make(map[int]CharacterItem)
	for _, item := range characterItems {
		ownedItems[item.ItemID] = item
	}

	// Calculate effects from owned items
	passiveIncome, containerBonus, moodBonus, err := GetShopItemEffects(d.DB, char.ID)
	if err != nil {
		slog.Error("Failed to calculate shop item effects", "error", err, "characterID", char.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	// Group items by category for display
	equipmentItems := make([]ShopItem, 0)
	instrumentItems := make([]ShopItem, 0)
	housingItems := make([]ShopItem, 0)

	for _, item := range items {
		switch item.Category {
		case "equipment":
			equipmentItems = append(equipmentItems, item)
		case "instrument":
			instrumentItems = append(instrumentItems, item)
		case "housing":
			housingItems = append(housingItems, item)
		}
	}

	// Format the values for display
	formattedPassiveIncome := web.FormatCents(passiveIncome)
	formattedContainerBonus := web.FormatCents(containerBonus)

	data := d.baseData(w, r)
	d.addCharacterStatus(data, char)
	data["Character"] = char
	data["Money"] = web.FormatCents(char.MoneyCents)
	data["ShopItems"] = items
	data["OwnedItems"] = ownedItems
	data["PassiveIncome"] = formattedPassiveIncome
	data["ContainerBonus"] = formattedContainerBonus
	data["MoodBonus"] = moodBonus
	data["Error"] = r.URL.Query().Get("error")

	slog.Info("Shop data",
		"characterID", char.ID,
		"money", char.MoneyCents,
		"totalItems", len(items),
		"equipmentItems", len(equipmentItems),
		"instrumentItems", len(instrumentItems),
		"housingItems", len(housingItems),
		"ownedItems", len(ownedItems),
		"passiveIncome", passiveIncome,
		"containerBonus", containerBonus,
		"moodBonus", moodBonus,
		"csrfToken", data["CSRFToken"])

	data["EquipmentItems"] = equipmentItems
	data["InstrumentItems"] = instrumentItems
	data["HousingItems"] = housingItems

	web.Render(w, "shop.html", data)
}

// Purchase handles the purchase of an item from the shop.
func (d *Deps) Purchase(u *user, w http.ResponseWriter, r *http.Request) {
	char, err := AnyCharacterForUser(d.DB, u.ID)
	if errors.Is(err, ErrCharacterNotFound) {
		http.Redirect(w, r, "/character/new", http.StatusSeeOther)
		return
	}
	if err != nil {
		slog.Error("Failed to load character", "error", err, "userID", u.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	// Get item ID from form
	itemID, err := strconv.Atoi(r.FormValue("item_id"))
	if err != nil {
		http.Redirect(w, r, "/shop?error=Ungültige+Artikel-ID", http.StatusSeeOther)
		return
	}

	// Attempt to purchase the item
	updatedChar, err := PurchaseItem(d.DB, char, itemID)
	if err != nil {
		var errorMsg string
		switch {
		case errors.Is(err, ErrNotEnoughMoney):
			errorMsg = "Nicht genug Geld für diesen Artikel"
		case errors.Is(err, ErrItemNotFound):
			errorMsg = "Artikel nicht gefunden"
		case errors.Is(err, ErrCannotPurchase):
			errorMsg = err.Error()
		default:
			slog.Error("Failed to purchase item", "error", err, "characterID", char.ID, "itemID", itemID)
			errorMsg = "Interner Fehler beim Kauf"
		}
		http.Redirect(w, r, "/shop?error="+errorMsg, http.StatusSeeOther)
		return
	}

	slog.Info("Item purchased", "characterID", updatedChar.ID, "itemID", itemID)
	http.Redirect(w, r, "/shop", http.StatusSeeOther)
}
