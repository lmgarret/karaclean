package engine

import (
	"strings"
	"time"

	"github.com/lmgarret/karaclean/internal/config"
	"github.com/lmgarret/karaclean/internal/duration"
)

// MatchesConditions returns true if the bookmark satisfies all non-nil conditions (AND semantics).
// Short-circuits on first mismatch. The runTime parameter is captured once per run for consistency.
// Duration strings in conditions are assumed pre-validated by config.Validate().
func MatchesConditions(b Bookmark, c *config.Conditions, runTime time.Time, listSets map[string]map[string]bool) bool {
	if c == nil {
		return true
	}

	if c.OlderThan != nil && !matchesOlderThan(b, *c.OlderThan, runTime) {
		return false
	}

	if c.Source != nil && b.Source != *c.Source {
		return false
	}

	if c.Archived != nil && b.Archived != *c.Archived {
		return false
	}

	if c.Favourited != nil && b.Favourited != *c.Favourited {
		return false
	}

	if !matchesTags(b.Tags, c.HasTag, c.LacksTag) {
		return false
	}

	if c.InList != nil && !inAnyList(b.ID, c.InList, listSets) {
		return false
	}

	return true
}

// matchesTags returns true if tag conditions are satisfied.
func matchesTags(tags []string, mustHave, mustLack *string) bool {
	if mustHave != nil && !hasTag(tags, *mustHave) {
		return false
	}
	if mustLack != nil && hasTag(tags, *mustLack) {
		return false
	}
	return true
}

// inAnyList returns true if the bookmark ID belongs to any of the named lists.
func inAnyList(bookmarkID string, listNames []string, listSets map[string]map[string]bool) bool {
	for _, listName := range listNames {
		if set, ok := listSets[listName]; ok && set[bookmarkID] {
			return true
		}
	}
	return false
}

// matchesOlderThan returns true if the bookmark is strictly older than the given duration.
func matchesOlderThan(b Bookmark, olderThan string, runTime time.Time) bool {
	// Error ignored: config validation guarantees valid duration format at load time.
	dur, _ := duration.Parse(olderThan)
	return runTime.Sub(b.CreatedAt) > dur
}

// hasTag returns true if the tag list contains the given tag (case-sensitive).
func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// MatchesExceptions returns true if the bookmark is protected by any exception clause (OR semantics).
// Short-circuits on first match. Returns false for nil Exceptions.
func MatchesExceptions(b Bookmark, ex *config.Exceptions, listSets map[string]map[string]bool) bool {
	if ex == nil {
		return false
	}

	if ex.Favourited != nil {
		if b.Favourited == *ex.Favourited {
			return true
		}
	}

	if ex.HasTag != nil {
		for _, tag := range b.Tags {
			if tag == *ex.HasTag {
				return true
			}
		}
	}

	if ex.HasNote != nil {
		hasNote := strings.TrimSpace(b.Note) != ""
		if *ex.HasNote == hasNote {
			return true
		}
	}

	if ex.Archived != nil {
		if b.Archived == *ex.Archived {
			return true
		}
	}

	if ex.InList != nil {
		for _, listName := range ex.InList {
			if set, ok := listSets[listName]; ok && set[b.ID] {
				return true
			}
		}
	}

	return false
}
