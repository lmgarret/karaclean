package engine

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lm/karaclean/internal/config"
)

// RunSummary records the outcome of a single Run() invocation.
// Field values sum to the total bookmark count: Archived + Deleted + NoMatch + Excepted + Errors.
type RunSummary struct {
	Archived int `json:"archived"`
	Deleted  int `json:"deleted"`
	NoMatch  int `json:"no_match"`
	Excepted int `json:"excepted"`
	Errors   int `json:"errors"`
}

// String returns a key=value summary suitable for structured log output.
func (s RunSummary) String() string {
	return fmt.Sprintf("archived=%d deleted=%d no_match=%d excepted=%d errors=%d",
		s.Archived, s.Deleted, s.NoMatch, s.Excepted, s.Errors)
}

// Run is the core orchestrator. It implements a collect-then-act pattern:
// 1. Paginate all bookmarks via api.ListBookmarks (fail-fast on error).
// 2. For each bookmark, evaluate rules with first-match-wins semantics.
// 3. Execute the matched rule's action (or count as excepted/no-match).
// 4. Return a RunSummary with counters for each outcome category.
//
// Per-bookmark action errors increment Errors and do not abort the run.
// Only a ListBookmarks failure returns a non-nil error.
func Run(ctx context.Context, api KarakeepAPI, rules []config.Rule, dryRun bool) (RunSummary, error) {
	runTime := time.Now()

	// Phase 1: Collect all bookmarks (fail-fast on error)
	bookmarks, err := api.ListBookmarks(ctx)
	if err != nil {
		return RunSummary{}, fmt.Errorf("collecting bookmarks: %w", err)
	}
	log.Printf("collected %d bookmarks", len(bookmarks))

	// Phase 2: Evaluate rules and act
	var summary RunSummary
	for _, b := range bookmarks {
		matched := false
		for _, rule := range rules {
			if !MatchesConditions(b, rule.Conditions, runTime) {
				continue
			}
			if MatchesExceptions(b, rule.Unless) {
				summary.Excepted++
				matched = true
				break
			}
			effectiveDryRun := ResolveRuleDryRun(rule.DryRun, dryRun)
		result := ExecuteAction(ctx, api, rule.Action, b, rule.Name, effectiveDryRun)
			if result.Err != nil {
				summary.Errors++
			} else {
				switch rule.Action {
				case "archive":
					summary.Archived++
				case "delete":
					summary.Deleted++
				}
			}
			matched = true
			break
		}
		if !matched {
			summary.NoMatch++
		}
	}
	return summary, nil
}

// ResolveRuleDryRun determines effective dry-run for a rule.
// Per-rule setting (non-nil) overrides global; nil inherits global.
func ResolveRuleDryRun(ruleDryRun *bool, globalDryRun bool) bool {
	if ruleDryRun != nil {
		return *ruleDryRun
	}
	return globalDryRun
}
