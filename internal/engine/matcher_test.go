package engine_test

import (
	"testing"
	"time"

	"github.com/lm/karaclean/internal/config"
	"github.com/lm/karaclean/internal/engine"
)

func strPtr(s string) *string { return &s }

func TestMatchesConditions(t *testing.T) {
	runTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		bookmark engine.Bookmark
		conds    *config.Conditions
		want     bool
	}{
		{
			name:     "nil conditions matches all",
			bookmark: engine.Bookmark{ID: "1", CreatedAt: runTime.Add(-24 * time.Hour)},
			conds:    nil,
			want:     true,
		},
		{
			name:     "olderThan 30d matches bookmark created 31 days ago",
			bookmark: engine.Bookmark{ID: "2", CreatedAt: runTime.Add(-31 * 24 * time.Hour)},
			conds:    &config.Conditions{OlderThan: strPtr("30d")},
			want:     true,
		},
		{
			name:     "olderThan 30d does not match bookmark created exactly 30 days ago (strictly greater than)",
			bookmark: engine.Bookmark{ID: "3", CreatedAt: runTime.Add(-30 * 24 * time.Hour)},
			conds:    &config.Conditions{OlderThan: strPtr("30d")},
			want:     false,
		},
		{
			name:     "olderThan 30d does not match bookmark created 29 days ago",
			bookmark: engine.Bookmark{ID: "4", CreatedAt: runTime.Add(-29 * 24 * time.Hour)},
			conds:    &config.Conditions{OlderThan: strPtr("30d")},
			want:     false,
		},
		{
			name:     "olderThan 0h matches bookmark with positive age",
			bookmark: engine.Bookmark{ID: "5", CreatedAt: runTime.Add(-1 * time.Hour)},
			conds:    &config.Conditions{OlderThan: strPtr("0h")},
			want:     true,
		},
		{
			name:     "olderThan 0h does not match bookmark created at runTime (0 is not > 0)",
			bookmark: engine.Bookmark{ID: "6", CreatedAt: runTime},
			conds:    &config.Conditions{OlderThan: strPtr("0h")},
			want:     false,
		},
		{
			name:     "source rss matches bookmark with source rss",
			bookmark: engine.Bookmark{ID: "7", CreatedAt: runTime, Source: "rss"},
			conds:    &config.Conditions{Source: strPtr("rss")},
			want:     true,
		},
		{
			name:     "source rss does not match bookmark with source web",
			bookmark: engine.Bookmark{ID: "8", CreatedAt: runTime, Source: "web"},
			conds:    &config.Conditions{Source: strPtr("rss")},
			want:     false,
		},
		{
			name: "AND: olderThan and source both match",
			bookmark: engine.Bookmark{
				ID: "9", CreatedAt: runTime.Add(-31 * 24 * time.Hour), Source: "rss",
			},
			conds: &config.Conditions{OlderThan: strPtr("30d"), Source: strPtr("rss")},
			want:  true,
		},
		{
			name: "AND: olderThan matches but source fails",
			bookmark: engine.Bookmark{
				ID: "10", CreatedAt: runTime.Add(-31 * 24 * time.Hour), Source: "web",
			},
			conds: &config.Conditions{OlderThan: strPtr("30d"), Source: strPtr("rss")},
			want:  false,
		},
		{
			name: "AND: source matches but age fails",
			bookmark: engine.Bookmark{
				ID: "11", CreatedAt: runTime.Add(-29 * 24 * time.Hour), Source: "rss",
			},
			conds: &config.Conditions{OlderThan: strPtr("30d"), Source: strPtr("rss")},
			want:  false,
		},
		{
			name:     "source-only condition matches when OlderThan is nil",
			bookmark: engine.Bookmark{ID: "12", CreatedAt: runTime, Source: "web"},
			conds:    &config.Conditions{Source: strPtr("web")},
			want:     true,
		},
		{
			name:     "olderThan 2w correctly compares against 14 days (15 days old bookmark matches)",
			bookmark: engine.Bookmark{ID: "13", CreatedAt: runTime.Add(-15 * 24 * time.Hour)},
			conds:    &config.Conditions{OlderThan: strPtr("2w")},
			want:     true,
		},
		{
			name:     "olderThan 1mo correctly compares against 30 days (31 days old bookmark matches)",
			bookmark: engine.Bookmark{ID: "14", CreatedAt: runTime.Add(-31 * 24 * time.Hour)},
			conds:    &config.Conditions{OlderThan: strPtr("1mo")},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.MatchesConditions(tt.bookmark, tt.conds, runTime)
			if got != tt.want {
				t.Errorf("MatchesConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}
