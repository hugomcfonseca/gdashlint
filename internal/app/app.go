package app

import (
	"context"
	"fmt"
	"io"

	"github.com/hugomcfonseca/gdashlint/internal/config"
	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/lint"
	"github.com/hugomcfonseca/gdashlint/internal/output"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
	"github.com/hugomcfonseca/gdashlint/internal/rules/builtin"
	configrules "github.com/hugomcfonseca/gdashlint/internal/rules/config"
)

// Options contains app-level CLI options.
type Options struct {
	ConfigPath string
	Format     string
	Sort       string
	FailOn     string
	Paths      []string
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	FixMode    string
	FixSuffix  string
	DryRun     bool
}

// Lint runs read-only dashboard linting.
func Lint(ctx context.Context, opts Options) (int, error) {
	run, err := prepareRun(opts)
	if err != nil {
		return 2, err
	}

	result, err := lintResult(ctx, run.rules, run.dashboards, run.cfg)
	if err != nil {
		return 2, err
	}
	if err := output.Render(opts.Stdout, result, run.cfg.Output.Format, run.cfg.Output.Sort); err != nil {
		return 2, err
	}
	if result.HasFailures(run.cfg.FailOnSeverity()) {
		return 1, nil
	}
	return 0, nil
}

// Fix applies safe automatic dashboard remediations and reports remaining findings.
func Fix(ctx context.Context, opts Options) (int, error) {
	run, err := prepareRun(opts)
	if err != nil {
		return 2, err
	}
	if err := validateFixOptions(opts, run.dashboards); err != nil {
		return 2, err
	}

	fixes, err := applyFixes(ctx, run.rules, run.dashboards, opts)
	if err != nil {
		return 2, err
	}
	if run.cfg.Output.Format == "text" {
		if err := output.FixSummary(opts.Stderr, fixes, opts.DryRun); err != nil {
			return 2, err
		}
	}

	result, err := lintResult(ctx, run.rules, run.dashboards, run.cfg)
	if err != nil {
		return 2, err
	}
	if run.cfg.Output.Format != "text" {
		result.Fixes = fixes
	}
	if run.cfg.Output.Format == "text" && len(result.Findings) > 0 {
		if err := output.RemainingFindingsHeader(opts.Stdout); err != nil {
			return 2, err
		}
	}
	if err := output.Render(opts.Stdout, result, run.cfg.Output.Format, run.cfg.Output.Sort); err != nil {
		return 2, err
	}
	if opts.DryRun && len(fixes) > 0 {
		return 1, nil
	}
	if result.HasFailures(run.cfg.FailOnSeverity()) {
		return 1, nil
	}
	return 0, nil
}

// Rules lists registered rules.
func Rules(opts Options) (int, error) {
	cfg, err := config.LoadDiscovered(opts.ConfigPath)
	if err != nil {
		return 2, err
	}
	applyCLIOverrides(&cfg, opts)
	if err := cfg.Validate(); err != nil {
		return 2, err
	}
	registry, err := buildRegistry(cfg)
	if err != nil {
		return 2, err
	}

	metadata := make([]rule.Metadata, 0)
	for _, lintRule := range registry.All() {
		metadata = append(metadata, lintRule.Metadata())
	}

	if err := output.Rules(opts.Stdout, metadata, cfg.Output.Format); err != nil {
		return 2, err
	}
	return 0, nil
}

type runData struct {
	cfg        config.Config
	rules      []rule.Rule
	dashboards []dashboard.Dashboard
}

func prepareRun(opts Options) (runData, error) {
	cfg, err := config.LoadDiscovered(opts.ConfigPath)
	if err != nil {
		return runData{}, err
	}
	applyCLIOverrides(&cfg, opts)
	if err := cfg.Validate(); err != nil {
		return runData{}, err
	}

	registry, err := buildRegistry(cfg)
	if err != nil {
		return runData{}, err
	}
	rules, err := enabledRules(registry, cfg)
	if err != nil {
		return runData{}, err
	}

	dashboards, err := dashboard.Loader{Stdin: opts.Stdin}.Load(opts.Paths)
	if err != nil {
		return runData{}, err
	}
	return runData{cfg: cfg, rules: rules, dashboards: dashboards}, nil
}

func lintResult(ctx context.Context, rules []rule.Rule, dashboards []dashboard.Dashboard, cfg config.Config) (lint.Result, error) {
	result, err := lint.Runner{Rules: rules}.Run(ctx, dashboards)
	if err != nil {
		return lint.Result{}, err
	}
	filtered := config.ApplyIgnores(result.Findings, cfg.Ignore)
	return lint.NewResult(result.Summary.Files, filtered), nil
}

func applyCLIOverrides(cfg *config.Config, opts Options) {
	if opts.Format != "" {
		cfg.Output.Format = opts.Format
	}
	if opts.Sort != "" {
		cfg.Output.Sort = opts.Sort
	}
	if opts.FailOn != "" {
		cfg.FailOn = opts.FailOn
	}
}

func buildRegistry(cfg config.Config) (*rule.Registry, error) {
	registry := rule.NewRegistry()
	if err := builtin.Register(registry); err != nil {
		return nil, err
	}
	if err := configrules.Register(registry, cfg.CustomRules); err != nil {
		return nil, err
	}
	return registry, nil
}

func enabledRules(registry *rule.Registry, cfg config.Config) ([]rule.Rule, error) {
	known := make(map[string]bool)
	for _, lintRule := range registry.All() {
		known[lintRule.Metadata().ID] = true
	}
	for id := range cfg.Rules {
		if !known[id] {
			return nil, fmt.Errorf("unknown rule %q", id)
		}
	}

	var rules []rule.Rule
	for _, lintRule := range registry.All() {
		metadata := lintRule.Metadata()
		override, ok := cfg.Rules[metadata.ID]
		if ok && override.Enabled != nil && !*override.Enabled {
			continue
		}
		if ok && len(override.Options) > 0 {
			configured, configurable := lintRule.(interface {
				WithOptions(map[string]any) (rule.Rule, error)
			})
			if !configurable {
				return nil, fmt.Errorf("rule %q does not support options", metadata.ID)
			}
			var err error
			lintRule, err = configured.WithOptions(override.Options)
			if err != nil {
				return nil, fmt.Errorf("rule %s options: %w", metadata.ID, err)
			}
		}
		if ok && override.Severity != "" {
			severity, err := rule.ParseSeverity(override.Severity)
			if err != nil {
				return nil, err
			}
			lintRule = severityRule{Rule: lintRule, severity: severity}
		}
		rules = append(rules, lintRule)
	}
	return rules, nil
}

type severityRule struct {
	rule.Rule
	severity rule.Severity
}

func (r severityRule) Metadata() rule.Metadata {
	metadata := r.Rule.Metadata()
	metadata.Severity = r.severity
	return metadata
}

func (r severityRule) Check(ctx context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	findings, err := r.Rule.Check(ctx, dash)
	if err != nil {
		return nil, err
	}
	for index := range findings {
		findings[index].Severity = r.severity
	}
	return findings, nil
}

func (r severityRule) Fix(ctx context.Context, dash *dashboard.Dashboard) ([]rule.Fix, error) {
	fixable, ok := r.Rule.(rule.FixableRule)
	if !ok {
		return nil, nil
	}
	return fixable.Fix(ctx, dash)
}

func validateFixOptions(opts Options, dashboards []dashboard.Dashboard) error {
	mode := opts.FixMode
	if mode == "" {
		mode = "in-place"
	}
	if mode != "in-place" && mode != "copy" {
		return fmt.Errorf("invalid --mode %q", opts.FixMode)
	}
	if mode == "copy" {
		if err := dashboard.ValidateCopySuffix(opts.FixSuffix); err != nil {
			return err
		}
	}
	for _, dash := range dashboards {
		if dash.Source.Stdin {
			return fmt.Errorf("fix does not support stdin input yet")
		}
		if dash.Source.Path == "" {
			return fmt.Errorf("fix requires file-backed dashboard inputs")
		}
	}
	return nil
}

func applyFixes(ctx context.Context, rules []rule.Rule, dashboards []dashboard.Dashboard, opts Options) ([]rule.Fix, error) {
	allFixes := make([]rule.Fix, 0)
	for index := range dashboards {
		dash := &dashboards[index]

		fileFixes := make([]rule.Fix, 0)
		for _, lintRule := range rules {
			fixable, ok := lintRule.(rule.FixableRule)
			if !ok {
				continue
			}
			fixes, err := fixable.Fix(ctx, dash)
			if err != nil {
				return nil, fmt.Errorf("rule %s fix failed for %s: %w", lintRule.Metadata().ID, dash.Source.Name, err)
			}
			fileFixes = append(fileFixes, fixes...)
		}
		allFixes = append(allFixes, fileFixes...)
		if opts.DryRun || len(fileFixes) == 0 {
			continue
		}
		path := dashboards[index].Source.Path
		if opts.FixMode == "copy" {
			path = dashboard.CopyPath(path, opts.FixSuffix)
			dashboards[index].Source.Path = path
			dashboards[index].Source.Name = path
		}
		if err := dashboard.WriteFile(path, dashboards[index]); err != nil {
			return nil, err
		}
	}
	return allFixes, nil
}
