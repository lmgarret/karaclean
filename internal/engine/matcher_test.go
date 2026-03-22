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
			got := engine.MatchesConditions(tt.bookmark, tt.conds, runTime, nil)
			if got != tt.want {
				t.Errorf("MatchesConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchesExceptions(t *testing.T) {
	tests := []struct {
		name       string
		bookmark   engine.Bookmark
		exceptions *config.Exceptions
		want       bool
	}{
		// --- nil / empty ---
		{
			name:       "nil exceptions returns false (no protection)",
			bookmark:   engine.Bookmark{ID: "1"},
			exceptions: nil,
			want:       false,
		},
		{
			name:       "all-nil-fields exceptions returns false",
			bookmark:   engine.Bookmark{ID: "2"},
			exceptions: &config.Exceptions{},
			want:       false,
		},
		// --- Favourited ---
		{
			name:       "Favourited=true fires when bookmark is favourited",
			bookmark:   engine.Bookmark{ID: "3", Favourited: true},
			exceptions: &config.Exceptions{Favourited: boolPtr(true)},
			want:       true,
		},
		{
			name:       "Favourited=true does NOT fire when bookmark is not favourited",
			bookmark:   engine.Bookmark{ID: "4", Favourited: false},
			exceptions: &config.Exceptions{Favourited: boolPtr(true)},
			want:       false,
		},
		{
			name:       "Favourited=false fires when bookmark is not favourited",
			bookmark:   engine.Bookmark{ID: "5", Favourited: false},
			exceptions: &config.Exceptions{Favourited: boolPtr(false)},
			want:       true,
		},
		// --- HasTag ---
		{
			name:       "HasTag important fires when bookmark has that tag",
			bookmark:   engine.Bookmark{ID: "6", Tags: []string{"news", "important"}},
			exceptions: &config.Exceptions{HasTag: strPtr("important")},
			want:       true,
		},
		{
			name:       "HasTag important does NOT fire when bookmark lacks that tag",
			bookmark:   engine.Bookmark{ID: "7", Tags: []string{"news", "tech"}},
			exceptions: &config.Exceptions{HasTag: strPtr("important")},
			want:       false,
		},
		{
			name:       "HasTag important does NOT fire when bookmark tags is nil",
			bookmark:   engine.Bookmark{ID: "8"},
			exceptions: &config.Exceptions{HasTag: strPtr("important")},
			want:       false,
		},
		{
			name:       "HasTag is case-sensitive",
			bookmark:   engine.Bookmark{ID: "9", Tags: []string{"important"}},
			exceptions: &config.Exceptions{HasTag: strPtr("Important")},
			want:       false,
		},
		// --- HasNote ---
		{
			name:       "HasNote=true fires when bookmark has a non-empty note",
			bookmark:   engine.Bookmark{ID: "10", Note: "my note"},
			exceptions: &config.Exceptions{HasNote: boolPtr(true)},
			want:       true,
		},
		{
			name:       "HasNote=true does NOT fire when bookmark note is empty",
			bookmark:   engine.Bookmark{ID: "11", Note: ""},
			exceptions: &config.Exceptions{HasNote: boolPtr(true)},
			want:       false,
		},
		{
			name:       "HasNote=true does NOT fire when bookmark note is whitespace-only",
			bookmark:   engine.Bookmark{ID: "12", Note: "   "},
			exceptions: &config.Exceptions{HasNote: boolPtr(true)},
			want:       false,
		},
		// --- Archived ---
		{
			name:       "Archived=true fires when bookmark is archived",
			bookmark:   engine.Bookmark{ID: "13", Archived: true},
			exceptions: &config.Exceptions{Archived: boolPtr(true)},
			want:       true,
		},
		{
			name:       "Archived=true does NOT fire when bookmark is not archived",
			bookmark:   engine.Bookmark{ID: "14", Archived: false},
			exceptions: &config.Exceptions{Archived: boolPtr(true)},
			want:       false,
		},
		{
			name:       "Archived=false fires when bookmark is not archived",
			bookmark:   engine.Bookmark{ID: "15", Archived: false},
			exceptions: &config.Exceptions{Archived: boolPtr(false)},
			want:       true,
		},
		// --- OR semantics ---
		{
			name:       "OR: Favourited=true + HasTag=keep fires when only Favourited matches",
			bookmark:   engine.Bookmark{ID: "16", Favourited: true, Tags: []string{"news"}},
			exceptions: &config.Exceptions{Favourited: boolPtr(true), HasTag: strPtr("keep")},
			want:       true,
		},
		{
			name:       "OR: Favourited=true + HasTag=keep fires when only HasTag matches",
			bookmark:   engine.Bookmark{ID: "17", Favourited: false, Tags: []string{"keep"}},
			exceptions: &config.Exceptions{Favourited: boolPtr(true), HasTag: strPtr("keep")},
			want:       true,
		},
		{
			name:       "OR: Favourited=true + HasTag=keep does NOT fire when neither matches",
			bookmark:   engine.Bookmark{ID: "18", Favourited: false, Tags: []string{"news"}},
			exceptions: &config.Exceptions{Favourited: boolPtr(true), HasTag: strPtr("keep")},
			want:       false,
		},
		// --- All four set, only one fires ---
		{
			name: "all four exceptions set, only HasNote fires -> returns true",
			bookmark: engine.Bookmark{
				ID: "19", Favourited: false, Tags: []string{"news"},
				Note: "important note", Archived: false,
			},
			exceptions: &config.Exceptions{
				Favourited: boolPtr(true), HasTag: strPtr("keep"),
				HasNote: boolPtr(true), Archived: boolPtr(true),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.MatchesExceptions(tt.bookmark, tt.exceptions, nil)
			if got != tt.want {
				t.Errorf("MatchesExceptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchesConditions_InList(t *testing.T) {
	runTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	listSets := map[string]map[string]bool{
		"Read Later": {"bk-1": true, "bk-2": true},
		"RSS Feeds":  {"bk-3": true},
	}

	tests := []struct {
		name     string
		bookmark engine.Bookmark
		conds    *config.Conditions
		listSets map[string]map[string]bool
		want     bool
	}{
		{
			name:     "inList single match",
			bookmark: engine.Bookmark{ID: "bk-1", CreatedAt: runTime},
			conds:    &config.Conditions{InList: config.StringOrSlice{"Read Later"}},
			listSets: listSets,
			want:     true,
		},
		{
			name:     "inList single no match",
			bookmark: engine.Bookmark{ID: "bk-3", CreatedAt: runTime},
			conds:    &config.Conditions{InList: config.StringOrSlice{"Read Later"}},
			listSets: listSets,
			want:     false,
		},
		{
			name:     "inList OR semantics bookmark in second list",
			bookmark: engine.Bookmark{ID: "bk-3", CreatedAt: runTime},
			conds:    &config.Conditions{InList: config.StringOrSlice{"Read Later", "RSS Feeds"}},
			listSets: listSets,
			want:     true,
		},
		{
			name:     "inList OR semantics bookmark in neither list",
			bookmark: engine.Bookmark{ID: "bk-99", CreatedAt: runTime},
			conds:    &config.Conditions{InList: config.StringOrSlice{"Read Later", "RSS Feeds"}},
			listSets: listSets,
			want:     false,
		},
		{
			name:     "nil InList ignores check (backward compat)",
			bookmark: engine.Bookmark{ID: "bk-99", CreatedAt: runTime},
			conds:    &config.Conditions{},
			listSets: listSets,
			want:     true,
		},
		{
			name:     "nil listSets with non-nil InList returns false",
			bookmark: engine.Bookmark{ID: "bk-1", CreatedAt: runTime},
			conds:    &config.Conditions{InList: config.StringOrSlice{"Read Later"}},
			listSets: nil,
			want:     false,
		},
		{
			name:     "case_sensitive_no_match",
			bookmark: engine.Bookmark{ID: "bk-1", CreatedAt: runTime},
			conds:    &config.Conditions{InList: config.StringOrSlice{"read later"}},
			listSets: listSets,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.MatchesConditions(tt.bookmark, tt.conds, runTime, tt.listSets)
			if got != tt.want {
				t.Errorf("MatchesConditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchesExceptions_InList(t *testing.T) {
	listSets := map[string]map[string]bool{
		"Read Later": {"bk-1": true, "bk-2": true},
		"RSS Feeds":  {"bk-3": true},
	}

	tests := []struct {
		name       string
		bookmark   engine.Bookmark
		exceptions *config.Exceptions
		listSets   map[string]map[string]bool
		want       bool
	}{
		{
			name:       "inList protects bookmark in list",
			bookmark:   engine.Bookmark{ID: "bk-1"},
			exceptions: &config.Exceptions{InList: config.StringOrSlice{"Read Later"}},
			listSets:   listSets,
			want:       true,
		},
		{
			name:       "inList does not protect bookmark not in list",
			bookmark:   engine.Bookmark{ID: "bk-99"},
			exceptions: &config.Exceptions{InList: config.StringOrSlice{"Read Later"}},
			listSets:   listSets,
			want:       false,
		},
		{
			name:       "inList OR semantics bookmark in first list",
			bookmark:   engine.Bookmark{ID: "bk-1"},
			exceptions: &config.Exceptions{InList: config.StringOrSlice{"Read Later", "RSS Feeds"}},
			listSets:   listSets,
			want:       true,
		},
		{
			name:       "nil InList ignores check (backward compat)",
			bookmark:   engine.Bookmark{ID: "bk-1"},
			exceptions: &config.Exceptions{},
			listSets:   listSets,
			want:       false,
		},
		{
			name:       "case_sensitive_no_match",
			bookmark:   engine.Bookmark{ID: "bk-1"},
			exceptions: &config.Exceptions{InList: config.StringOrSlice{"read later"}},
			listSets:   listSets,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.MatchesExceptions(tt.bookmark, tt.exceptions, tt.listSets)
			if got != tt.want {
				t.Errorf("MatchesExceptions() = %v, want %v", got, tt.want)
			}
		})
	}
}
