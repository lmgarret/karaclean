package engine

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lmgarret/karaclean/internal/config"
)

// RunSummary records the outcome of a single Run() invocation.
// The per-action counters plus NoMatch + Excepted + Errors sum to the total
// bookmark count.
type RunSummary struct {
	Archived     int   `json:"archived"`
	Unarchived   int   `json:"unarchived"`
	Deleted      int   `json:"deleted"`
	Tagged       int   `json:"tagged"`
	Untagged     int   `json:"untagged"`
	Favourited   int   `json:"favourited"`
	Unfavourited int   `json:"unfavourited"`
	NoMatch      int   `json:"no_match"`
	Excepted     int   `json:"excepted"`
	Errors       int   `json:"errors"`
	TotalBytes   int64 `json:"total_bytes"`
}

// record increments the counter for a successfully executed action and adds
// its byte size to the running total.
func (s *RunSummary) record(action string, size int64) {
	switch action {
	case "archive":
		s.Archived++
	case "unarchive":
		s.Unarchived++
	case "delete":
		s.Deleted++
	case "tag":
		s.Tagged++
	case "untag":
		s.Untagged++
	case "favourite":
		s.Favourited++
	case "unfavourite":
		s.Unfavourited++
	}
	s.TotalBytes += size
}

// String returns a key=value summary suitable for structured log output.
// archive, delete and the run totals are always shown; the remaining action
// counters are included only when non-zero to keep common output compact.
func (s RunSummary) String() string {
	parts := []string{fmt.Sprintf("archived=%d", s.Archived)}
	if s.Unarchived > 0 {
		parts = append(parts, fmt.Sprintf("unarchived=%d", s.Unarchived))
	}
	parts = append(parts, fmt.Sprintf("deleted=%d", s.Deleted))
	if s.Tagged > 0 {
		parts = append(parts, fmt.Sprintf("tagged=%d", s.Tagged))
	}
	if s.Untagged > 0 {
		parts = append(parts, fmt.Sprintf("untagged=%d", s.Untagged))
	}
	if s.Favourited > 0 {
		parts = append(parts, fmt.Sprintf("favourited=%d", s.Favourited))
	}
	if s.Unfavourited > 0 {
		parts = append(parts, fmt.Sprintf("unfavourited=%d", s.Unfavourited))
	}
	parts = append(parts,
		fmt.Sprintf("no_match=%d", s.NoMatch),
		fmt.Sprintf("excepted=%d", s.Excepted),
		fmt.Sprintf("errors=%d", s.Errors),
	)
	base := strings.Join(parts, " ")
	if s.TotalBytes > 0 {
		base += fmt.Sprintf(" total_size=%s", HumanSize(s.TotalBytes))
	}
	return base
}

// PreloadListSets fetches list membership data from the API for all lists
// referenced in rule conditions or exceptions. Returns nil if no rules use inList (D-05).
func PreloadListSets(ctx context.Context, api KarakeepAPI, rules []config.Rule) (map[string]map[string]bool, error) {
	nameSet := make(map[string]bool)
	for _, r := range rules {
		if r.Conditions != nil {
			for _, name := range r.Conditions.InList {
				nameSet[name] = true
			}
		}
		if r.Unless != nil {
			for _, name := range r.Unless.InList {
				nameSet[name] = true
			}
		}
	}
	if len(nameSet) == 0 {
		return nil, nil
	}

	lists, err := api.ListLists(ctx)
	if err != nil {
		return nil, fmt.Errorf("preloading lists: %w", err)
	}
	nameToID := make(map[string]string)
	for _, l := range lists {
		if nameSet[l.Name] {
			nameToID[l.Name] = l.ID
		}
	}

	result := make(map[string]map[string]bool)
	for name, id := range nameToID {
		ids, err := api.GetListBookmarks(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("preloading list %q: %w", name, err)
		}
		set := make(map[string]bool, len(ids))
		for _, bid := range ids {
			set[bid] = true
		}
		result[name] = set
	}
	return result, nil
}

// Run is the core orchestrator. It implements a collect-then-act pattern:
// 1. Paginate all bookmarks via api.ListBookmarks (fail-fast on error).
// 2. For each bookmark, evaluate rules with first-match-wins semantics.
// 3. Execute the matched rule's action (or count as excepted/no-match).
// 4. Return a RunSummary with counters for each outcome category.
//
// Per-bookmark action errors increment Errors and do not abort the run.
// Only a ListBookmarks failure returns a non-nil error.
func Run(ctx context.Context, api KarakeepAPI, rules []config.Rule, dryRun bool, notifications *config.Notifications, notifier Notifier) (RunSummary, error) {
	runTime := time.Now()

	// Phase 1: Collect all bookmarks (fail-fast on error)
	bookmarks, err := api.ListBookmarks(ctx)
	if err != nil {
		return RunSummary{}, fmt.Errorf("collecting bookmarks: %w", err)
	}
	log.Printf("collected %d bookmarks", len(bookmarks))

	// Phase 1.5: Preload list membership data (D-06: after ListBookmarks, before evaluation)
	listSets, err := PreloadListSets(ctx, api, rules)
	if err != nil {
		return RunSummary{}, err
	}
	if listSets != nil {
		log.Printf("preloaded %d list(s) for inList filtering", len(listSets))
	}

	// Initialize per-rule summaries for notification dispatch
	ruleSummaries := make([]*RuleSummary, len(rules))
	for i, rule := range rules {
		ruleSummaries[i] = &RuleSummary{RuleName: rule.Name}
	}

	// Phase 2: Evaluate rules and act
	var summary RunSummary
	for _, b := range bookmarks {
		matched := false
		for ruleIdx, rule := range rules {
			if !MatchesConditions(b, rule.Conditions, runTime, listSets) {
				continue
			}
			if MatchesExceptions(b, rule.Unless, listSets) {
				summary.Excepted++
				ruleSummaries[ruleIdx].Excepted++
				matched = true
				break
			}
			effectiveDryRun := ResolveRuleDryRun(rule.DryRun, dryRun)
			param := ""
			if rule.Tag != nil {
				param = *rule.Tag
			}
			result := ExecuteAction(ctx, api, rule.Action, b, rule.Name, param, effectiveDryRun)
			if result.Err != nil {
				summary.Errors++
				ruleSummaries[ruleIdx].Errors++
			} else {
				summary.record(rule.Action, result.Size)
				ruleSummaries[ruleIdx].record(rule.Action, result.Size)
			}
			matched = true
			break
		}
		if !matched {
			summary.NoMatch++
		}
	}

	// Phase 3: Dispatch per-rule notifications
	dispatchNotifications(notifier, rules, ruleSummaries, notifications, dryRun)

	return summary, nil
}

// dispatchNotifications sends a per-rule notification for every rule that has activity.
// Failures are logged and do not abort the run.
func dispatchNotifications(notifier Notifier, rules []config.Rule, summaries []*RuleSummary, notifications *config.Notifications, dryRun bool) {
	if notifier == nil {
		return
	}
	for i, rule := range rules {
		rs := summaries[i]
		if !rs.HasActivity() {
			continue
		}
		channelURL := ResolveChannelURL(notifications, rule.Notify)
		if channelURL == "" {
			continue
		}
		effectiveDryRun := ResolveRuleDryRun(rule.DryRun, dryRun)
		msg := FormatNotification(rs, effectiveDryRun)
		title := FormatNotificationTitle(rs.RuleName, effectiveDryRun)
		if err := notifier.Send(channelURL, msg, title); err != nil {
			log.Printf("notification failed for rule %q: %v", rule.Name, err)
		}
	}
}

// ResolveRuleDryRun determines effective dry-run for a rule.
// Per-rule setting (non-nil) overrides global; nil inherits global.
func ResolveRuleDryRun(ruleDryRun *bool, globalDryRun bool) bool {
	if ruleDryRun != nil {
		return *ruleDryRun
	}
	return globalDryRun
}
