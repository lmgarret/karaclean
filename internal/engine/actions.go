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

// actionFunc applies a single action to a bookmark via the API. The param
// carries the action's argument (e.g. the tag name for tag/untag); it is
// ignored by actions that take no argument.
type actionFunc func(ctx context.Context, api KarakeepAPI, bookmark Bookmark, param string) error

// actionFuncs maps each supported action name to its API call. The set of keys
// must stay in sync with config.Actions -- TestActionRegistryMatchesConfig
// enforces this. Adding a new action means adding an entry here and in
// config.Actions; no other code in the dispatch path changes.
var actionFuncs = map[string]actionFunc{
	"archive": func(ctx context.Context, api KarakeepAPI, b Bookmark, _ string) error {
		return api.ArchiveBookmark(ctx, b.ID)
	},
	"unarchive": func(ctx context.Context, api KarakeepAPI, b Bookmark, _ string) error {
		return api.UnarchiveBookmark(ctx, b.ID)
	},
	"delete": func(ctx context.Context, api KarakeepAPI, b Bookmark, _ string) error {
		return api.DeleteBookmark(ctx, b.ID)
	},
	"favourite": func(ctx context.Context, api KarakeepAPI, b Bookmark, _ string) error {
		return api.FavouriteBookmark(ctx, b.ID)
	},
	"unfavourite": func(ctx context.Context, api KarakeepAPI, b Bookmark, _ string) error {
		return api.UnfavouriteBookmark(ctx, b.ID)
	},
	"tag": func(ctx context.Context, api KarakeepAPI, b Bookmark, param string) error {
		return api.AddTagToBookmark(ctx, b.ID, param)
	},
	"untag": func(ctx context.Context, api KarakeepAPI, b Bookmark, param string) error {
		return api.RemoveTagFromBookmark(ctx, b.ID, param)
	},
}

// actionLabel renders an action plus its optional param for log output,
// e.g. "tag" with param "review" becomes `tag "review"`.
func actionLabel(action, param string) string {
	if param != "" {
		return fmt.Sprintf("%s %q", action, param)
	}
	return action
}

// ExecuteAction performs the given action on a bookmark. In dry-run mode, it logs
// the intended action without calling the API. Returns an ActionResult.
//
// action must be a key of actionFuncs (validated at config load time). param
// carries the action's argument (the tag name for tag/untag, empty otherwise).
// The error format includes bookmark details and rule name for log-and-continue callers.
func ExecuteAction(ctx context.Context, api KarakeepAPI, action string, bookmark Bookmark, ruleName, param string, dryRun bool) ActionResult {
	result := ActionResult{
		BookmarkID: bookmark.ID,
		RuleName:   ruleName,
		Action:     action,
		DryRun:     dryRun,
		Size:       bookmark.Size,
	}

	label := actionLabel(action, param)

	if dryRun {
		log.Printf("DRY-RUN %s: %s (rule: %s)", label, bookmarkSummary(bookmark), ruleName)
		return result
	}

	fn, ok := actionFuncs[action]
	if !ok {
		result.Err = fmt.Errorf("unknown action %q: %s (rule: %s)", action, bookmarkSummary(bookmark), ruleName)
		log.Printf("ERROR %s", result.Err)
		return result
	}

	if err := fn(ctx, api, bookmark, param); err != nil {
		result.Err = fmt.Errorf("%s failed: %s (rule: %s): %w", label, bookmarkSummary(bookmark), ruleName, err)
		log.Printf("ERROR %s", result.Err)
	} else {
		log.Printf("%s: %s (rule: %s)", label, bookmarkSummary(bookmark), ruleName)
	}

	return result
}
