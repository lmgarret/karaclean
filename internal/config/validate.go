package config

import (
	"fmt"
	"strings"

	"github.com/lm/karaclean/internal/duration"
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

	if len(c.Rules) == 0 {
		errs = append(errs, ValidationError{
			Field:   "rules",
			Message: "at least one rule required",
		})
	}

	for i, rule := range c.Rules {
		prefix := fmt.Sprintf("rules[%d]", i)

		// Check action
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

		// Check conditions
		if rule.Conditions == nil {
			errs = append(errs, ValidationError{
				Field:   prefix,
				Message: `missing required field "conditions"`,
			})
		} else {
			// Check all condition fields are nil (empty conditions)
			if rule.Conditions.OlderThan == nil &&
				rule.Conditions.Source == nil &&
				rule.Conditions.Archived == nil &&
				rule.Conditions.Favourited == nil &&
				rule.Conditions.HasTag == nil &&
				rule.Conditions.LacksTag == nil {
				errs = append(errs, ValidationError{
					Field:   prefix + ".conditions",
					Message: "at least one condition required",
				})
			}

			// Validate source enum
			if rule.Conditions.Source != nil && !contains(validSources, *rule.Conditions.Source) {
				errs = append(errs, ValidationError{
					Field:   prefix + ".conditions.source",
					Message: fmt.Sprintf("invalid value %q (must be %s)", *rule.Conditions.Source, strings.Join(validSources, ", ")),
				})
			}

			// Validate olderThan is a valid duration string
			if rule.Conditions.OlderThan != nil {
				if _, err := duration.Parse(*rule.Conditions.OlderThan); err != nil {
					errs = append(errs, ValidationError{
						Field:   prefix + ".conditions.olderThan",
						Message: err.Error(),
					})
				}
			}

			// Validate hasTag is non-empty
			if rule.Conditions.HasTag != nil && *rule.Conditions.HasTag == "" {
				errs = append(errs, ValidationError{
					Field:   prefix + ".conditions.hasTag",
					Message: "must not be empty",
				})
			}

			// Validate lacksTag is non-empty
			if rule.Conditions.LacksTag != nil && *rule.Conditions.LacksTag == "" {
				errs = append(errs, ValidationError{
					Field:   prefix + ".conditions.lacksTag",
					Message: "must not be empty",
				})
			}
		}
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
