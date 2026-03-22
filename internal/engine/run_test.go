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
	s := engine.RunSummary{Archived: 2, Deleted: 1, NoMatch: 5, Excepted: 3, Errors: 0, TotalBytes: 1048576}
	got := s.String()
	for _, want := range []string{"archived=2", "deleted=1", "no_match=5", "excepted=3", "errors=0", "total_size=1.0 MB"} {
		if !strings.Contains(got, want) {
			t.Errorf("String() = %q, missing %q", got, want)
		}
	}
}

func TestRunSummary_String_ZeroBytes(t *testing.T) {
	s := engine.RunSummary{Archived: 1, Deleted: 0, NoMatch: 0, Excepted: 0, Errors: 0, TotalBytes: 0}
	got := s.String()
	if strings.Contains(got, "total_size") {
		t.Errorf("String() = %q, should not contain 'total_size' when TotalBytes=0", got)
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
		wantTotalBytes     int64
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
				{ID: "bk-1", CreatedAt: oldCreatedAt, Size: 5000},
			}},
			rules: []config.Rule{
				{Name: "old-stuff", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "archive"},
			},
			want:             engine.RunSummary{Archived: 1, TotalBytes: 5000},
			wantArchiveCalls: []string{"bk-1"},
			wantTotalBytes:   5000,
		},
		{
			name: "delete action increments Deleted",
			api: &mockAPI{listBookmarksRet: []engine.Bookmark{
				{ID: "bk-1", CreatedAt: oldCreatedAt, Size: 2048},
			}},
			rules: []config.Rule{
				{Name: "old-stuff", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "delete"},
			},
			want:            engine.RunSummary{Deleted: 1, TotalBytes: 2048},
			wantDeleteCalls: []string{"bk-1"},
			wantTotalBytes:  2048,
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
			name: "action error increments Errors not TotalBytes",
			api: &mockAPI{
				listBookmarksRet:   []engine.Bookmark{{ID: "bk-1", CreatedAt: oldCreatedAt, Size: 9999}},
				archiveBookmarkErr: errors.New("api down"),
			},
			rules: []config.Rule{
				{Name: "old-stuff", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "archive"},
			},
			want:             engine.RunSummary{Errors: 1},
			wantArchiveCalls: []string{"bk-1"},
			wantTotalBytes:   0,
		},
		{
			name:   "dry run passes through and counts",
			dryRun: true,
			api: &mockAPI{listBookmarksRet: []engine.Bookmark{
				{ID: "bk-1", CreatedAt: oldCreatedAt, Size: 3072},
			}},
			rules: []config.Rule{
				{Name: "old-stuff", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "archive"},
			},
			want:           engine.RunSummary{Archived: 1, TotalBytes: 3072},
			wantTotalBytes: 3072,
			// In dry-run, ExecuteAction skips the API call, so no archiveBookmarkCalls
		},
		{
			name:   "per-rule dryRun true overrides global false",
			dryRun: false,
			api: &mockAPI{listBookmarksRet: []engine.Bookmark{
				{ID: "bk-1", CreatedAt: oldCreatedAt},
			}},
			rules: []config.Rule{
				{Name: "old-stuff", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "archive", DryRun: ptrBool(true)},
			},
			want: engine.RunSummary{Archived: 1},
			// Per-rule dryRun=true means no API calls despite global dryRun=false
		},
		{
			name:   "per-rule dryRun false overrides global true",
			dryRun: true,
			api: &mockAPI{listBookmarksRet: []engine.Bookmark{
				{ID: "bk-1", CreatedAt: oldCreatedAt},
			}},
			rules: []config.Rule{
				{Name: "old-stuff", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "archive", DryRun: ptrBool(false)},
			},
			want:             engine.RunSummary{Archived: 1},
			wantArchiveCalls: []string{"bk-1"},
		},
		{
			name:   "per-rule dryRun nil inherits global true",
			dryRun: true,
			api: &mockAPI{listBookmarksRet: []engine.Bookmark{
				{ID: "bk-1", CreatedAt: oldCreatedAt},
			}},
			rules: []config.Rule{
				{Name: "old-stuff", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "archive", DryRun: nil},
			},
			want: engine.RunSummary{Archived: 1},
			// nil DryRun inherits global=true, so no API calls
		},
		{
			name: "mixed scenario",
			api: &mockAPI{listBookmarksRet: []engine.Bookmark{
				{ID: "bk-1", CreatedAt: oldCreatedAt, Size: 4096},               // matches rule1 -> archive
				{ID: "bk-2", CreatedAt: oldCreatedAt, Favourited: true, Size: 8192}, // matches conditions but excepted
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
			want:             engine.RunSummary{Archived: 1, Excepted: 1, NoMatch: 1, TotalBytes: 4096},
			wantArchiveCalls: []string{"bk-1"},
			wantTotalBytes:   4096,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Run(context.Background(), tt.api, tt.rules, tt.dryRun, nil, nil)
			if err != nil {
				t.Fatalf("Run() returned unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Run() = %+v, want %+v", got, tt.want)
			}
			assertCalls(t, "archiveBookmarkCalls", tt.api.archiveBookmarkCalls, tt.wantArchiveCalls)
			assertCalls(t, "deleteBookmarkCalls", tt.api.deleteBookmarkCalls, tt.wantDeleteCalls)
			assertNoActionCalls(t, tt.name, tt.api)
		})
	}
}

func assertCalls(t *testing.T, label string, got, want []string) {
	t.Helper()
	if want == nil {
		return
	}
	if len(got) != len(want) {
		t.Errorf("%s = %v, want %v", label, got, want)
		return
	}
	for i, id := range want {
		if got[i] != id {
			t.Errorf("%s[%d] = %q, want %q", label, i, got[i], id)
		}
	}
}

func assertNoActionCalls(t *testing.T, name string, api *mockAPI) {
	t.Helper()
	switch name {
	case "first match wins stops after first rule":
		if len(api.deleteBookmarkCalls) != 0 {
			t.Errorf("expected no delete calls, got %v", api.deleteBookmarkCalls)
		}
	case "excepted bookmark increments Excepted":
		if len(api.archiveBookmarkCalls) != 0 {
			t.Errorf("expected no archive calls for excepted bookmark, got %v", api.archiveBookmarkCalls)
		}
		if len(api.deleteBookmarkCalls) != 0 {
			t.Errorf("expected no delete calls for excepted bookmark, got %v", api.deleteBookmarkCalls)
		}
	case "dry run passes through and counts":
		if len(api.archiveBookmarkCalls) != 0 {
			t.Errorf("expected no archive calls in dry-run, got %v", api.archiveBookmarkCalls)
		}
	case "per-rule dryRun true overrides global false":
		if len(api.archiveBookmarkCalls) != 0 {
			t.Errorf("expected no archive calls when per-rule dryRun=true, got %v", api.archiveBookmarkCalls)
		}
	case "per-rule dryRun nil inherits global true":
		if len(api.archiveBookmarkCalls) != 0 {
			t.Errorf("expected no archive calls when inheriting global dryRun=true, got %v", api.archiveBookmarkCalls)
		}
	}
}

func TestResolveRuleDryRun(t *testing.T) {
	tests := []struct {
		name        string
		ruleDryRun  *bool
		globalDryRun bool
		want        bool
	}{
		{"nil inherits global false", nil, false, false},
		{"nil inherits global true", nil, true, true},
		{"ptr true overrides global false", ptrBool(true), false, true},
		{"ptr false overrides global true", ptrBool(false), true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.ResolveRuleDryRun(tt.ruleDryRun, tt.globalDryRun)
			if got != tt.want {
				t.Errorf("ResolveRuleDryRun(%v, %v) = %v, want %v", tt.ruleDryRun, tt.globalDryRun, got, tt.want)
			}
		})
	}
}

// mockNotifier records Send calls for test assertions.
type mockNotifier struct {
	calls []notifierCall
	err   error // error to return from Send
}

type notifierCall struct {
	url     string
	message string
	title   string
}

func (m *mockNotifier) Send(url, message, title string) error {
	m.calls = append(m.calls, notifierCall{url: url, message: message, title: title})
	return m.err
}

// testNotifications builds a *config.Notifications for test use.
func testNotifications(channels map[string]string, defaultCh string) *config.Notifications {
	n := &config.Notifications{
		Channels: make(map[string]config.NotificationChannel),
		Default:  defaultCh,
	}
	for name, url := range channels {
		n.Channels[name] = config.NotificationChannel{URL: url}
	}
	return n
}

func TestRunNotification_ActiveRule(t *testing.T) {
	now := time.Now()
	old := now.Add(-100 * 24 * time.Hour)
	mn := &mockNotifier{}
	notifs := testNotifications(map[string]string{"alerts": "ntfy://ntfy.sh/test"}, "alerts")

	api := &mockAPI{listBookmarksRet: []engine.Bookmark{
		{ID: "bk-1", CreatedAt: old, Size: 4096},
	}}
	rules := []config.Rule{
		{Name: "cleanup", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "delete"},
	}

	_, err := engine.Run(context.Background(), api, rules, false, notifs, mn)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if len(mn.calls) != 1 {
		t.Fatalf("expected 1 notification call, got %d", len(mn.calls))
	}
	if mn.calls[0].url != "ntfy://ntfy.sh/test" {
		t.Errorf("url = %q, want ntfy://ntfy.sh/test", mn.calls[0].url)
	}
	if !strings.Contains(mn.calls[0].message, "Summary:") {
		t.Errorf("message %q missing Summary:", mn.calls[0].message)
	}
	if !strings.Contains(mn.calls[0].message, "deleted:") {
		t.Errorf("message %q missing deleted:", mn.calls[0].message)
	}
}

func TestRunNotification_Override(t *testing.T) {
	now := time.Now()
	old := now.Add(-100 * 24 * time.Hour)
	mn := &mockNotifier{}
	notifs := testNotifications(map[string]string{
		"default-ch": "ntfy://ntfy.sh/default",
		"override":   "ntfy://ntfy.sh/override",
	}, "default-ch")

	api := &mockAPI{listBookmarksRet: []engine.Bookmark{
		{ID: "bk-1", CreatedAt: old},
	}}
	rules := []config.Rule{
		{Name: "r1", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "delete", Notify: ptrStr("override")},
	}

	_, err := engine.Run(context.Background(), api, rules, false, notifs, mn)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if len(mn.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mn.calls))
	}
	if mn.calls[0].url != "ntfy://ntfy.sh/override" {
		t.Errorf("url = %q, want override URL", mn.calls[0].url)
	}
}

func TestRunNotification_Silent_NoActivity(t *testing.T) {
	now := time.Now()
	old := now.Add(-100 * 24 * time.Hour)
	mn := &mockNotifier{}
	notifs := testNotifications(map[string]string{"alerts": "ntfy://ntfy.sh/test"}, "alerts")

	api := &mockAPI{listBookmarksRet: []engine.Bookmark{
		{ID: "bk-1", CreatedAt: old, Favourited: true},
	}}
	rules := []config.Rule{
		{
			Name:       "excepted-only",
			Conditions: &config.Conditions{OlderThan: ptrStr("30d")},
			Unless:     &config.Exceptions{Favourited: ptrBool(true)},
			Action:     "delete",
		},
	}

	_, err := engine.Run(context.Background(), api, rules, false, notifs, mn)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if len(mn.calls) != 0 {
		t.Errorf("expected no notification for excepted-only rule, got %d calls", len(mn.calls))
	}
}

func TestRunNotification_Silent_NoChannel(t *testing.T) {
	now := time.Now()
	old := now.Add(-100 * 24 * time.Hour)
	mn := &mockNotifier{}
	notifs := testNotifications(map[string]string{"alerts": "ntfy://ntfy.sh/test"}, "") // empty default

	api := &mockAPI{listBookmarksRet: []engine.Bookmark{
		{ID: "bk-1", CreatedAt: old},
	}}
	rules := []config.Rule{
		{Name: "no-channel", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "delete"},
	}

	_, err := engine.Run(context.Background(), api, rules, false, notifs, mn)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if len(mn.calls) != 0 {
		t.Errorf("expected no notification with no channel, got %d calls", len(mn.calls))
	}
}

func TestRunNotification_NilNotifier(t *testing.T) {
	now := time.Now()
	old := now.Add(-100 * 24 * time.Hour)
	notifs := testNotifications(map[string]string{"alerts": "ntfy://ntfy.sh/test"}, "alerts")

	api := &mockAPI{listBookmarksRet: []engine.Bookmark{
		{ID: "bk-1", CreatedAt: old},
	}}
	rules := []config.Rule{
		{Name: "r1", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "delete"},
	}

	// Should not panic with nil notifier
	summary, err := engine.Run(context.Background(), api, rules, false, notifs, nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if summary.Deleted != 1 {
		t.Errorf("summary.Deleted = %d, want 1", summary.Deleted)
	}
}

func TestRunNotification_NilNotifications(t *testing.T) {
	now := time.Now()
	old := now.Add(-100 * 24 * time.Hour)
	mn := &mockNotifier{}

	api := &mockAPI{listBookmarksRet: []engine.Bookmark{
		{ID: "bk-1", CreatedAt: old},
	}}
	rules := []config.Rule{
		{Name: "r1", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "delete"},
	}

	// nil notifications, non-nil notifier: should not send anything
	_, err := engine.Run(context.Background(), api, rules, false, nil, mn)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if len(mn.calls) != 0 {
		t.Errorf("expected no notification with nil notifications, got %d calls", len(mn.calls))
	}
}

func TestRunNotification_FailureNonFatal(t *testing.T) {
	now := time.Now()
	old := now.Add(-100 * 24 * time.Hour)
	mn := &mockNotifier{err: errors.New("send failed")}
	notifs := testNotifications(map[string]string{"alerts": "ntfy://ntfy.sh/test"}, "alerts")

	api := &mockAPI{listBookmarksRet: []engine.Bookmark{
		{ID: "bk-1", CreatedAt: old},
	}}
	rules := []config.Rule{
		{Name: "r1", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "delete"},
	}

	summary, err := engine.Run(context.Background(), api, rules, false, notifs, mn)
	if err != nil {
		t.Fatalf("Run() should not return error on notification failure, got: %v", err)
	}
	if summary.Deleted != 1 {
		t.Errorf("summary.Deleted = %d, want 1", summary.Deleted)
	}
}

func TestRunNotification_DryRun(t *testing.T) {
	now := time.Now()
	old := now.Add(-100 * 24 * time.Hour)
	mn := &mockNotifier{}
	notifs := testNotifications(map[string]string{"alerts": "ntfy://ntfy.sh/test"}, "alerts")

	api := &mockAPI{listBookmarksRet: []engine.Bookmark{
		{ID: "bk-1", CreatedAt: old},
	}}
	rules := []config.Rule{
		{Name: "r1", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "delete"},
	}

	_, err := engine.Run(context.Background(), api, rules, true, notifs, mn)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if len(mn.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mn.calls))
	}
	if !strings.Contains(mn.calls[0].message, "Summary:") {
		t.Errorf("message %q missing Summary:", mn.calls[0].message)
	}
}

func TestRun_ListBookmarksError(t *testing.T) {
	api := &mockAPI{listBookmarksErr: errors.New("api unreachable")}
	rules := []config.Rule{{Name: "r1", Action: "archive"}}

	got, err := engine.Run(context.Background(), api, rules, false, nil, nil)
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

func TestPreloadListSets_NoInList(t *testing.T) {
	api := &mockAPI{}
	rules := []config.Rule{
		{Name: "r1", Conditions: &config.Conditions{OlderThan: ptrStr("30d")}, Action: "archive"},
	}
	got, err := engine.PreloadListSets(context.Background(), api, rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil map for no inList rules, got %v", got)
	}
}

func TestPreloadListSets_ResolvesAndFetches(t *testing.T) {
	api := &mockAPI{
		listListsRet: []engine.ListInfo{
			{ID: "list-1", Name: "Read Later"},
			{ID: "list-2", Name: "RSS Feeds"},
		},
		getListBookmarksRet: map[string][]string{
			"list-1": {"bk-1", "bk-3"},
			"list-2": {"bk-2"},
		},
	}
	rules := []config.Rule{
		{Name: "r1", Conditions: &config.Conditions{InList: config.StringOrSlice{"Read Later"}}, Action: "archive"},
		{Name: "r2", Unless: &config.Exceptions{InList: config.StringOrSlice{"RSS Feeds"}}, Action: "delete"},
	}
	got, err := engine.PreloadListSets(context.Background(), api, rules)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil map")
	}
	// Check Read Later
	if !got["Read Later"]["bk-1"] || !got["Read Later"]["bk-3"] {
		t.Errorf("Read Later set = %v, want bk-1 and bk-3", got["Read Later"])
	}
	// Check RSS Feeds
	if !got["RSS Feeds"]["bk-2"] {
		t.Errorf("RSS Feeds set = %v, want bk-2", got["RSS Feeds"])
	}
}

func TestPreloadListSets_ListListsError(t *testing.T) {
	api := &mockAPI{listListsErr: errors.New("api down")}
	rules := []config.Rule{
		{Name: "r1", Conditions: &config.Conditions{InList: config.StringOrSlice{"Read Later"}}, Action: "archive"},
	}
	_, err := engine.PreloadListSets(context.Background(), api, rules)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "preloading lists") {
		t.Errorf("error %q does not contain 'preloading lists'", err.Error())
	}
}

func TestPreloadListSets_GetListBookmarksError(t *testing.T) {
	api := &mockAPI{
		listListsRet:       []engine.ListInfo{{ID: "list-1", Name: "Read Later"}},
		getListBookmarksErr: errors.New("fetch failed"),
	}
	rules := []config.Rule{
		{Name: "r1", Conditions: &config.Conditions{InList: config.StringOrSlice{"Read Later"}}, Action: "archive"},
	}
	_, err := engine.PreloadListSets(context.Background(), api, rules)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "preloading list") {
		t.Errorf("error %q does not contain 'preloading list'", err.Error())
	}
}

func TestRun_InListCondition(t *testing.T) {
	now := time.Now()
	old := now.Add(-100 * 24 * time.Hour)

	api := &mockAPI{
		listBookmarksRet: []engine.Bookmark{
			{ID: "bk-1", CreatedAt: old},
			{ID: "bk-2", CreatedAt: old},
		},
		listListsRet: []engine.ListInfo{
			{ID: "list-2", Name: "RSS Feeds"},
		},
		getListBookmarksRet: map[string][]string{
			"list-2": {"bk-2"},
		},
	}

	rules := []config.Rule{
		{
			Name:       "rss-cleanup",
			Conditions: &config.Conditions{OlderThan: ptrStr("30d"), InList: config.StringOrSlice{"RSS Feeds"}},
			Action:     "delete",
		},
	}

	got, err := engine.Run(context.Background(), api, rules, false, nil, nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	// bk-2 is in RSS Feeds -> deleted; bk-1 is not in RSS Feeds -> no match
	if got.Deleted != 1 {
		t.Errorf("Deleted = %d, want 1", got.Deleted)
	}
	if got.NoMatch != 1 {
		t.Errorf("NoMatch = %d, want 1", got.NoMatch)
	}
	if len(api.deleteBookmarkCalls) != 1 || api.deleteBookmarkCalls[0] != "bk-2" {
		t.Errorf("deleteBookmarkCalls = %v, want [bk-2]", api.deleteBookmarkCalls)
	}
}

func TestRun_InListException(t *testing.T) {
	now := time.Now()
	old := now.Add(-100 * 24 * time.Hour)

	api := &mockAPI{
		listBookmarksRet: []engine.Bookmark{
			{ID: "bk-1", CreatedAt: old},
			{ID: "bk-3", CreatedAt: old},
		},
		listListsRet: []engine.ListInfo{
			{ID: "list-1", Name: "Read Later"},
		},
		getListBookmarksRet: map[string][]string{
			"list-1": {"bk-1", "bk-3"},
		},
	}

	rules := []config.Rule{
		{
			Name:       "old-cleanup",
			Conditions: &config.Conditions{OlderThan: ptrStr("30d")},
			Unless:     &config.Exceptions{InList: config.StringOrSlice{"Read Later"}},
			Action:     "delete",
		},
	}

	got, err := engine.Run(context.Background(), api, rules, false, nil, nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	// Both bk-1 and bk-3 are in Read Later -> both excepted
	if got.Excepted != 2 {
		t.Errorf("Excepted = %d, want 2", got.Excepted)
	}
	if got.Deleted != 0 {
		t.Errorf("Deleted = %d, want 0", got.Deleted)
	}
}
