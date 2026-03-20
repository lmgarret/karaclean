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
	Size       int64
	Err        error
}

// HumanSize formats bytes as a human-readable string using 1024-based units.
func HumanSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	units := []string{"KB", "MB", "GB", "TB"}
	size := float64(bytes)
	for _, unit := range units {
		size /= 1024
		if size < 1024 || unit == "TB" {
			return fmt.Sprintf("%.1f %s", size, unit)
		}
	}
	return fmt.Sprintf("%.1f TB", size)
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
	s := fmt.Sprintf("id=%s source=%s tags=%s", b.ID, source, tags)
	if b.Size > 0 {
		s += fmt.Sprintf(" size=%s", HumanSize(b.Size))
	}
	return s
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
		Size:       bookmark.Size,
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
