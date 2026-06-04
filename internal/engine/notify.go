package engine

import (
	"fmt"
	"strings"

	"github.com/lmgarret/karaclean/internal/config"
	"github.com/nicholas-fedor/shoutrrr"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

// RuleSummary records per-rule action counts for notification dispatch.
type RuleSummary struct {
	RuleName      string
	Archived      int
	Unarchived    int
	Deleted       int
	Tagged        int
	Untagged      int
	Favourited    int
	Unfavourited  int
	Excepted      int
	Errors        int
	DeletedBytes  int64
	ArchivedBytes int64
}

// record increments the counter for a successfully executed action. Byte sizes
// are tracked only for archive/delete, where reclaimable content size is meaningful.
func (s *RuleSummary) record(action string, size int64) {
	switch action {
	case "archive":
		s.Archived++
		s.ArchivedBytes += size
	case "unarchive":
		s.Unarchived++
	case "delete":
		s.Deleted++
		s.DeletedBytes += size
	case "tag":
		s.Tagged++
	case "untag":
		s.Untagged++
	case "favourite":
		s.Favourited++
	case "unfavourite":
		s.Unfavourited++
	}
}

// HasActivity returns true if this rule should trigger a notification.
// Notify when any action was taken or an error occurred; excepted-only rules are silent.
func (s *RuleSummary) HasActivity() bool {
	actions := s.Archived + s.Unarchived + s.Deleted + s.Tagged + s.Untagged + s.Favourited + s.Unfavourited
	return actions > 0 || s.Errors > 0
}

// Notifier sends a notification message to a URL. Implementations:
// - ShoutrrrNotifier for production (sends via Shoutrrr library)
// - mockNotifier in tests (records calls).
type Notifier interface {
	Send(url, message, title string) error
}

// ShoutrrrNotifier sends notifications via the Shoutrrr library.
type ShoutrrrNotifier struct{}

// Send dispatches a notification using shoutrrr.CreateSender with title params.
func (n *ShoutrrrNotifier) Send(url, message, title string) error {
	sender, err := shoutrrr.CreateSender(url)
	if err != nil {
		return fmt.Errorf("creating sender: %w", err)
	}
	params := types.Params{}
	params.SetTitle(title)
	errs := sender.Send(message, &params)
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

// FormatNotification formats a per-rule notification message body.
// The body contains only stats prefixed with "Summary:" — the rule name and
// dry-run indicator are conveyed via the separate title (FormatNotificationTitle).
//
//	Summary:
//	deleted: N (X.X MB)   <- size shown when bytes > 0
//	archived: N (X.X MB)
//	tagged: N             <- each action line only when > 0
//	excepted: N           <- only when > 0
//	errors: N             <- only when > 0
func FormatNotification(rs *RuleSummary, _ bool) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Summary:")

	appendLine := func(label string, count int, bytes int64) {
		if count <= 0 {
			return
		}
		if bytes > 0 {
			fmt.Fprintf(&b, "\n%s: %d (%s)", label, count, HumanSize(bytes))
		} else {
			fmt.Fprintf(&b, "\n%s: %d", label, count)
		}
	}

	appendLine("deleted", rs.Deleted, rs.DeletedBytes)
	appendLine("archived", rs.Archived, rs.ArchivedBytes)
	appendLine("unarchived", rs.Unarchived, 0)
	appendLine("tagged", rs.Tagged, 0)
	appendLine("untagged", rs.Untagged, 0)
	appendLine("favourited", rs.Favourited, 0)
	appendLine("unfavourited", rs.Unfavourited, 0)

	if rs.Excepted > 0 {
		fmt.Fprintf(&b, "\nexcepted: %d", rs.Excepted)
	}
	if rs.Errors > 0 {
		fmt.Fprintf(&b, "\nerrors: %d", rs.Errors)
	}

	return b.String()
}

// FormatNotificationTitle returns the title string for services that support it.
// Format: "[karaclean] <rule-name>" or "[DRY-RUN] [karaclean] <rule-name>".
func FormatNotificationTitle(ruleName string, dryRun bool) string {
	if dryRun {
		return fmt.Sprintf("[DRY-RUN] [karaclean] %s", ruleName)
	}
	return fmt.Sprintf("[karaclean] %s", ruleName)
}

// ResolveChannelURL determines the notification URL for a rule.
// Priority: rule.Notify override > notifications.default > "" (silent).
// Returns "" if no channel is configured (silent, no error).
func ResolveChannelURL(notifications *config.Notifications, ruleNotify *string) string {
	if notifications == nil {
		return ""
	}
	name := notifications.Default
	if ruleNotify != nil {
		name = *ruleNotify
	}
	if name == "" {
		return ""
	}
	ch, ok := notifications.Channels[name]
	if !ok {
		return "" // should not happen if validation passed
	}
	return ch.URL
}
