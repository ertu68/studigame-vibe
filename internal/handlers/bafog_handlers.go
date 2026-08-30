package handlers

import (
	"errors"
	"log/slog"
	"net/http"
)

// Bafog toggles the player's passive BAföG-Förderung and returns to the dashboard.
// "enabled=1" applies, anything else cancels. It is not a shift: once applied the
// payout arrives on its own every 6 h (see CollectBafog).
func (d *Deps) Bafog(u *user, w http.ResponseWriter, r *http.Request) {
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

	enabled := r.FormValue("enabled") == "1"
	if _, err := SetBafog(d.DB, char, enabled, d.now()); err != nil {
		slog.Error("Failed to toggle BAföG", "error", err, "characterID", char.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}
	slog.Info("BAföG toggled", "characterID", char.ID, "enabled", enabled)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
