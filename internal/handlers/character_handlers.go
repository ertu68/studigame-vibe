package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"studigame/internal/web"
)

// CharacterNewForm shows the Studienfach picker and name field for onboarding.
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
	subjectID, err := strconv.ParseInt(r.FormValue("subject"), 10, 64)

	if n := utf8.RuneCountInString(name); n < 3 || n > 20 {
		d.renderCharacterForm(w, r, name, "Der Name braucht 3 bis 20 Zeichen.")
		return
	}
	if err != nil {
		d.renderCharacterForm(w, r, name, "Bitte wähle ein Studienfach.")
		return
	}

	edition, err := activeEditionSlug(d.DB)
	if err != nil {
		d.renderCharacterForm(w, r, name, "Gerade ist keine Stadt offen.")
		return
	}

	char, err := CreateCharacter(d.DB, u.ID, edition, subjectID, name)
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
	case errors.Is(err, ErrSubjectNotFound):
		d.renderCharacterForm(w, r, name, "Dieses Studienfach gibt es hier nicht.")
		return
	case err != nil:
		slog.Error("Character creation failed", "error", err, "userID", u.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	slog.Info("Character created", "characterID", char.ID, "userID", u.ID, "subject", subjectID)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// activeEditionSlug returns the slug of the single open edition (MVP).
func activeEditionSlug(db *sql.DB) (string, error) {
	editions, err := ActiveEditions(db)
	if err != nil {
		return "", err
	}
	if len(editions) != 1 {
		return "", ErrEditionNotFound
	}
	return editions[0].Slug, nil
}

func (d *Deps) renderCharacterForm(w http.ResponseWriter, r *http.Request, name, msg string) {
	editions, err := ActiveEditions(d.DB)
	if err != nil {
		slog.Error("Failed to load editions", "error", err)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}
	if len(editions) == 0 {
		slog.Error("no active edition")
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}
	edition := editions[0]
	subjects, err := ActiveSubjects(d.DB, edition.ID)
	if err != nil {
		slog.Error("Failed to load subjects", "error", err)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}
	data := d.baseData(w, r)
	data["Edition"] = edition
	data["Subjects"] = subjects
	data["Name"] = name
	data["Error"] = msg
	web.Render(w, "character_new.html", data)
}
