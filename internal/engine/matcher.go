package engine

import (
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

	return true
}
