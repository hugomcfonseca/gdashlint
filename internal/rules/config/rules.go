package configrules

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	appconfig "github.com/hugomcfonseca/gdashlint/internal/config"
	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

type requiredRule struct {
	metadata rule.Metadata
	path     string
	message  string
}

func (r requiredRule) Metadata() rule.Metadata { return r.metadata }

func (r requiredRule) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	exists, err := dashboard.Exists(dash.Root, r.path)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, nil
	}
	return []rule.Finding{r.finding()}, nil
}

func (r requiredRule) finding() rule.Finding {
	return rule.Finding{RuleID: r.metadata.ID, Severity: r.metadata.Severity, Message: r.message, Path: r.path}
}

type forbiddenRule struct {
	metadata rule.Metadata
	path     string
	message  string
}

func (r forbiddenRule) Metadata() rule.Metadata { return r.metadata }

func (r forbiddenRule) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	exists, err := dashboard.Exists(dash.Root, r.path)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	return []rule.Finding{{RuleID: r.metadata.ID, Severity: r.metadata.Severity, Message: r.message, Path: r.path}}, nil
}

type matchRule struct {
	metadata rule.Metadata
	path     string
	pattern  *regexp.Regexp
	message  string
}

func newMatchRule(metadata rule.Metadata, cfg appconfig.CustomRule) (rule.Rule, error) {
	if cfg.Pattern == "" {
		return nil, fmt.Errorf("custom rule %s: pattern is required", metadata.ID)
	}
	pattern, err := regexp.Compile(cfg.Pattern)
	if err != nil {
		return nil, fmt.Errorf("custom rule %s: invalid pattern: %w", metadata.ID, err)
	}
	return matchRule{metadata: metadata, path: cfg.Path, pattern: pattern, message: messageOrDefault(cfg.Message, fmt.Sprintf("%s must match %s", cfg.Path, cfg.Pattern))}, nil
}

func (r matchRule) Metadata() rule.Metadata { return r.metadata }

func (r matchRule) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	values, err := dashboard.Values(dash.Root, r.path)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	findings := make([]rule.Finding, 0)
	for _, value := range values {
		text, ok := value.(string)
		if !ok || !r.pattern.MatchString(text) {
			findings = append(findings, rule.Finding{RuleID: r.metadata.ID, Severity: r.metadata.Severity, Message: r.message, Path: r.path})
		}
	}
	return findings, nil
}

type oneOfRule struct {
	metadata rule.Metadata
	path     string
	values   []any
	message  string
}

func (r oneOfRule) Metadata() rule.Metadata { return r.metadata }

func (r oneOfRule) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	values, err := dashboard.Values(dash.Root, r.path)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	for _, value := range values {
		if !containsJSONEqual(r.values, value) {
			return []rule.Finding{{RuleID: r.metadata.ID, Severity: r.metadata.Severity, Message: r.message, Path: r.path}}, nil
		}
	}
	return nil, nil
}

func containsJSONEqual(allowed []any, value any) bool {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return false
	}
	for _, allowedValue := range allowed {
		allowedJSON, err := json.Marshal(allowedValue)
		if err == nil && string(allowedJSON) == string(valueJSON) {
			return true
		}
	}
	return false
}
