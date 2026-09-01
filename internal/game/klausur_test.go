package game

import (
	"testing"
	"time"
)

func TestResolveKlausur(t *testing.T) {
	tests := []struct {
		name     string
		params   ECTSParams
		wantPts  float64
		wantMood int
	}{
		{
			name: "rested, average skill",
			params: ECTSParams{
				Duration:    4 * time.Hour,
				Mood:        100,
				SkillFactor: 1.0,
			},
			wantPts:  1.0, // 0.25 * 4 * 1.0 * 1.0
			wantMood: -20,
		},
		{
			name: "exhausted, average skill",
			params: ECTSParams{
				Duration:    4 * time.Hour,
				Mood:        0,
				SkillFactor: 1.0,
			},
			wantPts:  0.6, // 0.25 * 4 * 0.6 * 1.0
			wantMood: -20,
		},
		{
			name: "rested, skilled",
			params: ECTSParams{
				Duration:    4 * time.Hour,
				Mood:        100,
				SkillFactor: 1.1,
			},
			wantPts:  1.1, // 0.25 * 4 * 1.0 * 1.1
			wantMood: -20,
		},
		{
			name: "short shift",
			params: ECTSParams{
				Duration:    1 * time.Hour,
				Mood:        100,
				SkillFactor: 1.0,
			},
			wantPts:  0.25,
			wantMood: -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveKlausur(tt.params)
			if got.ECTSPoints != tt.wantPts {
				t.Errorf("points: got %.2f, want %.2f", got.ECTSPoints, tt.wantPts)
			}
			if got.MoodDelta != tt.wantMood {
				t.Errorf("mood: got %d, want %d", got.MoodDelta, tt.wantMood)
			}
		})
	}
}

func TestResolveKlausurMoodBounds(t *testing.T) {
	p := ECTSParams{Duration: time.Hour, Mood: 100, SkillFactor: 1.0}
	got := ResolveKlausur(p)
	if got.MoodDelta != -5 {
		t.Errorf("short shift mood cost = %d, want -5", got.MoodDelta)
	}

	p.Duration = 15 * time.Minute
	got = ResolveKlausur(p)
	if got.MoodDelta != -2 {
		t.Errorf("15-min shift mood cost = %d, want -2", got.MoodDelta)
	}
}

func TestResolveKlausurSkillDefaults(t *testing.T) {
	rested := ECTSParams{Duration: time.Hour, Mood: 100, SkillFactor: 0}    // unset
	skilled := ECTSParams{Duration: time.Hour, Mood: 100, SkillFactor: 1.05} // buffed

	base := ResolveKlausur(rested).ECTSPoints
	buffed := ResolveKlausur(skilled).ECTSPoints

	if buffed <= base {
		t.Errorf("skill bonus didn't increase yield: base=%.2f, buffed=%.2f", base, buffed)
	}
}