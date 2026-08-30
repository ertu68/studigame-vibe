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

// SkillView is one skill as the template renders it.
type SkillView struct {
	Slug        string
	Name        string
	Description string
	Level       int
	MaxLevel    int
	CostCents   int // subject-adjusted price of the next level; 0 = max
	AtMax       bool
}

// Skills shows the skill tree, resolving any due training first.
func (d *Deps) Skills(u *user, w http.ResponseWriter, r *http.Request) {
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

	if err := ResolveSkillTrainings(d.DB, char.ID, d.now()); err != nil {
		slog.Error("Failed to resolve skill trainings", "error", err, "characterID", char.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	d.renderSkills(w, r, char, "")
}

// TrainSkill starts a level-up run for a skill.
func (d *Deps) TrainSkill(u *user, w http.ResponseWriter, r *http.Request) {
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

	if err := ResolveSkillTrainings(d.DB, char.ID, d.now()); err != nil {
		slog.Error("Failed to resolve skill trainings", "error", err, "characterID", char.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	slug := r.FormValue("skill")
	_, err = StartSkillTraining(d.DB, char, slug, d.now())
	switch {
	case errors.Is(err, ErrSkillTrainingRunning):
		d.renderSkills(w, r, char, "Eine Weiterbildung läuft bereits.")
		return
	case errors.Is(err, ErrSkillUnknown):
		d.renderSkills(w, r, char, "Diese Weiterbildung gibt es nicht.")
		return
	case errors.Is(err, ErrSkillAtMax):
		d.renderSkills(w, r, char, "Dieser Skill ist bereits ausgelernt.")
		return
	case errors.Is(err, ErrNotEnoughMoney):
		d.renderSkills(w, r, char, "Dafür reicht das Geld nicht.")
		return
	case err != nil:
		slog.Error("Failed to start skill training", "error", err, "characterID", char.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	slog.Info("Skill training started", "characterID", char.ID, "skill", slug)
	http.Redirect(w, r, "/skills", http.StatusSeeOther)
}

// CancelSkill aborts a running training and refunds 80 %.
func (d *Deps) CancelSkill(u *user, w http.ResponseWriter, r *http.Request) {
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

	// A due training must be settled before cancel, so a freshly completed run
	// cannot be cancelled for a refund.
	resolveChar := func() Character {
		if err := ResolveSkillTrainings(d.DB, char.ID, d.now()); err != nil {
			slog.Error("Failed to resolve skill trainings", "error", err, "characterID", char.ID)
		}
		c, err := AnyCharacterForUser(d.DB, u.ID)
		if err == nil {
			return c
		}
		return char
	}
	refreshed := resolveChar()

	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	refund, err := CancelSkillTraining(d.DB, refreshed, id)
	switch {
	case errors.Is(err, ErrNoSkillTraining):
		d.renderSkills(w, r, refreshed, "Diese Weiterbildung gibt es nicht mehr.")
		return
	case err != nil:
		slog.Error("Failed to cancel skill training", "error", err, "characterID", refreshed.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	slog.Info("Skill training cancelled", "characterID", refreshed.ID, "trainingID", id, "refund", refund)
	http.Redirect(w, r, "/skills", http.StatusSeeOther)
}

func (d *Deps) renderSkills(w http.ResponseWriter, r *http.Request, char Character, msg string) {
	data := d.baseData(w, r)
	d.addCharacterStatus(data, char)

	levels, err := SkillLevels(d.DB, char.ID)
	if err != nil {
		slog.Error("Failed to load skill levels", "error", err, "characterID", char.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	views := make([]SkillView, 0, len(game.SkillCatalog))
	for _, s := range game.SkillCatalog {
		lvl := levels[s.Slug]
		cost := 0
		if lvl < s.MaxLevel {
			base := game.SkillCostCents(s, lvl)
			cost = int(float64(base) * game.SubjectSkillCostFactor(char.SubjectName, s.Slug))
		}
		views = append(views, SkillView{
			Slug:        s.Slug,
			Name:        s.Name,
			Description: s.Description,
			Level:       lvl,
			MaxLevel:    s.MaxLevel,
			CostCents:   cost,
			AtMax:       lvl >= s.MaxLevel,
		})
	}

	data["Skills"] = views
	data["SkillLevels"] = levels
	data["Subject"] = char.SubjectName
	data["Error"] = msg
	data["Character"] = char

	if training, err := FindSkillTraining(d.DB, char.ID); err == nil {
		remaining := training.EndsAt.Sub(d.now())
		if remaining < 0 {
			remaining = 0
		}
		data["Training"] = training
		data["TrainingRemaining"] = formatDuration(remaining)
		data["TrainingEndsAt"] = training.EndsAt.UTC().Format(time.RFC3339)
		data["TrainingMetaRefresh"] = int(remaining.Seconds()) + 1
	}

	web.Render(w, "skills.html", data)
}
