package engine

import (
	"context"
	"fmt"
	"log"
	"strings"
)

// ActionResult records the outcome of executing an action on a single bookmark.
type ActionResult struct {
	BookmarkID string
	RuleName   string
	Action     string
	DryRun     bool
	Err        error
}

// bookmarkSummary formats bookmark details for log output.
func bookmarkSummary(b Bookmark) string {
	source := b.Source
	if source == "" {
		source = "(unknown)"
	}
	tags := "[]"
	if len(b.Tags) > 0 {
		tags = "[" + strings.Join(b.Tags, ", ") + "]"
	}
	return fmt.Sprintf("id=%s source=%s tags=%s", b.ID, source, tags)
}

// ExecuteAction performs the given action on a bookmark. In dry-run mode, it logs
// the intended action without calling the API. Returns an ActionResult.
//
// action must be "archive" or "delete" (validated at config load time).
// The error format includes bookmark details and rule name for log-and-continue callers.
func ExecuteAction(ctx context.Context, api KarakeepAPI, action string, bookmark Bookmark, ruleName string, dryRun bool) ActionResult {
	result := ActionResult{
		BookmarkID: bookmark.ID,
		RuleName:   ruleName,
		Action:     action,
		DryRun:     dryRun,
	}

	if dryRun {
		log.Printf("DRY-RUN %s: %s (rule: %s)", action, bookmarkSummary(bookmark), ruleName)
		return result
	}

	var err error
	switch action {
	case "archive":
		err = api.ArchiveBookmark(ctx, bookmark.ID)
	case "delete":
		err = api.DeleteBookmark(ctx, bookmark.ID)
	default:
		err = fmt.Errorf("unknown action %q", action)
	}

	if err != nil {
		result.Err = fmt.Errorf("%s failed: %s (rule: %s): %w", action, bookmarkSummary(bookmark), ruleName, err)
		log.Printf("ERROR %s", result.Err)
	} else {
		log.Printf("%s: %s (rule: %s)", action, bookmarkSummary(bookmark), ruleName)
	}

	return result
}
