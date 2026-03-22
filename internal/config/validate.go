package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/lm/karaclean/internal/duration"
	"github.com/nicholas-fedor/shoutrrr"
	"github.com/robfig/cron/v3"
)

// ValidationError represents a single validation error with a field path
// matching the YAML structure (e.g., "rules[0].action").
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors collects multiple validation errors and implements the error interface.
type ValidationErrors struct {
	Errors []ValidationError
}

func (e *ValidationErrors) Error() string {
	var b strings.Builder
	b.WriteString("config validation failed:\n")
	for _, err := range e.Errors {
		fmt.Fprintf(&b, "  - %s\n", err.Error())
	}
	return b.String()
}

var validActions = []string{"archive", "delete"}
var validSources = []string{"rss", "web", "api", "mobile", "extension", "cli", "import"}

// Validate checks the semantic correctness of a Config.
// It collects all errors and returns them at once (not fail-fast).
// Returns nil if the config is valid.
func (c *Config) Validate() []ValidationError {
	var errs []ValidationError

	errs = append(errs, c.validateSchedule()...)
	errs = append(errs, c.validateTimezone()...)
	errs = append(errs, validateNotifications(c.Notifications, c.Rules)...)

	if len(c.Rules) == 0 {
		errs = append(errs, ValidationError{
			Field:   "rules",
			Message: "at least one rule required",
		})
	}

	for i, rule := range c.Rules {
		prefix := fmt.Sprintf("rules[%d]", i)
		errs = append(errs, validateRule(rule, prefix)...)
	}

	return errs
}

// validateSchedule checks the schedule field (SCHED-01).
func (c *Config) validateSchedule() []ValidationError {
	if c.Schedule == "" {
		return []ValidationError{{Field: "schedule", Message: "schedule is required"}}
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	if _, err := parser.Parse(c.Schedule); err != nil {
		return []ValidationError{{Field: "schedule", Message: fmt.Sprintf("invalid cron expression: %v", err)}}
	}
	return nil
}

// validateTimezone checks the timezone field (SCHED-02) -- empty is valid (defaults to UTC at runtime).
func (c *Config) validateTimezone() []ValidationError {
	if c.Timezone != "" {
		if _, err := time.LoadLocation(c.Timezone); err != nil {
			return []ValidationError{{Field: "timezone", Message: fmt.Sprintf("invalid timezone: %v", err)}}
		}
	}
	return nil
}

// validateRule checks a single rule's action, conditions, and exceptions.
func validateRule(rule Rule, prefix string) []ValidationError {
	var errs []ValidationError

	// Check action.
	if rule.Action == "" {
		errs = append(errs, ValidationError{
			Field:   prefix,
			Message: `missing required field "action"`,
		})
	} else if !contains(validActions, rule.Action) {
		errs = append(errs, ValidationError{
			Field:   prefix + ".action",
			Message: fmt.Sprintf("invalid value %q (must be archive or delete)", rule.Action),
		})
	}

	// Check conditions.
	if rule.Conditions == nil {
		errs = append(errs, ValidationError{
			Field:   prefix,
			Message: `missing required field "conditions"`,
		})
	} else {
		errs = append(errs, validateConditions(rule.Conditions, prefix+".conditions")...)
	}

	// Validate exceptions.
	if rule.Unless != nil {
		if rule.Unless.HasTag != nil && *rule.Unless.HasTag == "" {
			errs = append(errs, ValidationError{
				Field:   prefix + ".unless.hasTag",
				Message: "must not be empty",
			})
		}
		for i, name := range rule.Unless.InList {
			if name == "" {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("%s.unless.inList[%d]", prefix, i),
					Message: "list name must not be empty",
				})
			}
		}
	}

	return errs
}

// validateConditions checks a conditions block for semantic correctness.
func validateConditions(cond *Conditions, prefix string) []ValidationError {
	var errs []ValidationError

	// Check all condition fields are nil (empty conditions).
	if cond.OlderThan == nil &&
		cond.Source == nil &&
		cond.Archived == nil &&
		cond.Favourited == nil &&
		cond.HasTag == nil &&
		cond.LacksTag == nil &&
		cond.InList == nil {
		return []ValidationError{{Field: prefix, Message: "at least one condition required"}}
	}

	// Validate source enum.
	if cond.Source != nil && !contains(validSources, *cond.Source) {
		errs = append(errs, ValidationError{
			Field:   prefix + ".source",
			Message: fmt.Sprintf("invalid value %q (must be %s)", *cond.Source, strings.Join(validSources, ", ")),
		})
	}

	// Validate olderThan is a valid duration string.
	if cond.OlderThan != nil {
		if _, err := duration.Parse(*cond.OlderThan); err != nil {
			errs = append(errs, ValidationError{
				Field:   prefix + ".olderThan",
				Message: err.Error(),
			})
		}
	}

	// Validate hasTag is non-empty.
	if cond.HasTag != nil && *cond.HasTag == "" {
		errs = append(errs, ValidationError{
			Field:   prefix + ".hasTag",
			Message: "must not be empty",
		})
	}

	// Validate lacksTag is non-empty.
	if cond.LacksTag != nil && *cond.LacksTag == "" {
		errs = append(errs, ValidationError{
			Field:   prefix + ".lacksTag",
			Message: "must not be empty",
		})
	}

	// Validate inList entries are non-empty.
	for i, name := range cond.InList {
		if name == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("%s.inList[%d]", prefix, i),
				Message: "list name must not be empty",
			})
		}
	}

	return errs
}

// validateNotifications checks notification channel definitions and references.
// Returns nil if notifications are not configured (opt-in).
func validateNotifications(n *Notifications, rules []Rule) []ValidationError {
	var errs []ValidationError
	if n == nil {
		// Check if any rule has notify set without a notifications block
		for i, rule := range rules {
			if rule.Notify != nil {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("rules[%d].notify", i),
					Message: fmt.Sprintf("references channel %q but no notifications block configured", *rule.Notify),
				})
			}
		}
		return errs
	}

	// Validate each channel URL via Shoutrrr
	for name, ch := range n.Channels {
		if ch.URL == "" {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("notifications.channels.%s.url", name),
				Message: "url is required",
			})
			continue
		}
		if _, err := shoutrrr.CreateSender(ch.URL); err != nil {
			errs = append(errs, ValidationError{
				Field:   fmt.Sprintf("notifications.channels.%s.url", name),
				Message: fmt.Sprintf("invalid shoutrrr URL: %v", err),
			})
		}
	}

	// Validate default references a defined channel
	if n.Default != "" {
		if _, ok := n.Channels[n.Default]; !ok {
			errs = append(errs, ValidationError{
				Field:   "notifications.default",
				Message: fmt.Sprintf("references undefined channel %q", n.Default),
			})
		}
	}

	// Validate per-rule notify references
	for i, rule := range rules {
		if rule.Notify != nil {
			if _, ok := n.Channels[*rule.Notify]; !ok {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("rules[%d].notify", i),
					Message: fmt.Sprintf("references undefined channel %q", *rule.Notify),
				})
			}
		}
	}

	return errs
}

// CollectListNames returns a deduplicated slice of all list names referenced
// in conditions.inList and unless.inList across all rules.
func (c *Config) CollectListNames() []string {
	seen := make(map[string]bool)
	for _, r := range c.Rules {
		if r.Conditions != nil {
			for _, name := range r.Conditions.InList {
				seen[name] = true
			}
		}
		if r.Unless != nil {
			for _, name := range r.Unless.InList {
				seen[name] = true
			}
		}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	return names
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
