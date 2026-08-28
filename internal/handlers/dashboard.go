package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"pennergame/internal/web"
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
	web.Render(w, "dashboard.html", data)
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
