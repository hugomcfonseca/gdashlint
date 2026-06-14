// Package app coordinates configuration, linting, fixing, and output rendering.
package app

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

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
	ConfigPath  string
	Format      string
	Sort        string
	FailOn      string
	Paths       []string
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	FixMode     string
	FixSuffix   string
	DryRun      bool
	AutoApprove bool
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

	fixedDashboards := cloneDashboards(run.dashboards)
	fixes, err := collectFixes(ctx, run.rules, fixedDashboards)
	if err != nil {
		return 2, err
	}
	previews, err := buildFixPreviews(run.dashboards, fixedDashboards, fixes, opts)
	if err != nil {
		return 2, err
	}
	result, err := lintResult(ctx, run.rules, fixedDashboards, run.cfg)
	if err != nil {
		return 2, err
	}

	if run.cfg.Output.Format == "text" {
		if opts.DryRun && len(fixes) > 0 {
			if err := output.FixSummary(opts.Stderr, fixes, len(previews), true); err != nil {
				return 2, err
			}
		}
		if err := renderFixDiffs(opts.Stdout, previews, output.IsColorWriter(opts.Stdout)); err != nil {
			return 2, err
		}
		if err := output.PostFixSummary(opts.Stdout, result.Summary.Files, result.Summary.Findings, result.Summary.Errors, result.Summary.Warnings, result.Summary.Infos, opts.DryRun); err != nil {
			return 2, err
		}
	} else {
		result.Fixes = fixes
		if err := output.Render(opts.Stdout, result, run.cfg.Output.Format, run.cfg.Output.Sort); err != nil {
			return 2, err
		}
	}

	if !opts.DryRun && len(previews) > 0 {
		if !opts.AutoApprove {
			approved, err := confirmFixes(opts.Stdin, opts.Stderr)
			if err != nil {
				return 2, err
			}
			if !approved {
				return 1, nil
			}
		}
		if err := writeFixedDashboards(fixedDashboards, previews); err != nil {
			return 2, err
		}
		if run.cfg.Output.Format == "text" {
			if err := output.FixSummary(opts.Stderr, fixes, len(previews), false); err != nil {
				return 2, err
			}
		}
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

	metadata := make([]rule.Metadata, 0, len(registry.All()))
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
		return runData{}, fmt.Errorf("loading config: %w", err)
	}
	applyCLIOverrides(&cfg, opts)
	if err := cfg.Validate(); err != nil {
		return runData{}, fmt.Errorf("validating config: %w", err)
	}

	registry, err := buildRegistry(cfg)
	if err != nil {
		return runData{}, fmt.Errorf("building rule registry: %w", err)
	}
	rules, err := enabledRules(registry, cfg)
	if err != nil {
		return runData{}, fmt.Errorf("enabling rules: %w", err)
	}

	dashboards, err := dashboard.Loader{Stdin: opts.Stdin}.Load(opts.Paths)
	if err != nil {
		return runData{}, fmt.Errorf("loading dashboards: %w", err)
	}
	return runData{cfg: cfg, rules: rules, dashboards: dashboards}, nil
}

func lintResult(ctx context.Context, rules []rule.Rule, dashboards []dashboard.Dashboard, cfg config.Config) (lint.Result, error) {
	result, err := lint.Runner{Rules: rules}.Run(ctx, dashboards)
	if err != nil {
		return lint.Result{}, fmt.Errorf("running lint: %w", err)
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
		return nil, fmt.Errorf("registering builtin rules: %w", err)
	}
	if err := configrules.Register(registry, cfg.CustomRules); err != nil {
		return nil, fmt.Errorf("registering config rules: %w", err)
	}
	return registry, nil
}

func enabledRules(registry *rule.Registry, cfg config.Config) ([]rule.Rule, error) {
	allRules := registry.All()
	known := make(map[string]bool, len(allRules))
	for _, lintRule := range allRules {
		known[lintRule.Metadata().ID] = true
	}
	for id := range cfg.Rules {
		if !known[id] {
			return nil, fmt.Errorf("unknown rule %q", id)
		}
	}

	rules := make([]rule.Rule, 0, len(allRules))
	for _, lintRule := range allRules {
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
				return nil, fmt.Errorf("parsing severity for rule %s: %w", metadata.ID, err)
			}
			lintRule = newSeverityRule(lintRule, severity)
		}
		rules = append(rules, lintRule)
	}
	return rules, nil
}

type severityRule struct {
	rule.Rule
	severity rule.Severity
}

// newSeverityRule constructs a severityRule with nil check.
func newSeverityRule(r rule.Rule, severity rule.Severity) rule.Rule {
	if r == nil {
		return nil
	}
	return severityRule{Rule: r, severity: severity}
}

func (r severityRule) Metadata() rule.Metadata {
	if r.Rule == nil {
		return rule.Metadata{}
	}
	metadata := r.Rule.Metadata()
	metadata.Severity = r.severity
	return metadata
}

func (r severityRule) Check(ctx context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	if r.Rule == nil {
		return nil, nil
	}
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
	if r.Rule == nil {
		return nil, nil
	}
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

func collectFixes(ctx context.Context, rules []rule.Rule, dashboards []dashboard.Dashboard) ([]rule.Fix, error) {
	allFixes := make([]rule.Fix, 0)
	for index := range dashboards {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context cancelled while applying fixes: %w", err)
		}
		dash := &dashboards[index]

		for _, lintRule := range rules {
			if err := ctx.Err(); err != nil {
				return nil, fmt.Errorf("context cancelled while applying fixes: %w", err)
			}
			fixable, ok := lintRule.(rule.FixableRule)
			if !ok {
				continue
			}
			fixes, err := fixable.Fix(ctx, dash)
			if err != nil {
				return nil, fmt.Errorf("rule %s fix failed for %s: %w", lintRule.Metadata().ID, dash.Source.Name, err)
			}
			allFixes = append(allFixes, fixes...)
		}
	}
	return allFixes, nil
}

type fixPreview struct {
	FromPath string
	ToPath   string
	Before   []byte
	After    []byte
}

func cloneDashboards(dashboards []dashboard.Dashboard) []dashboard.Dashboard {
	clones := make([]dashboard.Dashboard, len(dashboards))
	copy(clones, dashboards)
	for index := range dashboards {
		clones[index].Root = cloneJSONValue(dashboards[index].Root)
	}
	return clones
}

func cloneJSONValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		cloned := make(map[string]any, len(typed))
		for key, nested := range typed {
			cloned[key] = cloneJSONValue(nested)
		}
		return cloned
	case []any:
		cloned := make([]any, len(typed))
		for index, nested := range typed {
			cloned[index] = cloneJSONValue(nested)
		}
		return cloned
	default:
		return typed
	}
}

func buildFixPreviews(original []dashboard.Dashboard, fixed []dashboard.Dashboard, fixes []rule.Fix, opts Options) ([]fixPreview, error) {
	previews := make([]fixPreview, 0, len(fixed))
	fixesByFile := make(map[string][]rule.Fix)
	for _, fix := range fixes {
		fixesByFile[fix.File] = append(fixesByFile[fix.File], fix)
	}
	for index := range fixed {
		before := original[index].Raw
		if len(before) == 0 {
			return nil, fmt.Errorf("dashboard %s has no source bytes", original[index].Source.Name)
		}
		fileFixes := fixesByFile[sourceFile(original[index].Source)]
		edits := make([]dashboard.TextEdit, 0, len(fileFixes))
		for _, fix := range fileFixes {
			if fix.Operation == nil {
				return nil, fmt.Errorf("fix %s for %s has no text operation", fix.RuleID, sourceFile(original[index].Source))
			}
			edits = append(edits, dashboard.TextEdit{Path: fix.Operation.Path, Value: fix.Operation.Value})
		}
		after := before
		if len(edits) > 0 {
			var err error
			after, err = dashboard.ApplyTextEdits(before, edits)
			if err != nil {
				return nil, fmt.Errorf("building fixed dashboard %s: %w", original[index].Source.Name, err)
			}
		}
		fixed[index].Raw = after
		toPath := original[index].Source.Path
		if opts.FixMode == "copy" {
			toPath = dashboard.CopyPath(toPath, opts.FixSuffix)
			fixed[index].Source.Path = toPath
			fixed[index].Source.Name = toPath
		}
		if bytes.Equal(before, after) {
			continue
		}
		previews = append(previews, fixPreview{
			FromPath: original[index].Source.Path,
			ToPath:   toPath,
			Before:   before,
			After:    after,
		})
	}
	return previews, nil
}

func renderFixDiffs(writer io.Writer, previews []fixPreview, color bool) error {
	for _, preview := range previews {
		if err := output.UnifiedDiff(writer, preview.FromPath, preview.ToPath, preview.Before, preview.After, color); err != nil {
			return err
		}
	}
	return nil
}

func confirmFixes(reader io.Reader, writer io.Writer) (bool, error) {
	if _, err := fmt.Fprint(writer, "Apply fixes? [y/N] "); err != nil {
		return false, err
	}
	line, err := bufio.NewReader(reader).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

func sourceFile(source dashboard.Source) string {
	if source.Path != "" {
		return source.Path
	}
	return source.Name
}

func writeFixedDashboards(dashboards []dashboard.Dashboard, previews []fixPreview) error {
	previewByPath := make(map[string]fixPreview, len(previews))
	for _, preview := range previews {
		previewByPath[preview.ToPath] = preview
	}
	for _, dash := range dashboards {
		if _, ok := previewByPath[dash.Source.Path]; !ok {
			continue
		}
		if err := dashboard.WriteBytes(dash.Source.Path, dash.Raw); err != nil {
			return fmt.Errorf("writing fixed dashboard to %s: %w", dash.Source.Path, err)
		}
	}
	return nil
}
