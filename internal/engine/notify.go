package engine

import (
	"fmt"
	"strings"

	"github.com/lm/karaclean/internal/config"
	"github.com/nicholas-fedor/shoutrrr"
	"github.com/nicholas-fedor/shoutrrr/pkg/types"
)

// RuleSummary records per-rule action counts for notification dispatch.
type RuleSummary struct {
	RuleName   string
	Archived   int
	Deleted    int
	Excepted   int
	Errors     int
	TotalBytes int64
}

// HasActivity returns true if this rule should trigger a notification.
// Per user decision: notify only when deleted > 0 OR archived > 0 OR errors > 0.
// Excepted-only rules are silent.
func (s *RuleSummary) HasActivity() bool {
	return s.Deleted > 0 || s.Archived > 0 || s.Errors > 0
}

// Notifier sends a notification message to a URL. Implementations:
// - ShoutrrrNotifier for production (sends via Shoutrrr library)
// - mockNotifier in tests (records calls)
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

// FormatNotification formats a per-rule notification message.
// Format per CONTEXT.md:
//
//	[karaclean] <rule-name>
//	deleted: N (X.X MB)   <- size only when TotalBytes > 0
//	archived: N           <- only when > 0
//	excepted: N           <- only when > 0
//	errors: N             <- only when > 0
//
// Dry-run prefix: [DRY-RUN] [karaclean] <rule-name>
func FormatNotification(rs *RuleSummary, dryRun bool) string {
	var b strings.Builder

	if dryRun {
		fmt.Fprintf(&b, "[DRY-RUN] ")
	}
	fmt.Fprintf(&b, "[karaclean] %s", rs.RuleName)

	if rs.Deleted > 0 {
		if rs.TotalBytes > 0 {
			fmt.Fprintf(&b, "\ndeleted: %d (%s)", rs.Deleted, HumanSize(rs.TotalBytes))
		} else {
			fmt.Fprintf(&b, "\ndeleted: %d", rs.Deleted)
		}
	}
	if rs.Archived > 0 {
		fmt.Fprintf(&b, "\narchived: %d", rs.Archived)
	}
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
