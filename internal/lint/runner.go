// Package lint applies rules to parsed dashboards and summarizes findings.
package lint

import (
	"context"
	"fmt"

	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

// Result contains lint findings and summary details.
type Result struct {
	Summary  Summary        `json:"summary"`
	Findings []rule.Finding `json:"findings"`
	Fixes    []rule.Fix     `json:"fixes,omitempty"`
}

// Summary counts linted files and findings by severity.
type Summary struct {
	Files    int `json:"files"`
	Findings int `json:"findings"`
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Infos    int `json:"infos"`
}

// Runner applies rules to dashboards.
type Runner struct {
	Rules []rule.Rule
}

// Run applies all rules to all dashboards.
func (r Runner) Run(ctx context.Context, dashboards []dashboard.Dashboard) (Result, error) {
	var result Result
	result.Summary.Files = len(dashboards)

	for _, dash := range dashboards {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		for _, lintRule := range r.Rules {
			if err := ctx.Err(); err != nil {
				return Result{}, err
			}
			metadata := lintRule.Metadata()
			findings, err := lintRule.Check(ctx, dash)
			if err != nil {
				return Result{}, fmt.Errorf("rule %s failed for %s: %w", metadata.ID, dash.Source.Name, err)
			}
			for _, finding := range findings {
				finding.Fixable = metadata.Fixable
				if finding.File == "" {
					finding.File = dash.Source.Name
					if dash.Source.Path != "" {
						finding.File = dash.Source.Path
					}
				}
				if finding.Line < 1 && !dash.Source.Stdin && dash.Source.Path != "" {
					finding.Line = 1
					if line, ok := dashboard.LineForPath(dash.Raw, finding.Path); ok {
						finding.Line = line
					}
				}
				result.add(finding)
			}
		}
	}
	return result, nil
}

func (r *Result) add(finding rule.Finding) {
	r.Findings = append(r.Findings, finding)
	r.Summary.Findings++
	switch finding.Severity {
	case rule.SeverityError:
		r.Summary.Errors++
	case rule.SeverityWarning:
		r.Summary.Warnings++
	case rule.SeverityInfo:
		r.Summary.Infos++
	}
}

// NewResult builds a result from dashboards and findings.
func NewResult(files int, findings []rule.Finding) Result {
	result := Result{Summary: Summary{Files: files}}
	for _, finding := range findings {
		result.add(finding)
	}
	return result
}

// HasFailures reports whether any finding meets the fail threshold.
func (r Result) HasFailures(threshold rule.Severity) bool {
	for _, finding := range r.Findings {
		if finding.Severity.FailsAt(threshold) {
			return true
		}
	}
	return false
}
