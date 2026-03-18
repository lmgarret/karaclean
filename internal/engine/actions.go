package engine

import (
	"context"
	"fmt"
	"log"
)

// ActionResult records the outcome of executing an action on a single bookmark.
type ActionResult struct {
	BookmarkID string
	RuleName   string
	Action     string
	DryRun     bool
	Err        error
}

// ExecuteAction performs the given action on a bookmark. In dry-run mode, it logs
// the intended action without calling the API. Returns an ActionResult.
//
// action must be "archive" or "delete" (validated at config load time).
// The error format includes bookmark ID and rule name for log-and-continue callers.
func ExecuteAction(ctx context.Context, api KarakeepAPI, action string, bookmarkID string, ruleName string, dryRun bool) ActionResult {
	result := ActionResult{
		BookmarkID: bookmarkID,
		RuleName:   ruleName,
		Action:     action,
		DryRun:     dryRun,
	}

	if dryRun {
		log.Printf("DRY-RUN %s: bookmark %s (rule: %s)", action, bookmarkID, ruleName)
		return result
	}

	var err error
	switch action {
	case "archive":
		err = api.ArchiveBookmark(ctx, bookmarkID)
	case "delete":
		err = api.DeleteBookmark(ctx, bookmarkID)
	default:
		err = fmt.Errorf("unknown action %q", action)
	}

	if err != nil {
		result.Err = fmt.Errorf("%s failed: bookmark %s (rule: %s): %w", action, bookmarkID, ruleName, err)
		log.Printf("ERROR %s", result.Err)
	}

	return result
}
