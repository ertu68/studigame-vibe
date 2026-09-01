package game

import "math"

// SkillCatalog is the reduced MVP skill tree (plan §3b: "reduzierte Variante
// der 9 Original-Skills"). Each skill has a real effect on the current loop so
// the skill system is meaningful now and stays testable. All numbers here are
// provisional tuning knobs, not final balance.
var SkillCatalog = []Skill{
	{
		Slug:        "geschicklichkeit",
		Name:        "Lerneffizienz",
		Description: "Mehr Leergut pro Schicht.",
		BaseCost:    1000,
		MaxLevel:    20,
	},
	{
		Slug:        "auftreten",
		Name:        "Auftreten",
		Description: "Bessere Bezahlung im Werkstudenten-Job.",
		BaseCost:    1000,
		MaxLevel:    20,
	},
	{
		Slug:        "konzentration",
		Name:        "Konzentration",
		Description: "Schnellere Weiterbildungen.",
		BaseCost:    1200,
		MaxLevel:    10,
	},
}

// Skill is one entry of the reduced MVP skill tree.
type Skill struct {
	Slug        string
	Name        string
	Description string
	BaseCost    int
	MaxLevel    int
}

// SkillBySlug returns a skill by its slug, or nil.
func SkillBySlug(slug string) *Skill {
	for i := range SkillCatalog {
		if SkillCatalog[i].Slug == slug {
			return &SkillCatalog[i]
		}
	}
	return nil
}

// SkillCostCents is what training the next level costs: base × level^1.8,
// where level is the CURRENT level (training raises it by one). The exponent
// > 1 makes later levels overproportionally expensive (anti-snowball), the
// same curve as the container upgrade. level must be 0..MaxLevel-1; beyond
// MaxLevel it reports 0 (not buyable). With base 1000: 10 €, ~34 €, ~72 €, …
func SkillCostCents(s Skill, level int) int {
	if level < 0 || level >= s.MaxLevel {
		return 0
	}
	return int(float64(s.BaseCost) * math.Pow(float64(level+1), 1.8))
}

// SubjectSkillCostFactor is the price multiplier a Studienfach applies to a
// skill's training cost (plan §3b subject bonus/mali). Geisteswissenschaften
// trains cheap; Jura pays more; Informatik gets a discount on Konzentration.
func SubjectSkillCostFactor(subject, skill string) float64 {
	switch subject {
	case "Geisteswissenschaften":
		return 0.8
	case "Jura":
		return 1.2
	case "Informatik":
		if skill == "konzentration" {
			return 0.9
		}
	}
	return 1.0
}

// SubjectSkillDurationFactor is the training-time multiplier per subject.
// Jura trains slower (plan §3b "Teurere, längere Weiterbildungen").
func SubjectSkillDurationFactor(subject string) float64 {
	if subject == "Jura" {
		return 1.2
	}
	return 1.0
}

// SkillTrainingDuration is how long one level of training takes. Konzentration
// levels shorten it; the subject modifies it.
func SkillTrainingDuration(skill string, konzentrationLevel int, subject string) int {
	base := 2 * 60 // minutes, provisional
	konz := 1.0 - 0.05*float64(konzentrationLevel)
	if konz < 0.5 {
		konz = 0.5
	}
	return int(float64(base) * konz * SubjectSkillDurationFactor(subject))
}

// CollectSkillFactor is how much a character's Lerneffizienz raises the
// collect yield: +5 % per level.
func CollectSkillFactor(goodGrasLevel int) float64 {
	return 1 + 0.05*float64(goodGrasLevel)
}

// JobSkillFactor is how much Auftreten raises the job wage: +5 % per level.
func JobSkillFactor(auftretenLevel int) float64 {
	return 1 + 0.05*float64(auftretenLevel)
}

// KlausurSkillFactor is how much Konzentration raises ECTS yield: +5 % per level.
func KlausurSkillFactor(konzentrationLevel int) float64 {
	return 1 + 0.05*float64(konzentrationLevel)
}
