package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lm/karaclean/internal/config"
)

func validConfig() config.Config {
	return config.Config{
		Rules: []config.Rule{{
			Conditions: &config.Conditions{OlderThan: strPtr("30d")},
			Action:     "archive",
		}},
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		wantErr []string // expected substrings in error outputs; empty means no errors
	}{
		{
			name: "valid minimal",
			cfg:  validConfig(),
		},
		{
			name: "valid with all source values",
			cfg: config.Config{
				Rules: []config.Rule{
					{Action: "archive", Conditions: &config.Conditions{Source: strPtr("rss")}},
					{Action: "archive", Conditions: &config.Conditions{Source: strPtr("web")}},
					{Action: "archive", Conditions: &config.Conditions{Source: strPtr("api")}},
					{Action: "archive", Conditions: &config.Conditions{Source: strPtr("mobile")}},
					{Action: "archive", Conditions: &config.Conditions{Source: strPtr("extension")}},
					{Action: "archive", Conditions: &config.Conditions{Source: strPtr("cli")}},
					{Action: "delete", Conditions: &config.Conditions{Source: strPtr("import")}},
				},
			},
		},
		{
			name:    "empty rules",
			cfg:     config.Config{Rules: []config.Rule{}},
			wantErr: []string{"rules: at least one rule required"},
		},
		{
			name:    "nil rules",
			cfg:     config.Config{},
			wantErr: []string{"rules: at least one rule required"},
		},
		{
			name: "missing action",
			cfg: config.Config{
				Rules: []config.Rule{{
					Conditions: &config.Conditions{OlderThan: strPtr("30d")},
					Action:     "",
				}},
			},
			wantErr: []string{`rules[0]: missing required field "action"`},
		},
		{
			name: "invalid action",
			cfg: config.Config{
				Rules: []config.Rule{{
					Conditions: &config.Conditions{OlderThan: strPtr("30d")},
					Action:     "remove",
				}},
			},
			wantErr: []string{`rules[0].action: invalid value "remove"`},
		},
		{
			name: "valid action archive",
			cfg: config.Config{
				Rules: []config.Rule{{
					Conditions: &config.Conditions{OlderThan: strPtr("30d")},
					Action:     "archive",
				}},
			},
		},
		{
			name: "valid action delete",
			cfg: config.Config{
				Rules: []config.Rule{{
					Conditions: &config.Conditions{OlderThan: strPtr("30d")},
					Action:     "delete",
				}},
			},
		},
		{
			name: "missing conditions",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: nil,
				}},
			},
			wantErr: []string{`rules[0]: missing required field "conditions"`},
		},
		{
			name: "empty conditions",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{},
				}},
			},
			wantErr: []string{"rules[0].conditions: at least one condition required"},
		},
		{
			name: "invalid source",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{Source: strPtr("feed")},
				}},
			},
			wantErr: []string{`rules[0].conditions.source: invalid value "feed"`},
		},
		{
			name: "valid source rss",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{Source: strPtr("rss")},
				}},
			},
		},
		{
			name: "negative olderThan",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{OlderThan: strPtr("-1d")},
				}},
			},
			wantErr: []string{"invalid duration"},
		},
		{
			name: "zero olderThan",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{OlderThan: strPtr("0d")},
				}},
			},
		},
		{
			name: "valid olderThan 1 day",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{OlderThan: strPtr("1d")},
				}},
			},
		},
		{
			name: "valid olderThan weeks",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{OlderThan: strPtr("2w")},
				}},
			},
		},
		{
			name: "valid olderThan months",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{OlderThan: strPtr("1mo")},
				}},
			},
		},
		{
			name: "invalid olderThan format",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{OlderThan: strPtr("thirty")},
				}},
			},
			wantErr: []string{"invalid duration"},
		},
		{
			name: "invalid olderThan unit",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{OlderThan: strPtr("30m")},
				}},
			},
			wantErr: []string{"invalid duration"},
		},
		{
			name: "multiple errors same rule",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "remove",
					Conditions: nil,
				}},
			},
			wantErr: []string{
				`rules[0]: missing required field "conditions"`,
				`rules[0].action: invalid value "remove"`,
			},
		},
		{
			name: "multiple errors across rules",
			cfg: config.Config{
				Rules: []config.Rule{
					{Action: "bad", Conditions: &config.Conditions{OlderThan: strPtr("30d")}},
					{Action: "archive", Conditions: &config.Conditions{Source: strPtr("bad")}},
				},
			},
			wantErr: []string{
				"rules[0].action",
				"rules[1].conditions.source",
			},
		},
		{
			name: "empty hasTag rejected",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{HasTag: strPtr("")},
				}},
			},
			wantErr: []string{"rules[0].conditions.hasTag: must not be empty"},
		},
		{
			name: "empty lacksTag rejected",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{LacksTag: strPtr("")},
				}},
			},
			wantErr: []string{"rules[0].conditions.lacksTag: must not be empty"},
		},
		{
			name: "valid hasTag passes",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{HasTag: strPtr("read-later")},
				}},
			},
		},
		{
			name: "valid lacksTag passes",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{LacksTag: strPtr("keep")},
				}},
			},
		},
		{
			name: "both empty hasTag and lacksTag produce two errors",
			cfg: config.Config{
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{HasTag: strPtr(""), LacksTag: strPtr("")},
				}},
			},
			wantErr: []string{
				"rules[0].conditions.hasTag: must not be empty",
				"rules[0].conditions.lacksTag: must not be empty",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.cfg.Validate()

			if len(tt.wantErr) == 0 {
				if len(errs) != 0 {
					t.Errorf("expected 0 errors, got %d: %v", len(errs), errs)
				}
				return
			}

			// Collect all error strings
			var errStrs []string
			for _, e := range errs {
				errStrs = append(errStrs, e.Error())
			}
			allErrs := strings.Join(errStrs, "\n")

			for _, want := range tt.wantErr {
				if !strings.Contains(allErrs, want) {
					t.Errorf("expected error containing %q, got errors:\n%s", want, allErrs)
				}
			}
		})
	}
}

func TestValidationErrors_Error(t *testing.T) {
	ve := &config.ValidationErrors{
		Errors: []config.ValidationError{
			{Field: "rules[0].action", Message: `invalid value "remove" (must be archive or delete)`},
			{Field: "rules[1].conditions.source", Message: `invalid value "feed" (must be rss, web, api, mobile, extension, cli, import)`},
		},
	}
	out := ve.Error()

	if !strings.HasPrefix(out, "config validation failed:\n") {
		t.Errorf("expected output to start with 'config validation failed:\\n', got:\n%s", out)
	}
	if !strings.Contains(out, "  - rules[0].action:") {
		t.Errorf("expected '  - rules[0].action:' line, got:\n%s", out)
	}
	if !strings.Contains(out, "  - rules[1].conditions.source:") {
		t.Errorf("expected '  - rules[1].conditions.source:' line, got:\n%s", out)
	}
}

func TestLoad_ValidationIntegration(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `rules:
  - conditions:
      olderThan: "30d"
    action: remove
`
	path := filepath.Join(dir, "invalid.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid action, got nil")
	}
	if !strings.Contains(err.Error(), "config validation failed") {
		t.Errorf("expected 'config validation failed' in error, got: %s", err)
	}
	if !strings.Contains(err.Error(), `invalid value "remove"`) {
		t.Errorf("expected 'invalid value \"remove\"' in error, got: %s", err)
	}
}

func TestLoad_ValidConfigStillWorks(t *testing.T) {
	cfg, err := config.Load("testdata/valid_full.yaml")
	if err != nil {
		t.Fatalf("expected valid_full.yaml to pass validation, got error: %v", err)
	}
	if len(cfg.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(cfg.Rules))
	}
}
