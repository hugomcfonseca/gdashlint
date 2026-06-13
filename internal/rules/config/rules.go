package configrules

import (
	"bytes"
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
	compiled dashboard.CompiledPath
	message  string
}

func (r requiredRule) Metadata() rule.Metadata { return r.metadata }

func (r requiredRule) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	if r.compiled.Exists(dash.Root) {
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
	compiled dashboard.CompiledPath
	message  string
}

func (r forbiddenRule) Metadata() rule.Metadata { return r.metadata }

func (r forbiddenRule) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	if !r.compiled.Exists(dash.Root) {
		return nil, nil
	}
	return []rule.Finding{{RuleID: r.metadata.ID, Severity: r.metadata.Severity, Message: r.message, Path: r.path}}, nil
}

type matchRule struct {
	metadata rule.Metadata
	path     string
	compiled dashboard.CompiledPath
	pattern  *regexp.Regexp
	message  string
}

func newMatchRule(metadata rule.Metadata, cfg appconfig.CustomRule, compiled dashboard.CompiledPath) (rule.Rule, error) {
	if cfg.Pattern == "" {
		return nil, fmt.Errorf("custom rule %s: pattern is required", metadata.ID)
	}
	pattern, err := regexp.Compile(cfg.Pattern)
	if err != nil {
		return nil, fmt.Errorf("custom rule %s: invalid pattern: %w", metadata.ID, err)
	}
	return matchRule{metadata: metadata, path: cfg.Path, compiled: compiled, pattern: pattern, message: messageOrDefault(cfg.Message, fmt.Sprintf("%s must match %s", cfg.Path, cfg.Pattern))}, nil
}

func (r matchRule) Metadata() rule.Metadata { return r.metadata }

func (r matchRule) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	values := r.compiled.Values(dash.Root)
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
	metadata      rule.Metadata
	path          string
	compiled      dashboard.CompiledPath
	encodedValues [][]byte
	message       string
}

func newOneOfRule(metadata rule.Metadata, cfg appconfig.CustomRule, compiled dashboard.CompiledPath) (rule.Rule, error) {
	encodedValues := make([][]byte, 0, len(cfg.Values))
	for _, value := range cfg.Values {
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("custom rule %s: encode allowed value: %w", metadata.ID, err)
		}
		encodedValues = append(encodedValues, encoded)
	}
	return oneOfRule{metadata: metadata, path: cfg.Path, compiled: compiled, encodedValues: encodedValues, message: messageOrDefault(cfg.Message, fmt.Sprintf("%s must be one of the allowed values", cfg.Path))}, nil
}

func (r oneOfRule) Metadata() rule.Metadata { return r.metadata }

func (r oneOfRule) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	values := r.compiled.Values(dash.Root)
	if len(values) == 0 {
		return nil, nil
	}
	for _, value := range values {
		if !containsJSONEqual(r.encodedValues, value) {
			return []rule.Finding{{RuleID: r.metadata.ID, Severity: r.metadata.Severity, Message: r.message, Path: r.path}}, nil
		}
	}
	return nil, nil
}

func containsJSONEqual(encodedAllowed [][]byte, value any) bool {
	encodedValue, err := json.Marshal(value)
	if err != nil {
		return false
	}
	for _, encodedAllowedValue := range encodedAllowed {
		if bytes.Equal(encodedAllowedValue, encodedValue) {
			return true
		}
	}
	return false
}
