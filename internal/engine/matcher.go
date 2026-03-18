package engine

import (
	"strings"
	"time"

	"github.com/lm/karaclean/internal/config"
	"github.com/lm/karaclean/internal/duration"
)

// MatchesConditions returns true if the bookmark satisfies all non-nil conditions (AND semantics).
// Short-circuits on first mismatch. The runTime parameter is captured once per run for consistency.
// Duration strings in conditions are assumed pre-validated by config.Validate().
func MatchesConditions(b Bookmark, c *config.Conditions, runTime time.Time) bool {
	if c == nil {
		return true
	}

	if c.OlderThan != nil {
		// Error ignored: config validation guarantees valid duration format at load time.
		dur, _ := duration.Parse(*c.OlderThan)
		if runTime.Sub(b.CreatedAt) <= dur {
			return false
		}
	}

	if c.Source != nil {
		if b.Source != *c.Source {
			return false
		}
	}

	if c.Archived != nil {
		if b.Archived != *c.Archived {
			return false
		}
	}

	if c.Favourited != nil {
		if b.Favourited != *c.Favourited {
			return false
		}
	}

	if c.HasTag != nil {
		found := false
		for _, tag := range b.Tags {
			if tag == *c.HasTag {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if c.LacksTag != nil {
		for _, tag := range b.Tags {
			if tag == *c.LacksTag {
				return false
			}
		}
	}

	return true
}

// MatchesExceptions returns true if the bookmark is protected by any exception clause (OR semantics).
// Short-circuits on first match. Returns false for nil Exceptions.
func MatchesExceptions(b Bookmark, ex *config.Exceptions) bool {
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

	return false
}
