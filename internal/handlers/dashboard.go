package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"studigame/internal/game"
	"studigame/internal/web"
)

func (d *Deps) Dashboard(u *user, w http.ResponseWriter, r *http.Request) {
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

	data := d.baseData(w, r)
	d.addCharacterStatus(data, char)
	data["Character"] = char
	addUpgradeData(data, char)
	web.Render(w, "dashboard.html", data)
}

// ContainerUpgrade buys the next container tier and returns to the dashboard.
func (d *Deps) ContainerUpgrade(u *user, w http.ResponseWriter, r *http.Request) {
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

	refreshed, err := UpgradeContainer(d.DB, char)
	if errors.Is(err, ErrNotEnoughMoney) || errors.Is(err, ErrContainerMax) {
		// The button is hidden/disabled for these cases; a stray POST is safe
		// because UpgradeContainer refuses to charge before checking.
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	if err != nil {
		slog.Error("Container upgrade failed", "error", err, "characterID", char.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	slog.Info("Container upgraded", "characterID", char.ID, "level", refreshed.ContainerLevel)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// addUpgradeData fills the container upgrade card on the dashboard.
func addUpgradeData(data map[string]any, char Character) {
	level := char.ContainerLevel
	data["UpgradeCurrent"] = game.ContainerLevelName(level)
	data["UpgradeAtMax"] = level >= game.MaxContainerLevel
	data["UpgradeNext"] = game.ContainerLevelName(level + 1)
	if level >= game.MaxContainerLevel {
		data["UpgradeCost"] = ""
		data["UpgradeAffordable"] = false
		return
	}
	cost := game.UpgradeCostCents(level, game.ContainerUpgradeBaseCents)
	data["UpgradeCost"] = formatCents(cost)
	data["UpgradeAffordable"] = char.MoneyCents >= cost
}

// formatCents renders cents as German currency, e.g. 1234 -> "12,34 €".
func formatCents(cents int) string {
	return fmt.Sprintf("%d,%02d \u20ac", cents/100, cents%100)
}

// fillPercent is the container fill level for the status bar, capped at 100.
func fillPercent(cents, capacity int) int {
	if capacity <= 0 {
		return 0
	}
	if cents >= capacity {
		return 100
	}
	return cents * 100 / capacity
}
