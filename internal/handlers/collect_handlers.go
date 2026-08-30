package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"studigame/internal/game"
	"studigame/internal/web"
)

// shiftOption is one selectable shift length in the collect form.
type shiftOption struct {
	Seconds int
	Label   string
}

// Actions shows the collect screen: either the running shift or the form.
func (d *Deps) Actions(u *user, w http.ResponseWriter, r *http.Request) {
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

	d.renderActions(w, r, char, summary, "")
}

// Collect starts a shift of the requested length.
func (d *Deps) Collect(u *user, w http.ResponseWriter, r *http.Request) {
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

	err = StartCollect(d.DB, char, duration, d.now())
	switch {
	case errors.Is(err, ErrInvalidShift):
		d.renderActions(w, r, char, nil, "Diese Schichtlänge gibt es nicht.")
		return
	case errors.Is(err, ErrActionRunning):
		d.renderActions(w, r, char, nil, "Du bist noch unterwegs.")
		return
	case err != nil:
		slog.Error("Failed to start collect", "error", err, "characterID", char.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	slog.Info("Collect started", "characterID", char.ID, "seconds", seconds)
	http.Redirect(w, r, "/actions", http.StatusSeeOther)
}

// Job starts a working-student shift of the requested length.
func (d *Deps) Job(u *user, w http.ResponseWriter, r *http.Request) {
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

	err = StartJob(d.DB, char, duration, d.now())
	switch {
	case errors.Is(err, ErrInvalidShift):
		d.renderActions(w, r, char, nil, "Diese Schichtlänge gibt es nicht.")
		return
	case errors.Is(err, ErrActionRunning):
		d.renderActions(w, r, char, nil, "Du bist noch unterwegs.")
		return
	case err != nil:
		slog.Error("Failed to start job", "error", err, "characterID", char.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	slog.Info("Job started", "characterID", char.ID, "seconds", seconds)
	http.Redirect(w, r, "/actions", http.StatusSeeOther)
}

// settleDueAction resolves a finished shift and returns the refreshed character.
func (d *Deps) settleDueAction(char Character) (*ActionSummary, Character, error) {
	summary, err := ResolveDueAction(d.DB, char, d.now(), d.roll())
	if err != nil || summary == nil {
		return nil, char, err
	}
	refreshed, err := CharacterForUser(d.DB, char.UserID, char.EditionID)
	if err != nil {
		return summary, char, err
	}
	return summary, refreshed, nil
}

func (d *Deps) renderActions(w http.ResponseWriter, r *http.Request, char Character, summary *ActionSummary, msg string) {
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

	web.Render(w, "actions.html", data)
}

// gameLocation is the wall clock players read: the imaginary uni-city has
// German time, no matter where the server runs.
var gameLocation = mustLoadLocation("Europe/Berlin")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

// formatClock renders a time as the local wall clock, e.g. "14:35 Uhr".
func formatClock(t time.Time) string {
	return t.In(gameLocation).Format("15:04") + " Uhr"
}

func shiftOptions() []shiftOption {
	out := make([]shiftOption, 0, len(game.CollectDurations))
	for _, d := range game.CollectDurations {
		out = append(out, shiftOption{Seconds: int(d.Seconds()), Label: formatDuration(d)})
	}
	return out
}

// formatDuration renders a duration in German short form, e.g. "4 h 5 min".
// It truncates rather than rounds, so a countdown never shows more time than is left.
func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "fertig"
	}
	if d < time.Minute {
		return strconv.Itoa(int(d.Seconds())) + " s"
	}
	d = d.Truncate(time.Minute)
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	switch {
	case hours > 0 && minutes > 0:
		return strconv.Itoa(hours) + " h " + strconv.Itoa(minutes) + " min"
	case hours > 0:
		return strconv.Itoa(hours) + " h"
	default:
		return strconv.Itoa(minutes) + " min"
	}
}
