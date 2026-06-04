package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lmgarret/karaclean/internal/config"
)

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }

func TestLoad_ValidFull(t *testing.T) {
	cfg, err := config.Load("testdata/valid_full.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Timezone != "America/New_York" {
		t.Errorf("timezone = %q, want %q", cfg.Timezone, "America/New_York")
	}
	if cfg.Schedule != "0 3 * * *" {
		t.Errorf("schedule = %q, want %q", cfg.Schedule, "0 3 * * *")
	}
	if len(cfg.Rules) != 2 {
		t.Fatalf("len(rules) = %d, want 2", len(cfg.Rules))
	}

	t.Run("first rule", func(t *testing.T) {
		assertFirstRule(t, cfg.Rules[0])
	})

	t.Run("second rule", func(t *testing.T) {
		assertSecondRule(t, cfg.Rules[1])
	})
}

func assertFirstRule(t *testing.T, r0 config.Rule) {
	t.Helper()
	if r0.Name != "old-rss-cleanup" {
		t.Errorf("rules[0].name = %q, want %q", r0.Name, "old-rss-cleanup")
	}
	if r0.Conditions == nil {
		t.Fatal("rules[0].conditions is nil")
	}
	if r0.Conditions.OlderThan == nil || *r0.Conditions.OlderThan != "30d" {
		t.Errorf("rules[0].conditions.olderThan = %v, want 30d", r0.Conditions.OlderThan)
	}
	if r0.Conditions.Source == nil || *r0.Conditions.Source != "rss" {
		t.Errorf("rules[0].conditions.source = %v, want rss", r0.Conditions.Source)
	}
	if r0.Unless == nil {
		t.Fatal("rules[0].unless is nil")
	}
	if r0.Unless.Favourited == nil || *r0.Unless.Favourited != true {
		t.Errorf("rules[0].unless.favourited = %v, want true", r0.Unless.Favourited)
	}
	if r0.Action != "archive" {
		t.Errorf("rules[0].action = %q, want %q", r0.Action, "archive")
	}
}

func assertSecondRule(t *testing.T, r1 config.Rule) {
	t.Helper()
	if r1.Action != "delete" {
		t.Errorf("rules[1].action = %q, want %q", r1.Action, "delete")
	}
	if r1.Unless == nil {
		t.Fatal("rules[1].unless is nil")
	}
	if r1.Unless.HasTag == nil || *r1.Unless.HasTag != "keep-forever" {
		t.Errorf("rules[1].unless.hasTag = %v, want keep-forever", r1.Unless.HasTag)
	}
	if r1.Unless.HasNote == nil || *r1.Unless.HasNote != true {
		t.Errorf("rules[1].unless.hasNote = %v, want true", r1.Unless.HasNote)
	}
}

func TestLoad_ValidMinimal(t *testing.T) {
	cfg, err := config.Load("testdata/valid_minimal.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Rules) != 1 {
		t.Fatalf("len(rules) = %d, want 1", len(cfg.Rules))
	}

	r := cfg.Rules[0]
	if r.Conditions == nil {
		t.Fatal("rules[0].conditions is nil")
	}
	if r.Conditions.OlderThan == nil || *r.Conditions.OlderThan != "30d" {
		t.Errorf("rules[0].conditions.olderThan = %v, want 30d", r.Conditions.OlderThan)
	}
	if r.Action != "archive" {
		t.Errorf("rules[0].action = %q, want %q", r.Action, "archive")
	}
	if r.Name != "" {
		t.Errorf("rules[0].name = %q, want empty string", r.Name)
	}
	if r.Unless != nil {
		t.Errorf("rules[0].unless = %v, want nil", r.Unless)
	}
}

func TestLoad_PointerSemantics(t *testing.T) {
	cfg, err := config.Load("testdata/valid_minimal.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := cfg.Rules[0]
	if r.Conditions.Source != nil {
		t.Errorf("conditions.source = %v, want nil (absent)", r.Conditions.Source)
	}
	if r.Conditions.Archived != nil {
		t.Errorf("conditions.archived = %v, want nil (absent)", r.Conditions.Archived)
	}
	if r.Conditions.Favourited != nil {
		t.Errorf("conditions.favourited = %v, want nil (absent)", r.Conditions.Favourited)
	}
	if r.Conditions.HasTag != nil {
		t.Errorf("conditions.hasTag = %v, want nil (absent)", r.Conditions.HasTag)
	}
	if r.Conditions.LacksTag != nil {
		t.Errorf("conditions.lacksTag = %v, want nil (absent)", r.Conditions.LacksTag)
	}
}

func TestLoad_UnknownFieldTop(t *testing.T) {
	_, err := config.Load("testdata/unknown_field_top.yaml")
	if err == nil {
		t.Fatal("expected error for unknown top-level field, got nil")
	}
	if !strings.Contains(err.Error(), "unknownField") {
		t.Errorf("error should mention unknownField, got: %s", err)
	}
}

func TestLoad_UnknownFieldNested(t *testing.T) {
	_, err := config.Load("testdata/unknown_field_nested.yaml")
	if err == nil {
		t.Fatal("expected error for unknown nested field, got nil")
	}
	if !strings.Contains(err.Error(), "unknownCondition") {
		t.Errorf("error should mention unknownCondition, got: %s", err)
	}
}

func TestLoad_WrongType(t *testing.T) {
	_, err := config.Load("testdata/wrong_type.yaml")
	if err == nil {
		t.Fatal("expected error for invalid duration, got nil")
	}
	if !strings.Contains(err.Error(), "invalid duration") {
		t.Errorf("error should mention invalid duration, got: %s", err)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
	if !strings.Contains(err.Error(), "opening config") {
		t.Errorf("error should contain 'opening config', got: %s", err)
	}
}

func TestLoad_DryRunTrue(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `dryRun: true
schedule: "0 3 * * *"
rules:
  - name: test
    conditions:
      source: rss
    action: archive
`
	path := filepath.Join(dir, "dryrun_true.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DryRun != true {
		t.Errorf("cfg.DryRun = %v, want true", cfg.DryRun)
	}
}

func TestLoad_DryRunFalse(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `dryRun: false
schedule: "0 3 * * *"
rules:
  - name: test
    conditions:
      source: rss
    action: archive
`
	path := filepath.Join(dir, "dryrun_false.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DryRun != false {
		t.Errorf("cfg.DryRun = %v, want false", cfg.DryRun)
	}
}

func TestLoad_DryRunOmitted(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `schedule: "0 3 * * *"
rules:
  - name: test
    conditions:
      source: rss
    action: archive
`
	path := filepath.Join(dir, "dryrun_omitted.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DryRun != false {
		t.Errorf("cfg.DryRun = %v, want false (default)", cfg.DryRun)
	}
}

func TestLoad_RuleDryRunTrue(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `schedule: "0 3 * * *"
rules:
  - name: test
    conditions:
      source: rss
    action: archive
    dryRun: true
`
	path := filepath.Join(dir, "rule_dryrun_true.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Rules[0].DryRun == nil {
		t.Fatal("expected DryRun to be non-nil")
	}
	if *cfg.Rules[0].DryRun != true {
		t.Errorf("*cfg.Rules[0].DryRun = %v, want true", *cfg.Rules[0].DryRun)
	}
}

func TestLoad_RuleDryRunFalse(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `schedule: "0 3 * * *"
rules:
  - name: test
    conditions:
      source: rss
    action: archive
    dryRun: false
`
	path := filepath.Join(dir, "rule_dryrun_false.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Rules[0].DryRun == nil {
		t.Fatal("expected DryRun to be non-nil")
	}
	if *cfg.Rules[0].DryRun != false {
		t.Errorf("*cfg.Rules[0].DryRun = %v, want false", *cfg.Rules[0].DryRun)
	}
}

func TestLoad_RuleDryRunOmitted(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `schedule: "0 3 * * *"
rules:
  - name: test
    conditions:
      source: rss
    action: archive
`
	path := filepath.Join(dir, "rule_dryrun_omitted.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Rules[0].DryRun != nil {
		t.Errorf("expected DryRun to be nil (omitted), got %v", *cfg.Rules[0].DryRun)
	}
}

func TestLoad_ValidNotifications(t *testing.T) {
	cfg, err := config.Load("testdata/valid_notifications.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Notifications == nil {
		t.Fatal("expected Notifications to be non-nil")
	}
	if len(cfg.Notifications.Channels) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(cfg.Notifications.Channels))
	}
	if ch, ok := cfg.Notifications.Channels["my-ntfy"]; !ok {
		t.Error("expected channel 'my-ntfy' to exist")
	} else if ch.URL != "ntfy://ntfy.sh/karaclean-alerts" {
		t.Errorf("my-ntfy URL = %q, want %q", ch.URL, "ntfy://ntfy.sh/karaclean-alerts")
	}
	if ch, ok := cfg.Notifications.Channels["slack-team"]; !ok {
		t.Error("expected channel 'slack-team' to exist")
	} else if ch.URL != "ntfy://ntfy.sh/karaclean-slack-team" {
		t.Errorf("slack-team URL = %q, want %q", ch.URL, "ntfy://ntfy.sh/karaclean-slack-team")
	}
	if cfg.Notifications.Default != "my-ntfy" {
		t.Errorf("default = %q, want %q", cfg.Notifications.Default, "my-ntfy")
	}

	// Rules[0] has notify: slack-team
	if cfg.Rules[0].Notify == nil {
		t.Fatal("expected Rules[0].Notify to be non-nil")
	}
	if *cfg.Rules[0].Notify != "slack-team" {
		t.Errorf("Rules[0].Notify = %q, want %q", *cfg.Rules[0].Notify, "slack-team")
	}

	// Rules[1] has no notify field
	if cfg.Rules[1].Notify != nil {
		t.Errorf("expected Rules[1].Notify to be nil, got %q", *cfg.Rules[1].Notify)
	}
}

func TestLoad_NoNotifications(t *testing.T) {
	cfg, err := config.Load("testdata/valid_full.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Notifications != nil {
		t.Errorf("expected Notifications to be nil for valid_full.yaml, got %+v", cfg.Notifications)
	}
}

func TestLoad_InListString(t *testing.T) {
	cfg, err := config.Load("testdata/valid_inlist_string.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Rules) != 1 {
		t.Fatalf("len(rules) = %d, want 1", len(cfg.Rules))
	}

	r := cfg.Rules[0]
	if r.Conditions == nil {
		t.Fatal("rules[0].conditions is nil")
	}
	if r.Conditions.InList == nil {
		t.Fatal("rules[0].conditions.inList is nil")
	}
	if len(r.Conditions.InList) != 1 || r.Conditions.InList[0] != "Read Later" {
		t.Errorf("rules[0].conditions.inList = %v, want [Read Later]", r.Conditions.InList)
	}
}

func TestLoad_InListList(t *testing.T) {
	cfg, err := config.Load("testdata/valid_inlist_list.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Rules) != 1 {
		t.Fatalf("len(rules) = %d, want 1", len(cfg.Rules))
	}

	r := cfg.Rules[0]
	if r.Unless == nil {
		t.Fatal("rules[0].unless is nil")
	}
	if r.Unless.InList == nil {
		t.Fatal("rules[0].unless.inList is nil")
	}
	if len(r.Unless.InList) != 2 {
		t.Fatalf("len(unless.inList) = %d, want 2", len(r.Unless.InList))
	}
	if r.Unless.InList[0] != "Read Later" || r.Unless.InList[1] != "Favorites" {
		t.Errorf("rules[0].unless.inList = %v, want [Read Later, Favorites]", r.Unless.InList)
	}
}

// mockNotifier records Send calls for testing.
type mockNotifier struct {
	calls []sendCall
	err   error // error to return from Send, if any
}

type sendCall struct {
	url, message, title string
}

func (m *mockNotifier) Send(url, message, title string) error {
	m.calls = append(m.calls, sendCall{url: url, message: message, title: title})
	return m.err
}

func TestSendConfigError(t *testing.T) {
	t.Run("nil notifications is no-op", func(t *testing.T) {
		mock := &mockNotifier{}
		config.SendConfigError(nil, mock, fmt.Errorf("some error"))
		if len(mock.calls) != 0 {
			t.Errorf("expected no calls, got %d", len(mock.calls))
		}
	})

	t.Run("notifyOnError false is no-op", func(t *testing.T) {
		mock := &mockNotifier{}
		n := &config.Notifications{
			Channels:      map[string]config.NotificationChannel{"ch": {URL: "ntfy://ntfy.sh/test"}},
			Default:       "ch",
			NotifyOnError: boolPtr(false),
		}
		config.SendConfigError(n, mock, fmt.Errorf("some error"))
		if len(mock.calls) != 0 {
			t.Errorf("expected no calls, got %d", len(mock.calls))
		}
	})

	t.Run("notifyOnError nil is no-op", func(t *testing.T) {
		mock := &mockNotifier{}
		n := &config.Notifications{
			Channels: map[string]config.NotificationChannel{"ch": {URL: "ntfy://ntfy.sh/test"}},
			Default:  "ch",
		}
		config.SendConfigError(n, mock, fmt.Errorf("some error"))
		if len(mock.calls) != 0 {
			t.Errorf("expected no calls, got %d", len(mock.calls))
		}
	})

	t.Run("no default channel is no-op", func(t *testing.T) {
		mock := &mockNotifier{}
		n := &config.Notifications{
			Channels:      map[string]config.NotificationChannel{"ch": {URL: "ntfy://ntfy.sh/test"}},
			NotifyOnError: boolPtr(true),
		}
		config.SendConfigError(n, mock, fmt.Errorf("some error"))
		if len(mock.calls) != 0 {
			t.Errorf("expected no calls, got %d", len(mock.calls))
		}
	})

	t.Run("successful send", func(t *testing.T) {
		mock := &mockNotifier{}
		n := &config.Notifications{
			Channels:      map[string]config.NotificationChannel{"my-ntfy": {URL: "ntfy://ntfy.sh/karaclean-alerts"}},
			Default:       "my-ntfy",
			NotifyOnError: boolPtr(true),
		}
		configErr := fmt.Errorf("config validation failed:\n  - rules[0]: missing required field \"action\"")
		config.SendConfigError(n, mock, configErr)
		if len(mock.calls) != 1 {
			t.Fatalf("expected 1 call, got %d", len(mock.calls))
		}
		call := mock.calls[0]
		if call.url != "ntfy://ntfy.sh/karaclean-alerts" {
			t.Errorf("url = %q, want %q", call.url, "ntfy://ntfy.sh/karaclean-alerts")
		}
		if call.title != "[ERROR] [karaclean] config validation failed" {
			t.Errorf("title = %q, want %q", call.title, "[ERROR] [karaclean] config validation failed")
		}
		if call.message != configErr.Error() {
			t.Errorf("message = %q, want %q", call.message, configErr.Error())
		}
	})

	t.Run("send failure is logged not propagated", func(t *testing.T) {
		mock := &mockNotifier{err: fmt.Errorf("network error")}
		n := &config.Notifications{
			Channels:      map[string]config.NotificationChannel{"ch": {URL: "ntfy://ntfy.sh/test"}},
			Default:       "ch",
			NotifyOnError: boolPtr(true),
		}
		// Should not panic or return error — best effort.
		config.SendConfigError(n, mock, fmt.Errorf("validation error"))
		if len(mock.calls) != 1 {
			t.Errorf("expected 1 call even on error, got %d", len(mock.calls))
		}
	})

	t.Run("default channel not found is no-op", func(t *testing.T) {
		mock := &mockNotifier{}
		n := &config.Notifications{
			Channels:      map[string]config.NotificationChannel{"other": {URL: "ntfy://ntfy.sh/test"}},
			Default:       "nonexistent",
			NotifyOnError: boolPtr(true),
		}
		config.SendConfigError(n, mock, fmt.Errorf("validation error"))
		if len(mock.calls) != 0 {
			t.Errorf("expected no calls for missing channel, got %d", len(mock.calls))
		}
	})
}

func TestLoad_NotifyOnError(t *testing.T) {
	t.Run("valid config with notifyOnError true", func(t *testing.T) {
		cfg, err := config.Load("testdata/valid_notify_on_error.yaml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Notifications == nil {
			t.Fatal("expected Notifications to be non-nil")
		}
		if cfg.Notifications.NotifyOnError == nil || !*cfg.Notifications.NotifyOnError {
			t.Error("expected NotifyOnError to be true")
		}
	})

	t.Run("invalid config with notifyOnError triggers notification", func(t *testing.T) {
		mock := &mockNotifier{}
		_, err := config.Load("testdata/invalid_with_notify_on_error.yaml", mock)
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
		if len(mock.calls) != 1 {
			t.Fatalf("expected 1 notification call, got %d", len(mock.calls))
		}
		call := mock.calls[0]
		if call.title != "[ERROR] [karaclean] config validation failed" {
			t.Errorf("title = %q, want %q", call.title, "[ERROR] [karaclean] config validation failed")
		}
		if !strings.Contains(call.message, "action") {
			t.Errorf("message should contain 'action', got %q", call.message)
		}
	})

	t.Run("invalid config without notifyOnError does not notify", func(t *testing.T) {
		mock := &mockNotifier{}
		// Use a temp file with invalid config but no notifyOnError
		dir := t.TempDir()
		yamlContent := `schedule: "0 3 * * *"
notifications:
  channels:
    my-ntfy:
      url: "ntfy://ntfy.sh/karaclean-alerts"
  default: my-ntfy
rules:
  - name: missing-action
    conditions:
      olderThan: "30d"
`
		path := filepath.Join(dir, "no_notify.yaml")
		if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("failed to write temp file: %v", err)
		}
		_, err := config.Load(path, mock)
		if err == nil {
			t.Fatal("expected validation error")
		}
		if len(mock.calls) != 0 {
			t.Errorf("expected no notification calls, got %d", len(mock.calls))
		}
	})
}

func TestLoad_LenientFallback(t *testing.T) {
	t.Run("syntax error with notifications triggers lenient parse and notification", func(t *testing.T) {
		mock := &mockNotifier{}
		_, err := config.Load("testdata/syntax_error_with_notifications.yaml", mock)
		if err == nil {
			t.Fatal("expected parse error, got nil")
		}
		if len(mock.calls) != 1 {
			t.Fatalf("expected 1 notification call from lenient fallback, got %d", len(mock.calls))
		}
		call := mock.calls[0]
		if call.title != "[ERROR] [karaclean] config validation failed" {
			t.Errorf("title = %q, want %q", call.title, "[ERROR] [karaclean] config validation failed")
		}
	})
}

func TestResolvePath_Flag(t *testing.T) {
	got := config.ResolvePath("explicit.yaml")
	if got != "explicit.yaml" {
		t.Errorf("ResolvePath(explicit.yaml) = %q, want %q", got, "explicit.yaml")
	}
}

func TestResolvePath_EnvVar(t *testing.T) {
	t.Setenv("KARACLEAN_CONFIG", "/custom/path.yaml")
	got := config.ResolvePath("")
	if got != "/custom/path.yaml" {
		t.Errorf("ResolvePath('') with env = %q, want %q", got, "/custom/path.yaml")
	}
}

func TestResolvePath_Default(t *testing.T) {
	t.Setenv("KARACLEAN_CONFIG", "")
	got := config.ResolvePath("")
	if got != "/config/karaclean.yaml" {
		t.Errorf("ResolvePath('') default = %q, want %q", got, "/config/karaclean.yaml")
	}
}
