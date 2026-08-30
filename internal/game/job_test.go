package game

import (
	"testing"
	"time"
)

func TestResolveJob(t *testing.T) {
	tests := []struct {
		name          string
		params        JobParams
		wantEarnings  int
		wantMoodDelta int
	}{
		{
			name:          "one hour rested pays the flat wage",
			params:        JobParams{Duration: time.Hour, Mood: 100},
			wantEarnings:  800,
			wantMoodDelta: -5,
		},
		{
			name:          "exhausted character earns 60 %",
			params:        JobParams{Duration: time.Hour, Mood: 0},
			wantEarnings:  480,
			wantMoodDelta: -5,
		},
		{
			name:          "four hours scale linearly",
			params:        JobParams{Duration: 4 * time.Hour, Mood: 100},
			wantEarnings:  3200,
			wantMoodDelta: -20,
		},
		{
			name:          "short shift still costs mood",
			params:        JobParams{Duration: 15 * time.Minute, Mood: 100},
			wantEarnings:  200,
			wantMoodDelta: -2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveJob(tt.params)
			if got.EarningsCents != tt.wantEarnings {
				t.Errorf("earnings: got %d, want %d", got.EarningsCents, tt.wantEarnings)
			}
			if got.MoodDelta != tt.wantMoodDelta {
				t.Errorf("mood delta: got %d, want %d", got.MoodDelta, tt.wantMoodDelta)
			}
		})
	}
}

// A job pays a flat rate: the same hour must cost/earn identically regardless
// of any noise — market the contrast to the luck-driven collect action.
func TestResolveJobRateIsSteady(t *testing.T) {
	a := ResolveJob(JobParams{Duration: time.Hour, Mood: 100})
	b := ResolveJob(JobParams{Duration: time.Hour, Mood: 100})
	if a != b {
		t.Fatalf("job wage not steady: %+v vs %+v", a, b)
	}
}
