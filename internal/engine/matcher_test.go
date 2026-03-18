package engine_test

import (
	"testing"
	"time"

	"github.com/lm/karaclean/internal/config"
	"github.com/lm/karaclean/internal/engine"
)

func strPtr(s string) *string  { return &s }
func boolPtr(b bool) *bool     { return &b }

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
		// --- archived condition ---
		{
			name:     "archived true matches archived bookmark",
			bookmark: engine.Bookmark{ID: "20", CreatedAt: runTime, Archived: true},
			conds:    &config.Conditions{Archived: boolPtr(true)},
			want:     true,
		},
		{
			name:     "archived true does not match non-archived bookmark",
			bookmark: engine.Bookmark{ID: "21", CreatedAt: runTime, Archived: false},
			conds:    &config.Conditions{Archived: boolPtr(true)},
			want:     false,
		},
		{
			name:     "archived false matches non-archived bookmark",
			bookmark: engine.Bookmark{ID: "22", CreatedAt: runTime, Archived: false},
			conds:    &config.Conditions{Archived: boolPtr(false)},
			want:     true,
		},
		{
			name:     "archived false does not match archived bookmark",
			bookmark: engine.Bookmark{ID: "23", CreatedAt: runTime, Archived: true},
			conds:    &config.Conditions{Archived: boolPtr(false)},
			want:     false,
		},
		// --- favourited condition ---
		{
			name:     "favourited true matches favourited bookmark",
			bookmark: engine.Bookmark{ID: "30", CreatedAt: runTime, Favourited: true},
			conds:    &config.Conditions{Favourited: boolPtr(true)},
			want:     true,
		},
		{
			name:     "favourited true does not match non-favourited bookmark",
			bookmark: engine.Bookmark{ID: "31", CreatedAt: runTime, Favourited: false},
			conds:    &config.Conditions{Favourited: boolPtr(true)},
			want:     false,
		},
		{
			name:     "favourited false matches non-favourited bookmark",
			bookmark: engine.Bookmark{ID: "32", CreatedAt: runTime, Favourited: false},
			conds:    &config.Conditions{Favourited: boolPtr(false)},
			want:     true,
		},
		{
			name:     "favourited false does not match favourited bookmark",
			bookmark: engine.Bookmark{ID: "33", CreatedAt: runTime, Favourited: true},
			conds:    &config.Conditions{Favourited: boolPtr(false)},
			want:     false,
		},
		// --- hasTag condition ---
		{
			name:     "hasTag read-later matches bookmark with matching tag",
			bookmark: engine.Bookmark{ID: "40", CreatedAt: runTime, Tags: []string{"news", "read-later"}},
			conds:    &config.Conditions{HasTag: strPtr("read-later")},
			want:     true,
		},
		{
			name:     "hasTag read-later does not match bookmark without that tag",
			bookmark: engine.Bookmark{ID: "41", CreatedAt: runTime, Tags: []string{"news", "tech"}},
			conds:    &config.Conditions{HasTag: strPtr("read-later")},
			want:     false,
		},
		{
			name:     "hasTag read-later does not match bookmark with nil tags",
			bookmark: engine.Bookmark{ID: "42", CreatedAt: runTime},
			conds:    &config.Conditions{HasTag: strPtr("read-later")},
			want:     false,
		},
		{
			name:     "hasTag read-later does not match bookmark with empty tags",
			bookmark: engine.Bookmark{ID: "43", CreatedAt: runTime, Tags: []string{}},
			conds:    &config.Conditions{HasTag: strPtr("read-later")},
			want:     false,
		},
		// --- lacksTag condition ---
		{
			name:     "lacksTag keep matches bookmark without that tag",
			bookmark: engine.Bookmark{ID: "50", CreatedAt: runTime, Tags: []string{"news", "tech"}},
			conds:    &config.Conditions{LacksTag: strPtr("keep")},
			want:     true,
		},
		{
			name:     "lacksTag keep does not match bookmark with that tag",
			bookmark: engine.Bookmark{ID: "51", CreatedAt: runTime, Tags: []string{"keep", "tech"}},
			conds:    &config.Conditions{LacksTag: strPtr("keep")},
			want:     false,
		},
		{
			name:     "lacksTag keep matches bookmark with nil tags",
			bookmark: engine.Bookmark{ID: "52", CreatedAt: runTime},
			conds:    &config.Conditions{LacksTag: strPtr("keep")},
			want:     true,
		},
		{
			name:     "lacksTag keep matches bookmark with empty tags",
			bookmark: engine.Bookmark{ID: "53", CreatedAt: runTime, Tags: []string{}},
			conds:    &config.Conditions{LacksTag: strPtr("keep")},
			want:     true,
		},
		// --- case sensitivity ---
		{
			name:     "hasTag is case-sensitive",
			bookmark: engine.Bookmark{ID: "60", CreatedAt: runTime, Tags: []string{"read-later"}},
			conds:    &config.Conditions{HasTag: strPtr("Read-Later")},
			want:     false,
		},
		// --- combined AND with all six conditions ---
		{
			name: "AND: all six conditions match",
			bookmark: engine.Bookmark{
				ID: "c1", CreatedAt: runTime.Add(-31 * 24 * time.Hour),
				Source: "rss", Archived: true, Favourited: false,
				Tags: []string{"news", "read-later"},
			},
			conds: &config.Conditions{
				OlderThan: strPtr("30d"), Source: strPtr("rss"),
				Archived: boolPtr(true), Favourited: boolPtr(false),
				HasTag: strPtr("read-later"), LacksTag: strPtr("keep"),
			},
			want: true,
		},
		{
			name: "AND: five match but favourited fails",
			bookmark: engine.Bookmark{
				ID: "c2", CreatedAt: runTime.Add(-31 * 24 * time.Hour),
				Source: "rss", Archived: true, Favourited: false,
				Tags: []string{"news", "read-later"},
			},
			conds: &config.Conditions{
				OlderThan: strPtr("30d"), Source: strPtr("rss"),
				Archived: boolPtr(true), Favourited: boolPtr(true),
				HasTag: strPtr("read-later"), LacksTag: strPtr("keep"),
			},
			want: false,
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
