package lint

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

func TestRunnerRunReturnsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Runner{Rules: []rule.Rule{neverCalledRule{}}}.Run(ctx, []dashboard.Dashboard{{Root: map[string]any{"title": "API"}}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestRunnerRunAssignsFileBackedFindingLines(t *testing.T) {
	raw := []byte(`{
  "title": "API",
  "refresh": "30s"
}`)
	var root any
	if err := json.Unmarshal(raw, &root); err != nil {
		t.Fatalf("unmarshal dashboard: %v", err)
	}

	result, err := Runner{Rules: []rule.Rule{staticFindingRule{path: "$.refresh"}}}.Run(context.Background(), []dashboard.Dashboard{{
		Source: dashboard.Source{Name: "api.json", Path: "dashboards/api.json"},
		Root:   root,
		Raw:    raw,
	}})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("expected one finding, got %#v", result.Findings)
	}
	finding := result.Findings[0]
	if finding.File != "dashboards/api.json" {
		t.Fatalf("finding file = %q, want dashboards/api.json", finding.File)
	}
	if finding.Line != 3 {
		t.Fatalf("finding line = %d, want 3", finding.Line)
	}
}

func TestRunnerRunFallsBackToLineOneForUnmappedFileBackedFindings(t *testing.T) {
	result, err := Runner{Rules: []rule.Rule{staticFindingRule{path: "$.missing"}}}.Run(context.Background(), []dashboard.Dashboard{{
		Source: dashboard.Source{Name: "api.json", Path: "dashboards/api.json"},
		Root:   map[string]any{"title": "API"},
		Raw:    []byte(`{"title":"API"}`),
	}})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("expected one finding, got %#v", result.Findings)
	}
	if result.Findings[0].Line != 1 {
		t.Fatalf("finding line = %d, want 1", result.Findings[0].Line)
	}
}

func TestRunnerRunDoesNotAssignLineForStdinFindings(t *testing.T) {
	result, err := Runner{Rules: []rule.Rule{staticFindingRule{path: "$.title"}}}.Run(context.Background(), []dashboard.Dashboard{{
		Source: dashboard.Source{Name: "<stdin>", Stdin: true},
		Root:   map[string]any{"title": "API"},
	}})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("expected one finding, got %#v", result.Findings)
	}
	if result.Findings[0].Line != 0 {
		t.Fatalf("stdin finding line = %d, want 0", result.Findings[0].Line)
	}
}

type staticFindingRule struct {
	path string
}

func (r staticFindingRule) Metadata() rule.Metadata {
	return rule.Metadata{ID: "test.static", Severity: rule.SeverityWarning}
}

func (r staticFindingRule) Check(context.Context, dashboard.Dashboard) ([]rule.Finding, error) {
	return []rule.Finding{{RuleID: "test.static", Severity: rule.SeverityWarning, Message: "static finding", Path: r.path}}, nil
}

type neverCalledRule struct{}

func (neverCalledRule) Metadata() rule.Metadata {
	return rule.Metadata{ID: "test.never-called", Severity: rule.SeverityWarning}
}

func (neverCalledRule) Check(context.Context, dashboard.Dashboard) ([]rule.Finding, error) {
	panic("rule should not be called after context cancellation")
}
