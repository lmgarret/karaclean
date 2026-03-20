package engine_test

import (
	"testing"

	"github.com/lm/karaclean/internal/config"
	"github.com/lm/karaclean/internal/engine"
)

func TestFormatNotification(t *testing.T) {
	tests := []struct {
		name   string
		rs     *engine.RuleSummary
		dryRun bool
		want   string
	}{
		{
			name: "full rule with size",
			rs: &engine.RuleSummary{
				RuleName:   "old-rss",
				Deleted:    12,
				TotalBytes: 4404019,
				Archived:   3,
				Excepted:   1,
				Errors:     0,
			},
			dryRun: false,
			want:   "Summary:\ndeleted: 12 (4.2 MB)\narchived: 3\nexcepted: 1",
		},
		{
			name: "dry run body unchanged",
			rs: &engine.RuleSummary{
				RuleName:   "old-rss",
				Deleted:    12,
				TotalBytes: 4404019,
				Archived:   3,
				Excepted:   1,
				Errors:     0,
			},
			dryRun: true,
			want:   "Summary:\ndeleted: 12 (4.2 MB)\narchived: 3\nexcepted: 1",
		},
		{
			name: "deleted without size",
			rs: &engine.RuleSummary{
				RuleName:   "web-junk",
				Deleted:    5,
				TotalBytes: 0,
			},
			dryRun: false,
			want:   "Summary:\ndeleted: 5",
		},
		{
			name: "errors shown when > 0",
			rs: &engine.RuleSummary{
				RuleName: "failing-rule",
				Errors:   3,
			},
			dryRun: false,
			want:   "Summary:\nerrors: 3",
		},
		{
			name: "errors omitted when 0",
			rs: &engine.RuleSummary{
				RuleName: "clean-rule",
				Deleted:  2,
				Errors:   0,
			},
			dryRun: false,
			want:   "Summary:\ndeleted: 2",
		},
		{
			name: "archived omitted when 0",
			rs: &engine.RuleSummary{
				RuleName: "delete-only",
				Deleted:  3,
				Archived: 0,
			},
			dryRun: false,
			want:   "Summary:\ndeleted: 3",
		},
		{
			name: "deleted omitted when 0",
			rs: &engine.RuleSummary{
				RuleName:   "archive-only",
				Deleted:    0,
				TotalBytes: 0,
				Archived:   5,
			},
			dryRun: false,
			want:   "Summary:\narchived: 5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.FormatNotification(tt.rs, tt.dryRun)
			if got != tt.want {
				t.Errorf("FormatNotification() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestFormatNotificationTitle(t *testing.T) {
	tests := []struct {
		name     string
		ruleName string
		dryRun   bool
		want     string
	}{
		{
			name:     "normal title",
			ruleName: "old-rss",
			dryRun:   false,
			want:     "[karaclean] old-rss",
		},
		{
			name:     "dry run title",
			ruleName: "old-rss",
			dryRun:   true,
			want:     "[DRY-RUN] [karaclean] old-rss",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.FormatNotificationTitle(tt.ruleName, tt.dryRun)
			if got != tt.want {
				t.Errorf("FormatNotificationTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHasActivity(t *testing.T) {
	tests := []struct {
		name string
		rs   *engine.RuleSummary
		want bool
	}{
		{
			name: "deleted only",
			rs:   &engine.RuleSummary{Deleted: 1},
			want: true,
		},
		{
			name: "archived only",
			rs:   &engine.RuleSummary{Archived: 1},
			want: true,
		},
		{
			name: "errors only",
			rs:   &engine.RuleSummary{Errors: 1},
			want: true,
		},
		{
			name: "excepted only is silent",
			rs:   &engine.RuleSummary{Excepted: 5},
			want: false,
		},
		{
			name: "all zeros",
			rs:   &engine.RuleSummary{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rs.HasActivity()
			if got != tt.want {
				t.Errorf("HasActivity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveChannelURL(t *testing.T) {
	tests := []struct {
		name          string
		notifications *config.Notifications
		ruleNotify    *string
		want          string
	}{
		{
			name: "rule override returns override URL",
			notifications: &config.Notifications{
				Channels: map[string]config.NotificationChannel{
					"slack-team": {URL: "slack://TOKEN-A/TOKEN-B/TOKEN-C"},
					"my-ntfy":   {URL: "ntfy://ntfy.sh/karaclean-alerts"},
				},
				Default: "my-ntfy",
			},
			ruleNotify: ptrStr("slack-team"),
			want:       "slack://TOKEN-A/TOKEN-B/TOKEN-C",
		},
		{
			name: "nil rule notify uses default",
			notifications: &config.Notifications{
				Channels: map[string]config.NotificationChannel{
					"my-ntfy": {URL: "ntfy://ntfy.sh/karaclean-alerts"},
				},
				Default: "my-ntfy",
			},
			ruleNotify: nil,
			want:       "ntfy://ntfy.sh/karaclean-alerts",
		},
		{
			name: "nil rule notify and empty default returns empty",
			notifications: &config.Notifications{
				Channels: map[string]config.NotificationChannel{},
				Default:  "",
			},
			ruleNotify: nil,
			want:       "",
		},
		{
			name:          "nil notifications returns empty",
			notifications: nil,
			ruleNotify:    nil,
			want:          "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.ResolveChannelURL(tt.notifications, tt.ruleNotify)
			if got != tt.want {
				t.Errorf("ResolveChannelURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
