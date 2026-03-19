package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/lm/karaclean/internal/duration"
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
		cond.LacksTag == nil {
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

	return errs
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
