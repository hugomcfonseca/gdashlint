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
	fix      *customFix
	message  string
}

func (r requiredRule) Metadata() rule.Metadata { return r.metadata }

func (r requiredRule) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	if r.compiled.Exists(dash.Root) {
		return nil, nil
	}
	return []rule.Finding{r.finding()}, nil
}

func (r requiredRule) Fix(ctx context.Context, dash *dashboard.Dashboard) ([]rule.Fix, error) {
	return fixIfFinding(ctx, dash, r, r.fix, r.message)
}

func (r requiredRule) finding() rule.Finding {
	return rule.Finding{RuleID: r.metadata.ID, Severity: r.metadata.Severity, Message: r.message, Path: r.path}
}

type forbiddenRule struct {
	metadata rule.Metadata
	path     string
	compiled dashboard.CompiledPath
	fix      *customFix
	message  string
}

func (r forbiddenRule) Metadata() rule.Metadata { return r.metadata }

func (r forbiddenRule) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	if !r.compiled.Exists(dash.Root) {
		return nil, nil
	}
	return []rule.Finding{{RuleID: r.metadata.ID, Severity: r.metadata.Severity, Message: r.message, Path: r.path}}, nil
}

func (r forbiddenRule) Fix(ctx context.Context, dash *dashboard.Dashboard) ([]rule.Fix, error) {
	return fixIfFinding(ctx, dash, r, r.fix, r.message)
}

type matchRule struct {
	metadata rule.Metadata
	path     string
	compiled dashboard.CompiledPath
	pattern  *regexp.Regexp
	fix      *customFix
	message  string
}

func newMatchRule(metadata rule.Metadata, cfg appconfig.CustomRule, compiled dashboard.CompiledPath, fix *customFix) (rule.Rule, error) {
	if cfg.Pattern == "" {
		return nil, fmt.Errorf("custom rule %s: pattern is required", metadata.ID)
	}
	pattern, err := regexp.Compile(cfg.Pattern)
	if err != nil {
		return nil, fmt.Errorf("custom rule %s: invalid pattern: %w", metadata.ID, err)
	}
	return matchRule{metadata: metadata, path: cfg.Path, compiled: compiled, pattern: pattern, fix: fix, message: messageOrDefault(cfg.Message, fmt.Sprintf("%s must match %s", cfg.Path, cfg.Pattern))}, nil
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

func (r matchRule) Fix(ctx context.Context, dash *dashboard.Dashboard) ([]rule.Fix, error) {
	return fixIfFinding(ctx, dash, r, r.fix, r.message)
}

type oneOfRule struct {
	metadata      rule.Metadata
	path          string
	compiled      dashboard.CompiledPath
	encodedValues [][]byte
	fix           *customFix
	message       string
}

func newOneOfRule(metadata rule.Metadata, cfg appconfig.CustomRule, compiled dashboard.CompiledPath, fix *customFix) (rule.Rule, error) {
	encodedValues := make([][]byte, 0, len(cfg.Values))
	for _, value := range cfg.Values {
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("custom rule %s: encode allowed value: %w", metadata.ID, err)
		}
		encodedValues = append(encodedValues, encoded)
	}
	if fix != nil && fix.path == cfg.Path && !containsJSONEqual(encodedValues, fix.value) {
		return nil, fmt.Errorf("custom rule %s: fix value must be one of the allowed values", metadata.ID)
	}
	return oneOfRule{metadata: metadata, path: cfg.Path, compiled: compiled, encodedValues: encodedValues, fix: fix, message: messageOrDefault(cfg.Message, fmt.Sprintf("%s must be one of the allowed values", cfg.Path))}, nil
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

func (r oneOfRule) Fix(ctx context.Context, dash *dashboard.Dashboard) ([]rule.Fix, error) {
	return fixIfFinding(ctx, dash, r, r.fix, r.message)
}

type customFix struct {
	action string
	path   string
	value  any
}

func newCustomFix(rulePath string, cfg *appconfig.CustomRuleFix) (*customFix, error) {
	if cfg == nil {
		return nil, nil
	}
	path := cfg.Path
	if path == "" {
		path = rulePath
	}
	if err := dashboard.ValidateWritablePath(path); err != nil {
		return nil, err
	}
	switch cfg.Action {
	case "set", "setDefault":
		if cfg.Value == nil {
			return nil, fmt.Errorf("value is required")
		}
	default:
		return nil, fmt.Errorf("unsupported action %q", cfg.Action)
	}
	return &customFix{action: cfg.Action, path: path, value: cfg.Value}, nil
}

func fixIfFinding(ctx context.Context, dash *dashboard.Dashboard, lintRule rule.Rule, fix *customFix, description string) ([]rule.Fix, error) {
	if fix == nil {
		return nil, nil
	}
	findings, err := lintRule.Check(ctx, *dash)
	if err != nil {
		return nil, err
	}
	if len(findings) == 0 {
		return nil, nil
	}

	changed := true
	switch fix.action {
	case "set":
		dash.Root, err = dashboard.Set(dash.Root, fix.path, fix.value)
	case "setDefault":
		dash.Root, changed, err = dashboard.SetDefault(dash.Root, fix.path, fix.value)
	default:
		return nil, fmt.Errorf("unsupported action %q", fix.action)
	}
	if err != nil {
		return nil, err
	}
	if !changed {
		return nil, nil
	}
	metadata := lintRule.Metadata()
	return []rule.Fix{{RuleID: metadata.ID, File: sourceFile(dash.Source), Path: fix.path, Description: customFixDescription(fix, description)}}, nil
}

func customFixDescription(fix *customFix, fallback string) string {
	if fallback != "" {
		return fallback
	}
	if fix.action == "setDefault" {
		return fmt.Sprintf("set default %s", fix.path)
	}
	return fmt.Sprintf("set %s", fix.path)
}

func sourceFile(source dashboard.Source) string {
	if source.Path != "" {
		return source.Path
	}
	return source.Name
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
