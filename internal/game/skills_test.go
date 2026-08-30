package game

import (
	"testing"
	"time"
)

func TestSkillCostCentsOverlinear(t *testing.T) {
	// The cost curve is base × level^1.8, so the gap between consecutive
	// levels grows — later levels are relatively more expensive.
	prev := 0
	for level := 0; level < 5; level++ {
		c := SkillCostCents(SkillCatalog[0], level)
		if c == 0 {
			t.Fatalf("level %d should be buyable", level)
		}
		if prev > 0 && c <= prev {
			t.Fatalf("cost must rise with level: level %d = %d, prev = %d", level, c, prev)
		}
		prev = c
	}
}

func TestSkillCostCentsAtMaxLevelIsZero(t *testing.T) {
	s := SkillCatalog[0]
	for level := s.MaxLevel; level < s.MaxLevel+2; level++ {
		if c := SkillCostCents(s, level); c != 0 {
			t.Fatalf("expected 0 at/above max level %d, got %d", level, c)
		}
	}
}

func TestSkillCostCentsNegativeIsZero(t *testing.T) {
	if c := SkillCostCents(SkillCatalog[0], -1); c != 0 {
		t.Fatalf("expected 0 for negative level, got %d", c)
	}
}

func TestSubjectSkillCostFactor(t *testing.T) {
	cases := []struct {
		subject, skill string
		want           float64
	}{
		{"Geisteswissenschaften", "geschicklichkeit", 0.8},
		{"Jura", "geschicklichkeit", 1.2},
		{"Informatik", "konzentration", 0.9},
		{"Informatik", "geschicklichkeit", 1.0},
		{"BWL", "geschicklichkeit", 1.0},
	}
	for _, c := range cases {
		if got := SubjectSkillCostFactor(c.subject, c.skill); got != c.want {
			t.Errorf("%s/%s: got %v, want %v", c.subject, c.skill, got, c.want)
		}
	}
}

func TestSubjectSkillDurationFactor(t *testing.T) {
	if got := SubjectSkillDurationFactor("Jura"); got != 1.2 {
		t.Fatalf("Jura: got %v, want 1.2", got)
	}
	if got := SubjectSkillDurationFactor("BWL"); got != 1.0 {
		t.Fatalf("BWL: got %v, want 1.0", got)
	}
}

func TestSkillTrainingDurationShortensWithKonzentration(t *testing.T) {
	base := SkillTrainingDuration("geschicklichkeit", 0, "BWL")
	fast := SkillTrainingDuration("geschicklichkeit", 5, "BWL")
	if fast >= base {
		t.Fatalf("konzentration must shorten training: base %d, fast %d", base, fast)
	}
	if j := SkillTrainingDuration("geschicklichkeit", 0, "Jura"); j <= base {
		t.Fatalf("Jura must lengthen training: %d vs base %d", j, base)
	}
}

func TestSkillFactorDefaultsToNoBonusWhenZero(t *testing.T) {
	base := ResolveCollect(CollectParams{Duration: 1 * hour}, 0.5).EarningsCents
	buffed := ResolveCollect(CollectParams{Duration: 1 * hour, SkillFactor: 1.05}, 0.5).EarningsCents
	if buffed <= base {
		t.Fatalf("skill factor must raise yield: base %d, buffed %d", base, buffed)
	}
}

const hour = time.Hour

func TestResolveJobSkillFactor(t *testing.T) {
	base := ResolveJob(JobParams{Duration: hour}).EarningsCents
	buffed := ResolveJob(JobParams{Duration: hour, SkillFactor: 1.05}).EarningsCents
	if buffed <= base {
		t.Fatalf("skill factor must raise job wage: base %d, buffed %d", base, buffed)
	}
}
