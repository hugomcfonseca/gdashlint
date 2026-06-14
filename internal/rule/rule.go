package rule

import (
	"context"

	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
)

// SourceKind identifies where a rule came from.
type SourceKind string

const (
	// SourceBuiltin identifies a rule provided by gdashlint.
	SourceBuiltin SourceKind = "builtin"
	// SourceConfig identifies a rule loaded from configuration.
	SourceConfig SourceKind = "config"
)

// Metadata describes a lint rule.
type Metadata struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Severity    Severity   `json:"severity"`
	Source      SourceKind `json:"source"`
	Fixable     bool       `json:"fixable"`
}

// Rule checks a dashboard and returns findings.
type Rule interface {
	Metadata() Metadata
	Check(context.Context, dashboard.Dashboard) ([]Finding, error)
}

// FixOperation describes the concrete JSON value update behind a remediation.
type FixOperation struct {
	Path  string
	Value any
}

// Fix describes one automatic remediation.
type Fix struct {
	RuleID      string        `json:"rule_id"`
	File        string        `json:"file"`
	Path        string        `json:"path"`
	Description string        `json:"description"`
	Operation   *FixOperation `json:"-"`
}

// FixableRule is implemented by rules that can safely remediate findings.
type FixableRule interface {
	Rule
	Fix(context.Context, *dashboard.Dashboard) ([]Fix, error)
}
