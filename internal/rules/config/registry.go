package configrules

import (
	"fmt"
	"strings"

	appconfig "github.com/hugomcfonseca/gdashlint/internal/config"
	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

// Register registers config-defined rules.
func Register(registry *rule.Registry, rules map[string]appconfig.CustomRule) error {
	for id, cfg := range rules {
		if !strings.Contains(id, ".") {
			id = "custom." + id
		}
		lintRule, err := New(id, cfg)
		if err != nil {
			return err
		}
		if err := registry.Register(lintRule); err != nil {
			return err
		}
	}
	return nil
}

// New creates a config-defined rule from config.
func New(id string, cfg appconfig.CustomRule) (rule.Rule, error) {
	severity := rule.SeverityWarning
	if cfg.Severity != "" {
		parsed, err := rule.ParseSeverity(cfg.Severity)
		if err != nil {
			return nil, err
		}
		severity = parsed
	}
	metadata := rule.Metadata{
		ID:          id,
		Name:        id,
		Description: cfg.Message,
		Severity:    severity,
		Source:      rule.SourceConfig,
	}
	compiledPath, err := dashboard.CompilePath(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("custom rule %s: invalid path: %w", id, err)
	}

	switch cfg.Type {
	case "required":
		return requiredRule{metadata: metadata, path: cfg.Path, compiled: compiledPath, message: messageOrDefault(cfg.Message, fmt.Sprintf("%s is required", cfg.Path))}, nil
	case "forbidden":
		return forbiddenRule{metadata: metadata, path: cfg.Path, compiled: compiledPath, message: messageOrDefault(cfg.Message, fmt.Sprintf("%s is forbidden", cfg.Path))}, nil
	case "match":
		return newMatchRule(metadata, cfg, compiledPath)
	case "oneOf":
		return newOneOfRule(metadata, cfg, compiledPath)
	default:
		return nil, fmt.Errorf("custom rule %s: unsupported type %q", id, cfg.Type)
	}
}

func messageOrDefault(message string, fallback string) string {
	if message != "" {
		return message
	}
	return fallback
}
