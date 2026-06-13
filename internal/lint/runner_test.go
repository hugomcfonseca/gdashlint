package lint

import (
	"context"
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

type neverCalledRule struct{}

func (neverCalledRule) Metadata() rule.Metadata {
	return rule.Metadata{ID: "test.never-called", Severity: rule.SeverityWarning}
}

func (neverCalledRule) Check(context.Context, dashboard.Dashboard) ([]rule.Finding, error) {
	panic("rule should not be called after context cancellation")
}
