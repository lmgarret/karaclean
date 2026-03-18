package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v3"
)

// Config represents the top-level karaclean configuration.
type Config struct {
	Timezone string `yaml:"timezone"`
	Schedule string `yaml:"schedule"`
	Rules    []Rule `yaml:"rules"`
}

// Rule defines a single cleanup rule with conditions, exceptions, and an action.
type Rule struct {
	Name       string      `yaml:"name"`
	Conditions *Conditions `yaml:"conditions"`
	Unless     *Exceptions `yaml:"unless"`
	Action     string      `yaml:"action"`
}

// Conditions specifies the matching criteria for bookmarks.
// All non-nil fields must match (AND semantics).
// Pointer types distinguish absent fields (nil) from zero-value fields.
type Conditions struct {
	OlderThan  *int    `yaml:"olderThan"`
	Source     *string `yaml:"source"`
	Archived   *bool   `yaml:"archived"`
	Favourited *bool   `yaml:"favourited"`
	HasTag     *string `yaml:"hasTag"`
	LacksTag   *string `yaml:"lacksTag"`
}

// Exceptions specifies criteria that protect bookmarks from a rule's action.
// Any non-nil field that matches protects the bookmark (OR semantics).
type Exceptions struct {
	Favourited *bool   `yaml:"favourited"`
	HasTag     *string `yaml:"hasTag"`
	HasNote    *bool   `yaml:"hasNote"`
	Archived   *bool   `yaml:"archived"`
}

// Load reads and parses a YAML config file from the given path.
// It uses KnownFields(true) to reject unknown YAML fields (CONF-02).
// On decode error, the partially populated struct is NOT returned.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening config: %w", err)
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// ResolvePath determines the config file path using the precedence:
// 1. Explicit flag value (if non-empty)
// 2. KARACLEAN_CONFIG environment variable (if set)
// 3. Default path: /config/karaclean.yaml
func ResolvePath(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if envPath := os.Getenv("KARACLEAN_CONFIG"); envPath != "" {
		return envPath
	}
	return "/config/karaclean.yaml"
}
