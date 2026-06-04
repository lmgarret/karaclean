package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lmgarret/karaclean/internal/config"
)

func validConfig() config.Config {
	return config.Config{
		Schedule: "0 3 * * *",
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
				Schedule: "0 3 * * *",
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
			cfg:     config.Config{Schedule: "0 3 * * *", Rules: []config.Rule{}},
			wantErr: []string{"rules: at least one rule required"},
		},
		{
			name:    "nil rules",
			cfg:     config.Config{Schedule: "0 3 * * *"},
			wantErr: []string{"rules: at least one rule required"},
		},
		{
			name: "missing action",
			cfg: config.Config{
				Schedule: "0 3 * * *",
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
				Schedule: "0 3 * * *",
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
				Schedule: "0 3 * * *",
				Rules: []config.Rule{{
					Conditions: &config.Conditions{OlderThan: strPtr("30d")},
					Action:     "archive",
				}},
			},
		},
		{
			name: "valid action delete",
			cfg: config.Config{
				Schedule: "0 3 * * *",
				Rules: []config.Rule{{
					Conditions: &config.Conditions{OlderThan: strPtr("30d")},
					Action:     "delete",
				}},
			},
		},
		{
			name: "missing conditions",
			cfg: config.Config{
				Schedule: "0 3 * * *",
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
				Schedule: "0 3 * * *",
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
				Schedule: "0 3 * * *",
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
				Schedule: "0 3 * * *",
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{Source: strPtr("rss")},
				}},
			},
		},
		{
			name: "negative olderThan",
			cfg: config.Config{
				Schedule: "0 3 * * *",
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
				Schedule: "0 3 * * *",
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{OlderThan: strPtr("0d")},
				}},
			},
		},
		{
			name: "valid olderThan 1 day",
			cfg: config.Config{
				Schedule: "0 3 * * *",
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{OlderThan: strPtr("1d")},
				}},
			},
		},
		{
			name: "valid olderThan weeks",
			cfg: config.Config{
				Schedule: "0 3 * * *",
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{OlderThan: strPtr("2w")},
				}},
			},
		},
		{
			name: "valid olderThan months",
			cfg: config.Config{
				Schedule: "0 3 * * *",
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{OlderThan: strPtr("1mo")},
				}},
			},
		},
		{
			name: "invalid olderThan format",
			cfg: config.Config{
				Schedule: "0 3 * * *",
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
				Schedule: "0 3 * * *",
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
				Schedule: "0 3 * * *",
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
				Schedule: "0 3 * * *",
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
				Schedule: "0 3 * * *",
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
				Schedule: "0 3 * * *",
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
				Schedule: "0 3 * * *",
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{HasTag: strPtr("read-later")},
				}},
			},
		},
		{
			name: "valid lacksTag passes",
			cfg: config.Config{
				Schedule: "0 3 * * *",
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{LacksTag: strPtr("keep")},
				}},
			},
		},
		{
			name: "both empty hasTag and lacksTag produce two errors",
			cfg: config.Config{
				Schedule: "0 3 * * *",
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
		{
			name: "empty unless hasTag rejected",
			cfg: config.Config{
				Schedule: "0 3 * * *",
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{OlderThan: strPtr("30d")},
					Unless:     &config.Exceptions{HasTag: strPtr("")},
				}},
			},
			wantErr: []string{"rules[0].unless.hasTag: must not be empty"},
		},
		{
			name: "valid unless hasTag passes",
			cfg: config.Config{
				Schedule: "0 3 * * *",
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{OlderThan: strPtr("30d")},
					Unless:     &config.Exceptions{HasTag: strPtr("important")},
				}},
			},
		},
		{
			name: "nil unless passes",
			cfg: config.Config{
				Schedule: "0 3 * * *",
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{OlderThan: strPtr("30d")},
					Unless:     nil,
				}},
			},
		},
		{
			name: "unless with nil hasTag passes",
			cfg: config.Config{
				Schedule: "0 3 * * *",
				Rules: []config.Rule{{
					Action:     "archive",
					Conditions: &config.Conditions{OlderThan: strPtr("30d")},
					Unless:     &config.Exceptions{Favourited: boolPtr(true)},
				}},
			},
		},
		// Schedule validation tests
		{
			name:    "missing schedule",
			cfg:     config.Config{Rules: []config.Rule{{Action: "archive", Conditions: &config.Conditions{OlderThan: strPtr("30d")}}}},
			wantErr: []string{"schedule: schedule is required"},
		},
		{
			name:    "invalid cron expression",
			cfg:     config.Config{Schedule: "not-a-cron", Rules: []config.Rule{{Action: "archive", Conditions: &config.Conditions{OlderThan: strPtr("30d")}}}},
			wantErr: []string{"schedule: invalid cron expression"},
		},
		{
			name:    "six-field cron rejected",
			cfg:     config.Config{Schedule: "0 0 3 * * *", Rules: []config.Rule{{Action: "archive", Conditions: &config.Conditions{OlderThan: strPtr("30d")}}}},
			wantErr: []string{"schedule: invalid cron expression"},
		},
		{
			name: "valid cron daily",
			cfg:  config.Config{Schedule: "0 3 * * *", Rules: []config.Rule{{Action: "archive", Conditions: &config.Conditions{OlderThan: strPtr("30d")}}}},
		},
		{
			name: "valid cron every 15 min",
			cfg:  config.Config{Schedule: "*/15 * * * *", Rules: []config.Rule{{Action: "archive", Conditions: &config.Conditions{OlderThan: strPtr("30d")}}}},
		},
		// Timezone validation tests
		{
			name: "empty timezone defaults to UTC no error",
			cfg:  config.Config{Schedule: "0 3 * * *", Timezone: "", Rules: []config.Rule{{Action: "archive", Conditions: &config.Conditions{OlderThan: strPtr("30d")}}}},
		},
		{
			name: "valid timezone America/New_York",
			cfg:  config.Config{Schedule: "0 3 * * *", Timezone: "America/New_York", Rules: []config.Rule{{Action: "archive", Conditions: &config.Conditions{OlderThan: strPtr("30d")}}}},
		},
		{
			name: "valid timezone UTC",
			cfg:  config.Config{Schedule: "0 3 * * *", Timezone: "UTC", Rules: []config.Rule{{Action: "archive", Conditions: &config.Conditions{OlderThan: strPtr("30d")}}}},
		},
		{
			name:    "invalid timezone",
			cfg:     config.Config{Schedule: "0 3 * * *", Timezone: "Mars/Olympus", Rules: []config.Rule{{Action: "archive", Conditions: &config.Conditions{OlderThan: strPtr("30d")}}}},
			wantErr: []string{"timezone: invalid timezone"},
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

func validConfigWithNotifications() config.Config {
	return config.Config{
		Schedule: "0 3 * * *",
		Notifications: &config.Notifications{
			Channels: map[string]config.NotificationChannel{
				"test-ch": {URL: "ntfy://ntfy.sh/test"},
			},
			Default: "test-ch",
		},
		Rules: []config.Rule{{
			Conditions: &config.Conditions{OlderThan: strPtr("30d")},
			Action:     "archive",
		}},
	}
}

func TestValidateNotifications(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		wantErr []string
	}{
		{
			name: "nil notifications no errors",
			cfg:  validConfig(),
		},
		{
			name: "channel with empty URL",
			cfg: func() config.Config {
				c := validConfigWithNotifications()
				c.Notifications.Channels["bad"] = config.NotificationChannel{URL: ""}
				return c
			}(),
			wantErr: []string{"notifications.channels.bad.url: url is required"},
		},
		{
			name: "channel with invalid shoutrrr URL",
			cfg: func() config.Config {
				c := validConfigWithNotifications()
				c.Notifications.Channels["bad"] = config.NotificationChannel{URL: "badscheme://foo"}
				return c
			}(),
			wantErr: []string{"invalid shoutrrr URL"},
		},
		{
			name: "default references undefined channel",
			cfg: func() config.Config {
				c := validConfigWithNotifications()
				c.Notifications.Default = "nonexistent"
				return c
			}(),
			wantErr: []string{"notifications.default: references undefined channel"},
		},
		{
			name: "rule notify references undefined channel",
			cfg: func() config.Config {
				c := validConfigWithNotifications()
				c.Rules[0].Notify = strPtr("nonexistent")
				return c
			}(),
			wantErr: []string{"rules[0].notify: references undefined channel"},
		},
		{
			name: "rule notify references defined channel no errors",
			cfg: func() config.Config {
				c := validConfigWithNotifications()
				c.Rules[0].Notify = strPtr("test-ch")
				return c
			}(),
		},
		{
			name: "no channels but default set",
			cfg: func() config.Config {
				c := validConfigWithNotifications()
				c.Notifications.Channels = map[string]config.NotificationChannel{}
				c.Notifications.Default = "missing"
				return c
			}(),
			wantErr: []string{"references undefined channel"},
		},
		{
			name: "rule notify set but notifications nil",
			cfg: func() config.Config {
				c := validConfig()
				c.Rules[0].Notify = strPtr("some-channel")
				return c
			}(),
			wantErr: []string{"rules[0].notify: references channel"},
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

func TestValidate_InList_EmptyName(t *testing.T) {
	cfg := config.Config{
		Schedule: "0 3 * * *",
		Rules: []config.Rule{{
			Action:     "archive",
			Conditions: &config.Conditions{OlderThan: strPtr("30d"), InList: config.StringOrSlice{"valid", ""}},
		}},
	}
	errs := cfg.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "conditions.inList[1]") && strings.Contains(e.Error(), "list name must not be empty") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected validation error for empty inList entry, got: %v", errs)
	}
}

func TestValidate_InList_UnlessEmptyName(t *testing.T) {
	cfg := config.Config{
		Schedule: "0 3 * * *",
		Rules: []config.Rule{{
			Action:     "archive",
			Conditions: &config.Conditions{OlderThan: strPtr("30d")},
			Unless:     &config.Exceptions{InList: config.StringOrSlice{""}},
		}},
	}
	errs := cfg.Validate()
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "unless.inList[0]") && strings.Contains(e.Error(), "list name must not be empty") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected validation error for empty unless.inList entry, got: %v", errs)
	}
}

func TestValidate_InListOnly_PassesConditionCheck(t *testing.T) {
	cfg := config.Config{
		Schedule: "0 3 * * *",
		Rules: []config.Rule{{
			Action:     "archive",
			Conditions: &config.Conditions{InList: config.StringOrSlice{"Read Later"}},
		}},
	}
	errs := cfg.Validate()
	for _, e := range errs {
		if strings.Contains(e.Error(), "at least one condition required") {
			t.Errorf("inList alone should satisfy condition requirement, got: %v", errs)
		}
	}
	if len(errs) != 0 {
		t.Errorf("expected 0 errors, got %d: %v", len(errs), errs)
	}
}

func TestCollectListNames(t *testing.T) {
	cfg := config.Config{
		Schedule: "0 3 * * *",
		Rules: []config.Rule{
			{
				Action:     "archive",
				Conditions: &config.Conditions{InList: config.StringOrSlice{"Read Later", "Favorites"}},
			},
			{
				Action:     "delete",
				Conditions: &config.Conditions{OlderThan: strPtr("30d")},
				Unless:     &config.Exceptions{InList: config.StringOrSlice{"Read Later", "Important"}},
			},
		},
	}
	names := cfg.CollectListNames()
	// Should be deduplicated: Read Later, Favorites, Important
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	if len(nameSet) != 3 {
		t.Errorf("expected 3 unique list names, got %d: %v", len(nameSet), names)
	}
	for _, want := range []string{"Read Later", "Favorites", "Important"} {
		if !nameSet[want] {
			t.Errorf("expected list name %q in result, got: %v", want, names)
		}
	}
}

func TestCollectListNames_Empty(t *testing.T) {
	cfg := validConfig()
	names := cfg.CollectListNames()
	if len(names) != 0 {
		t.Errorf("expected empty slice, got %v", names)
	}
}

func TestLoad_ValidationIntegration(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `schedule: "0 3 * * *"
rules:
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
