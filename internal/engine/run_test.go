package engine_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lm/karaclean/internal/config"
	"github.com/lm/karaclean/internal/engine"
)

func ptrBool(b bool) *bool   { return &b }
func ptrStr(s string) *string { return &s }

func TestRunSummary_String(t *testing.T) {
	s := engine.RunSummary{Archived: 2, Deleted: 1, NoMatch: 5, Excepted: 3, Errors: 0}
	got := s.String()
	for _, want := range []string{"archived=2", "deleted=1", "no_match=5", "excepted=3", "errors=0"} {
		if !strings.Contains(got, want) {
			t.Errorf("String() = %q, missing %q", got, want)
		}
	}
}

func TestRun(t *testing.T) {
	now := time.Now()
	oldCreatedAt := now.Add(-100 * 24 * time.Hour) // 100 days ago

	tests := []struct {
		name               string
		api                *mockAPI
		rules              []config.Rule
		dryRun             bool
		want               engine.RunSummary
		wantArchiveCalls   []string
		wantDeleteCalls    []string
	}{
		{
			name: "no bookmarks returns zero summary",
			api:  &mockAPI{listBookmarksRet: nil},
			rules: []config.Rule{
				{Name: "r1", Action: "archive"},
			},
			want: engine.RunSummary{},
		},
		{
			name: "no rules means all no_match",
			api: &mockAPI{listBookmarksRet: []engine.Bookmark{
				{ID: "bk-1"},
				{ID: "bk-2"},
			}},
			rules: nil,
			want:  engine.RunSummary{NoMatch: 2},
		},
		{
			name: "archive action increments Archived",
			api: &mockAPI{listBookmarksRet: []engine.Bookmark{
				{ID: "bk-1", CreatedAt: oldCreatedAt},
			}},
			rules: []config.Rule{
				{Name: "old-stuff", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "archive"},
			},
			want:             engine.RunSummary{Archived: 1},
			wantArchiveCalls: []string{"bk-1"},
		},
		{
			name: "delete action increments Deleted",
			api: &mockAPI{listBookmarksRet: []engine.Bookmark{
				{ID: "bk-1", CreatedAt: oldCreatedAt},
			}},
			rules: []config.Rule{
				{Name: "old-stuff", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "delete"},
			},
			want:            engine.RunSummary{Deleted: 1},
			wantDeleteCalls: []string{"bk-1"},
		},
		{
			name: "excepted bookmark increments Excepted",
			api: &mockAPI{listBookmarksRet: []engine.Bookmark{
				{ID: "bk-1", CreatedAt: oldCreatedAt, Favourited: true},
			}},
			rules: []config.Rule{
				{
					Name:       "old-stuff",
					Conditions: &config.Conditions{OlderThan: ptrStr("30d")},
					Unless:     &config.Exceptions{Favourited: ptrBool(true)},
					Action:     "archive",
				},
			},
			want: engine.RunSummary{Excepted: 1},
		},
		{
			name: "first match wins stops after first rule",
			api: &mockAPI{listBookmarksRet: []engine.Bookmark{
				{ID: "bk-1", CreatedAt: oldCreatedAt},
			}},
			rules: []config.Rule{
				{Name: "rule1", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "archive"},
				{Name: "rule2", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "delete"},
			},
			want:             engine.RunSummary{Archived: 1},
			wantArchiveCalls: []string{"bk-1"},
		},
		{
			name: "action error increments Errors",
			api: &mockAPI{
				listBookmarksRet:   []engine.Bookmark{{ID: "bk-1", CreatedAt: oldCreatedAt}},
				archiveBookmarkErr: errors.New("api down"),
			},
			rules: []config.Rule{
				{Name: "old-stuff", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "archive"},
			},
			want:             engine.RunSummary{Errors: 1},
			wantArchiveCalls: []string{"bk-1"},
		},
		{
			name:   "dry run passes through and counts",
			dryRun: true,
			api: &mockAPI{listBookmarksRet: []engine.Bookmark{
				{ID: "bk-1", CreatedAt: oldCreatedAt},
			}},
			rules: []config.Rule{
				{Name: "old-stuff", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "archive"},
			},
			want: engine.RunSummary{Archived: 1},
			// In dry-run, ExecuteAction skips the API call, so no archiveBookmarkCalls
		},
		{
			name: "mixed scenario",
			api: &mockAPI{listBookmarksRet: []engine.Bookmark{
				{ID: "bk-1", CreatedAt: oldCreatedAt},                           // matches rule1 -> archive
				{ID: "bk-2", CreatedAt: oldCreatedAt, Favourited: true},         // matches conditions but excepted
				{ID: "bk-3", CreatedAt: time.Now()},                             // matches no rule
			}},
			rules: []config.Rule{
				{
					Name:       "old-stuff",
					Conditions: &config.Conditions{OlderThan: ptrStr("30d")},
					Unless:     &config.Exceptions{Favourited: ptrBool(true)},
					Action:     "archive",
				},
			},
			want:             engine.RunSummary{Archived: 1, Excepted: 1, NoMatch: 1},
			wantArchiveCalls: []string{"bk-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Run(context.Background(), tt.api, tt.rules, tt.dryRun)
			if err != nil {
				t.Fatalf("Run() returned unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Run() = %+v, want %+v", got, tt.want)
			}
			if tt.wantArchiveCalls != nil {
				if len(tt.api.archiveBookmarkCalls) != len(tt.wantArchiveCalls) {
					t.Errorf("archiveBookmarkCalls = %v, want %v", tt.api.archiveBookmarkCalls, tt.wantArchiveCalls)
				} else {
					for i, id := range tt.wantArchiveCalls {
						if tt.api.archiveBookmarkCalls[i] != id {
							t.Errorf("archiveBookmarkCalls[%d] = %q, want %q", i, tt.api.archiveBookmarkCalls[i], id)
						}
					}
				}
			}
			if tt.wantDeleteCalls != nil {
				if len(tt.api.deleteBookmarkCalls) != len(tt.wantDeleteCalls) {
					t.Errorf("deleteBookmarkCalls = %v, want %v", tt.api.deleteBookmarkCalls, tt.wantDeleteCalls)
				} else {
					for i, id := range tt.wantDeleteCalls {
						if tt.api.deleteBookmarkCalls[i] != id {
							t.Errorf("deleteBookmarkCalls[%d] = %q, want %q", i, tt.api.deleteBookmarkCalls[i], id)
						}
					}
				}
			}
			// For "first match wins", also verify no delete calls happened
			if tt.name == "first match wins stops after first rule" {
				if len(tt.api.deleteBookmarkCalls) != 0 {
					t.Errorf("expected no delete calls, got %v", tt.api.deleteBookmarkCalls)
				}
			}
			// For "excepted bookmark", verify no action calls
			if tt.name == "excepted bookmark increments Excepted" {
				if len(tt.api.archiveBookmarkCalls) != 0 {
					t.Errorf("expected no archive calls for excepted bookmark, got %v", tt.api.archiveBookmarkCalls)
				}
				if len(tt.api.deleteBookmarkCalls) != 0 {
					t.Errorf("expected no delete calls for excepted bookmark, got %v", tt.api.deleteBookmarkCalls)
				}
			}
			// For "dry run", verify no archive calls on the mock
			if tt.name == "dry run passes through and counts" {
				if len(tt.api.archiveBookmarkCalls) != 0 {
					t.Errorf("expected no archive calls in dry-run, got %v", tt.api.archiveBookmarkCalls)
				}
			}
		})
	}
}

func TestRun_ListBookmarksError(t *testing.T) {
	api := &mockAPI{listBookmarksErr: errors.New("api unreachable")}
	rules := []config.Rule{{Name: "r1", Action: "archive"}}

	got, err := engine.Run(context.Background(), api, rules, false)
	if err == nil {
		t.Fatal("expected error from Run(), got nil")
	}
	if !strings.Contains(err.Error(), "api unreachable") {
		t.Errorf("error %q does not contain cause", err.Error())
	}
	zero := engine.RunSummary{}
	if got != zero {
		t.Errorf("expected zero RunSummary on error, got %+v", got)
	}
}
