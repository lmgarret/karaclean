package config

import (
	"fmt"
	"log"
	"os"

	"go.yaml.in/yaml/v3"
)

// ConfigErrorNotifier sends a notification message to a URL.
// This mirrors engine.Notifier to avoid import cycles (config -> engine -> config).
type ConfigErrorNotifier interface {
	Send(url, message, title string) error
}

// Notifications configures notification channels for per-rule action summaries.
type Notifications struct {
	Channels      map[string]NotificationChannel `yaml:"channels"`
	Default       string                         `yaml:"default"`
	NotifyOnError *bool                          `yaml:"notifyOnError"`
}

// NotificationChannel defines a single notification endpoint via Shoutrrr URL.
type NotificationChannel struct {
	URL string `yaml:"url"`
}

// StringOrSlice is a custom type that accepts either a single string or a list
// of strings in YAML. A scalar "value" unmarshals to []string{"value"}, while
// a sequence ["a", "b"] unmarshals to []string{"a", "b"}.
type StringOrSlice []string

func (s *StringOrSlice) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		*s = []string{value.Value}
		return nil
	case yaml.SequenceNode:
		var items []string
		if err := value.Decode(&items); err != nil {
			return err
		}
		*s = items
		return nil
	default:
		return fmt.Errorf("expected string or list of strings")
	}
}

// Config represents the top-level karaclean configuration.
type Config struct {
	Timezone      string         `yaml:"timezone"`
	Schedule      string         `yaml:"schedule"`
	DryRun        bool           `yaml:"dryRun"`
	Notifications *Notifications `yaml:"notifications"`
	Rules         []Rule         `yaml:"rules"`
}

// Rule defines a single cleanup rule with conditions, exceptions, and an action.
type Rule struct {
	Name       string      `yaml:"name"`
	Conditions *Conditions `yaml:"conditions"`
	Unless     *Exceptions `yaml:"unless"`
	Action     string      `yaml:"action"`
	Tag        *string     `yaml:"tag"`
	DryRun     *bool       `yaml:"dryRun"`
	Notify     *string     `yaml:"notify"`
}

// Conditions specifies the matching criteria for bookmarks.
// All non-nil fields must match (AND semantics).
// Pointer types distinguish absent fields (nil) from zero-value fields.
type Conditions struct {
	OlderThan  *string `yaml:"olderThan"`
	Source     *string `yaml:"source"`
	Archived   *bool   `yaml:"archived"`
	Favourited *bool   `yaml:"favourited"`
	HasTag     *string      `yaml:"hasTag"`
	LacksTag   *string      `yaml:"lacksTag"`
	InList     StringOrSlice `yaml:"inList"`
}

// Exceptions specifies criteria that protect bookmarks from a rule's action.
// Any non-nil field that matches protects the bookmark (OR semantics).
type Exceptions struct {
	Favourited *bool   `yaml:"favourited"`
	HasTag     *string `yaml:"hasTag"`
	HasNote    *bool         `yaml:"hasNote"`
	Archived   *bool         `yaml:"archived"`
	InList     StringOrSlice `yaml:"inList"`
}

// notificationsOnly is a minimal struct for lenient fallback parsing.
// When strict YAML decode fails (syntax errors, unknown fields), this struct
// extracts just the notifications section to enable error notification dispatch.
type notificationsOnly struct {
	Notifications *Notifications `yaml:"notifications"`
}

// Load reads and parses a YAML config file from the given path.
// It uses KnownFields(true) to reject unknown YAML fields (CONF-02).
// On decode error, the partially populated struct is NOT returned.
// An optional ConfigErrorNotifier may be passed to dispatch error notifications
// when notifyOnError is true and config validation or parsing fails.
func Load(path string, notifier ...ConfigErrorNotifier) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config: %w", err)
	}
	defer func() { _ = f.Close() }()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		decodeErr := fmt.Errorf("parsing config: %w", err)
		// Lenient fallback: attempt to extract notifications for error dispatch.
		if len(notifier) > 0 && notifier[0] != nil {
			lenientNotify(path, notifier[0], decodeErr)
		}
		return nil, decodeErr
	}

	if errs := cfg.Validate(); len(errs) > 0 {
		validationErr := &ValidationErrors{Errors: errs}
		if len(notifier) > 0 && notifier[0] != nil {
			SendConfigError(cfg.Notifications, notifier[0], validationErr)
		}
		return nil, validationErr
	}

	return &cfg, nil
}

// lenientNotify attempts a lenient parse of just the notifications section
// from a config file that failed strict YAML decode, and sends an error
// notification if notifyOnError is true.
func lenientNotify(path string, notifier ConfigErrorNotifier, configErr error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	var partial notificationsOnly
	decoder := yaml.NewDecoder(f)
	// No KnownFields(true) — lenient parse ignores unknown fields.
	if err := decoder.Decode(&partial); err != nil {
		return // Cannot parse even leniently; give up silently.
	}
	SendConfigError(partial.Notifications, notifier, configErr)
}

// SendConfigError sends a config validation error notification via the default channel.
// It is best-effort: if sending fails, the error is logged but not returned.
// If notifications is nil, notifyOnError is nil/false, or no default channel is set,
// this function is a no-op.
func SendConfigError(n *Notifications, notifier ConfigErrorNotifier, configErr error) {
	if n == nil {
		return
	}
	if n.NotifyOnError == nil || !*n.NotifyOnError {
		return
	}
	if n.Default == "" {
		return
	}
	ch, ok := n.Channels[n.Default]
	if !ok {
		return
	}

	title := "[ERROR] [karaclean] config validation failed"
	body := configErr.Error()
	if err := notifier.Send(ch.URL, body, title); err != nil {
		log.Printf("WARNING: failed to send config error notification: %v", err)
	}
}

// ResolvePath determines the config file path using the precedence:
// 1. Explicit flag value (if non-empty)
// 2. KARACLEAN_CONFIG environment variable (if set)
// 3. Default path: /config/karaclean.yaml.
func ResolvePath(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if envPath := os.Getenv("KARACLEAN_CONFIG"); envPath != "" {
		return envPath
	}
	return "/config/karaclean.yaml"
}
