package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"pennergame/internal/web"
)

// CharacterNewForm shows the edition picker and name field for onboarding.
func (d *Deps) CharacterNewForm(u *user, w http.ResponseWriter, r *http.Request) {
	if _, err := AnyCharacterForUser(d.DB, u.ID); err == nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	d.renderCharacterForm(w, r, u.Name, "")
}

// CharacterNew creates the character and drops the player on the dashboard.
func (d *Deps) CharacterNew(u *user, w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.FormValue("name"))
	edition := r.FormValue("edition")

	if len(name) < 3 || len(name) > 20 {
		d.renderCharacterForm(w, r, name, "Der Name braucht 3 bis 20 Zeichen.")
		return
	}

	char, err := CreateCharacter(d.DB, u.ID, edition, name)
	switch {
	case errors.Is(err, ErrNameTaken):
		d.renderCharacterForm(w, r, name, "Diesen Namen gibt es in der Stadt schon.")
		return
	case errors.Is(err, ErrCharacterExists):
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	case errors.Is(err, ErrEditionNotFound):
		d.renderCharacterForm(w, r, name, "Diese Stadt ist gerade nicht offen.")
		return
	case err != nil:
		slog.Error("Character creation failed", "error", err, "userID", u.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	slog.Info("Character created", "characterID", char.ID, "userID", u.ID, "edition", edition)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (d *Deps) renderCharacterForm(w http.ResponseWriter, r *http.Request, name, msg string) {
	editions, err := ActiveEditions(d.DB)
	if err != nil {
		slog.Error("Failed to load editions", "error", err)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}
	data := d.baseData(w, r)
	data["Editions"] = editions
	data["Name"] = name
	data["Error"] = msg
	web.Render(w, "character_new.html", data)
}
