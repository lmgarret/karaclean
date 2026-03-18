package duration

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var re = regexp.MustCompile(`^(\d+)(h|d|w|mo|y)$`)

// Parse converts a compact duration string (e.g., "30d", "2w", "6h", "1mo", "1y")
// into a time.Duration. Supported units: h (hours), d (days), w (weeks),
// mo (months, ~30 days), y (years, ~365 days).
func Parse(s string) (time.Duration, error) {
	m := re.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("invalid duration %q (use format like 30d, 2w, 6h, 1mo, 1y)", s)
	}
	n, _ := strconv.Atoi(m[1]) // regex guarantees digits
	switch m[2] {
	case "h":
		return time.Duration(n) * time.Hour, nil
	case "d":
		return time.Duration(n) * 24 * time.Hour, nil
	case "w":
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	case "mo":
		return time.Duration(n) * 30 * 24 * time.Hour, nil
	case "y":
		return time.Duration(n) * 365 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown unit %q", m[2])
	}
}
