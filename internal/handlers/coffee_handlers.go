package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"studigame/internal/game"
	"studigame/internal/web"
)

// Coffee buys one boost from the Campus-Kiosk and returns to the dashboard.
// A stray POST when the player cannot afford it is safe: BuyCoffee refuses to
// charge before checking.
func (d *Deps) Coffee(u *user, w http.ResponseWriter, r *http.Request) {
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

	refreshed, err := BuyCoffee(d.DB, char, d.now())
	if errors.Is(err, ErrNotEnoughMoney) {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	if err != nil {
		slog.Error("Coffee purchase failed", "error", err, "characterID", char.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	slog.Info("Coffee bought", "characterID", refreshed.ID,
		"price", web.FormatCents(game.CoffeeCostCents))
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
