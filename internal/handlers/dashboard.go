package handlers

import (
	"database/sql"
	"errors"
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

	// Settle passive income the moment it is due, so the toggle stays honest.
	bafog, char, err := d.collectBafog(char)
	if err != nil {
		slog.Error("Failed to collect BAföG", "error", err, "characterID", char.ID)
		http.Error(w, "interner Fehler", http.StatusInternalServerError)
		return
	}

	data := d.baseData(w, r)
	d.addCharacterStatus(data, char)
	data["Character"] = char
	addUpgradeData(data, char)
	addSkillOverview(data, d.DB, char)
	data["BafogEnabled"] = char.BafogNextAt != nil
	data["BafogPayoutText"] = web.FormatCents(game.BafogPayoutCents)
	data["BafogSummary"] = bafog
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
	data["UpgradeCost"] = web.FormatCents(cost)
	data["UpgradeAffordable"] = char.MoneyCents >= cost
}

// SkillOverviewItem is one skill's level for the dashboard "Dein Stand" card.
type SkillOverviewItem struct {
	Name     string
	Level    int
	MaxLevel int
}

// addSkillOverview fills the current skill levels shown in the "Dein Stand" card.
func addSkillOverview(data map[string]any, db *sql.DB, char Character) {
	levels, err := SkillLevels(db, char.ID)
	if err != nil {
		// Skills are cosmetic on the dashboard; a failure must not break it.
		data["SkillOverview"] = nil
		return
	}
	items := make([]SkillOverviewItem, 0, len(game.SkillCatalog))
	for _, s := range game.SkillCatalog {
		items = append(items, SkillOverviewItem{
			Name:     s.Name,
			Level:    levels[s.Slug],
			MaxLevel: s.MaxLevel,
		})
	}
	data["SkillOverview"] = items
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
