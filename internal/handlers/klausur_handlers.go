package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"studigame/internal/web"
)

// Klausur shows the exam prep screen: either the running shift or the form.
func (d *Deps) Klausur(u *user, w http.ResponseWriter, r *http.Request) {
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

	summary, char, err := d.settleDueAction(char)
	if err != nil {
		slog.Error("Failed to resolve action", "error", err, "characterID", char.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	d.renderKlausur(w, r, char, summary, "")
}

// StartKlausur starts an exam prep shift of the requested length.
func (d *Deps) StartKlausur(u *user, w http.ResponseWriter, r *http.Request) {
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

	// A finished shift is settled first, so the player cannot stack two.
	if _, char, err = d.settleDueAction(char); err != nil {
		slog.Error("Failed to resolve action", "error", err, "characterID", char.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	seconds, _ := strconv.Atoi(r.FormValue("duration"))
	duration := time.Duration(seconds) * time.Second

	err = StartKlausur(d.DB, char, duration, d.now())
	switch {
	case errors.Is(err, ErrInvalidShift):
		d.renderKlausur(w, r, char, nil, "Diese Schichtlänge gibt es nicht.")
		return
	case errors.Is(err, ErrActionRunning):
		d.renderKlausur(w, r, char, nil, "Du bist noch unterwegs.")
		return
	case err != nil:
		slog.Error("Failed to start klausur", "error", err, "characterID", char.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	slog.Info("Klausur started", "characterID", char.ID, "seconds", seconds)
	http.Redirect(w, r, "/klausur", http.StatusSeeOther)
}

func (d *Deps) renderKlausur(w http.ResponseWriter, r *http.Request, char Character, summary *ActionSummary, msg string) {
	data := d.baseData(w, r)
	d.addCharacterStatus(data, char)
	data["Character"] = char
	data["Summary"] = summary
	data["Error"] = msg

	if open, err := FindOpenAction(d.DB, char.ID); err == nil {
		remaining := open.Remaining(d.now())
		data["Open"] = open
		data["Remaining"] = formatDuration(remaining)
		data["ReturnsAt"] = formatClock(open.EndsAt)
		data["EndsAtIso"] = open.EndsAt.UTC().Format(time.RFC3339)
		// One second of slack so the reload never arrives before the shift is due.
		data["MetaRefresh"] = int(remaining.Seconds()) + 1
	} else if !errors.Is(err, ErrNoOpenAction) {
		slog.Error("Failed to load open action", "error", err, "characterID", char.ID)
	} else {
		data["Shifts"] = shiftOptions()
	}

	web.Render(w, "klausur.html", data)
}
