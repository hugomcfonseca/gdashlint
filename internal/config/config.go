package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

// Config is the gdashlint YAML configuration schema.
type Config struct {
	Version     int                     `yaml:"version"`
	FailOn      string                  `yaml:"failOn"`
	Output      Output                  `yaml:"output"`
	Rules       map[string]RuleOverride `yaml:"rules"`
	CustomRules map[string]CustomRule   `yaml:"customRules"`
	Ignore      []Ignore                `yaml:"ignore"`
	Path        string                  `yaml:"-"`
}

// Output configures default output behavior.
type Output struct {
	Format string `yaml:"format"`
	Sort   string `yaml:"sort"`
}

// RuleOverride configures a registered rule.
type RuleOverride struct {
	Enabled  *bool          `yaml:"enabled"`
	Severity string         `yaml:"severity"`
	Options  map[string]any `yaml:"options"`
}

// CustomRule defines a config-defined rule.
type CustomRule struct {
	Type     string `yaml:"type"`
	Severity string `yaml:"severity"`
	Path     string `yaml:"path"`
	Message  string `yaml:"message"`
	Pattern  string `yaml:"pattern"`
	Values   []any  `yaml:"values"`
}

// Ignore suppresses findings for a rule and path set.
type Ignore struct {
	Rule      string   `yaml:"rule"`
	Paths     []string `yaml:"paths"`
	JSONPaths []string `yaml:"jsonPaths"`
	Reason    string   `yaml:"reason"`
}

// Load reads a config file from path.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	cfg := Default()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	cfg.Path = path
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Default returns default configuration.
func Default() Config {
	return Config{
		Version:     1,
		FailOn:      string(rule.SeverityError),
		Output:      Output{Format: "text", Sort: "severity"},
		Rules:       make(map[string]RuleOverride),
		CustomRules: make(map[string]CustomRule),
	}
}

// Validate validates configuration values.
func (c Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported config version %d", c.Version)
	}
	failOn, err := rule.ParseSeverity(c.FailOn)
	if err != nil || !failOn.AllowsFailThreshold() {
		return fmt.Errorf("invalid failOn %q", c.FailOn)
	}
	if c.Output.Format != "text" && c.Output.Format != "json" && c.Output.Format != "github" {
		return fmt.Errorf("invalid output.format %q", c.Output.Format)
	}
	if c.Output.Sort != "severity" && c.Output.Sort != "file" {
		return fmt.Errorf("invalid output.sort %q", c.Output.Sort)
	}
	for id, override := range c.Rules {
		if id == "" {
			return fmt.Errorf("rules contains an empty rule ID")
		}
		if override.Severity != "" {
			if _, err := rule.ParseSeverity(override.Severity); err != nil {
				return fmt.Errorf("rule %s: %w", id, err)
			}
		}
	}
	for id, custom := range c.CustomRules {
		if id == "" {
			return fmt.Errorf("customRules contains an empty rule ID")
		}
		if custom.Type == "" {
			return fmt.Errorf("custom rule %s: type is required", id)
		}
		if custom.Path == "" {
			return fmt.Errorf("custom rule %s: path is required", id)
		}
		if err := dashboard.ValidatePath(custom.Path); err != nil {
			return fmt.Errorf("custom rule %s: invalid path: %w", id, err)
		}
		if custom.Severity != "" {
			if _, err := rule.ParseSeverity(custom.Severity); err != nil {
				return fmt.Errorf("custom rule %s: %w", id, err)
			}
		}
	}
	for _, ignore := range c.Ignore {
		if ignore.Rule == "" {
			return fmt.Errorf("ignore rule is required")
		}
		if len(ignore.Paths) == 0 && len(ignore.JSONPaths) == 0 {
			return fmt.Errorf("ignore for rule %s must include paths or jsonPaths", ignore.Rule)
		}
		for _, path := range ignore.JSONPaths {
			if err := dashboard.ValidatePath(path); err != nil {
				return fmt.Errorf("ignore for rule %s: invalid jsonPath %q: %w", ignore.Rule, path, err)
			}
		}
	}
	return nil
}

// FailOnSeverity returns parsed failOn severity.
func (c Config) FailOnSeverity() rule.Severity {
	severity, err := rule.ParseSeverity(c.FailOn)
	if err != nil {
		return rule.SeverityError
	}
	return severity
}
