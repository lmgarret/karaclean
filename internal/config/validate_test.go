package config_test

import (
	"strings"
	"testing"

	"github.com/lm/karaclean/internal/config"
)

func TestValidate_ValidConfig(t *testing.T) {
	cfg := config.Config{
		Rules: []config.Rule{{
			Conditions: &config.Conditions{OlderThan: intPtr(30)},
			Action:     "archive",
		}},
	}
	errs := cfg.Validate()
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidate_NoRules(t *testing.T) {
	cfg := config.Config{}
	errs := cfg.Validate()
	if len(errs) == 0 {
		t.Fatal("expected error for no rules, got none")
	}
	found := false
	for _, e := range errs {
		if e.Field == "rules" && strings.Contains(e.Message, "at least one rule required") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error about 'at least one rule required' with Field='rules', got: %v", errs)
	}
}

func TestValidate_MissingAction(t *testing.T) {
	cfg := config.Config{
		Rules: []config.Rule{{
			Conditions: &config.Conditions{OlderThan: intPtr(30)},
			Action:     "",
		}},
	}
	errs := cfg.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "rules[0]" && strings.Contains(e.Message, `missing required field "action"`) {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error about missing action at rules[0], got: %v", errs)
	}
}

func TestValidate_InvalidAction(t *testing.T) {
	cfg := config.Config{
		Rules: []config.Rule{{
			Conditions: &config.Conditions{OlderThan: intPtr(30)},
			Action:     "remove",
		}},
	}
	errs := cfg.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "rules[0].action" && strings.Contains(e.Message, `invalid value "remove"`) && strings.Contains(e.Message, "must be archive or delete") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error about invalid action 'remove' at rules[0].action, got: %v", errs)
	}
}

func TestValidate_MissingConditions(t *testing.T) {
	cfg := config.Config{
		Rules: []config.Rule{{
			Conditions: nil,
			Action:     "archive",
		}},
	}
	errs := cfg.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "rules[0]" && strings.Contains(e.Message, `missing required field "conditions"`) {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error about missing conditions at rules[0], got: %v", errs)
	}
}

func TestValidate_EmptyConditions(t *testing.T) {
	cfg := config.Config{
		Rules: []config.Rule{{
			Conditions: &config.Conditions{},
			Action:     "archive",
		}},
	}
	errs := cfg.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "rules[0].conditions" && strings.Contains(e.Message, "at least one condition required") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error about empty conditions at rules[0].conditions, got: %v", errs)
	}
}

func TestValidate_InvalidSource(t *testing.T) {
	cfg := config.Config{
		Rules: []config.Rule{{
			Conditions: &config.Conditions{Source: strPtr("feed")},
			Action:     "archive",
		}},
	}
	errs := cfg.Validate()
	found := false
	for _, e := range errs {
		if e.Field == "rules[0].conditions.source" && strings.Contains(e.Message, `invalid value "feed"`) {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error about invalid source 'feed', got: %v", errs)
	}
}

func TestValidate_NegativeOlderThan(t *testing.T) {
	cfg := config.Config{
		Rules: []config.Rule{{
			Conditions: &config.Conditions{OlderThan: intPtr(-1)},
			Action:     "archive",
		}},
	}
	errs := cfg.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Field, "olderThan") && strings.Contains(e.Message, "must be a positive") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected error about negative olderThan, got: %v", errs)
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := config.Config{
		Rules: []config.Rule{
			{Action: "remove", Conditions: nil},
			{Action: "archive", Conditions: &config.Conditions{Source: strPtr("feed")}},
		},
	}
	errs := cfg.Validate()
	if len(errs) < 3 {
		t.Errorf("expected at least 3 errors (invalid action, missing conditions, invalid source), got %d: %v", len(errs), errs)
	}
}

func TestValidationErrors_Error(t *testing.T) {
	ve := &config.ValidationErrors{
		Errors: []config.ValidationError{
			{Field: "rules[0].action", Message: `invalid value "remove"`},
			{Field: "rules[1].conditions.source", Message: `invalid value "feed"`},
		},
	}
	out := ve.Error()
	if !strings.HasPrefix(out, "config validation failed:") {
		t.Errorf("expected output to start with 'config validation failed:', got: %s", out)
	}
	if !strings.Contains(out, "  - ") {
		t.Errorf("expected '  - ' prefixed lines, got: %s", out)
	}
}
