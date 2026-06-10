package builtin

import (
	"context"
	"testing"

	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
)

func TestRefreshMinIntervalOptions(t *testing.T) {
	configured, err := (refreshMinInterval{}).WithOptions(map[string]any{"min": "5m"})
	if err != nil {
		t.Fatalf("WithOptions returned error: %v", err)
	}

	findings, err := configured.Check(context.Background(), dashboard.Dashboard{Root: map[string]any{"refresh": "1m"}})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected finding for refresh below configured minimum, got %#v", findings)
	}
}

func TestDashboardTagsRequiredOptions(t *testing.T) {
	configured, err := (dashboardTagsRequired{}).WithOptions(map[string]any{
		"min":          2,
		"requiredTags": []any{"team:platform"},
	})
	if err != nil {
		t.Fatalf("WithOptions returned error: %v", err)
	}

	findings, err := configured.Check(context.Background(), dashboard.Dashboard{Root: map[string]any{"tags": []any{"team:platform"}}})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected finding for fewer than configured tag minimum, got %#v", findings)
	}

	findings, err = configured.Check(context.Background(), dashboard.Dashboard{Root: map[string]any{"tags": []any{"team:platform", "service:api"}}})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}

func TestRefreshMinInterval(t *testing.T) {
	tests := []struct {
		name     string
		refresh  any
		findings int
	}{
		{name: "disabled", refresh: "", findings: 0},
		{name: "one minute", refresh: "1m", findings: 0},
		{name: "sixty seconds", refresh: "60s", findings: 0},
		{name: "thirty seconds", refresh: "30s", findings: 1},
		{name: "milliseconds", refresh: "500ms", findings: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lintRule := refreshMinInterval{}
			findings, err := lintRule.Check(context.Background(), dashboard.Dashboard{Root: map[string]any{"refresh": tt.refresh}})
			if err != nil {
				t.Fatalf("Check returned error: %v", err)
			}
			if len(findings) != tt.findings {
				t.Fatalf("expected %d findings, got %#v", tt.findings, findings)
			}
		})
	}
}

func TestDashboardNotEditable(t *testing.T) {
	lintRule := dashboardNotEditable{}
	findings, err := lintRule.Check(context.Background(), dashboard.Dashboard{Root: map[string]any{"editable": true}})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected finding, got %#v", findings)
	}

	findings, err = lintRule.Check(context.Background(), dashboard.Dashboard{Root: map[string]any{"editable": false}})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}

func TestVariableCurrentEmpty(t *testing.T) {
	root := map[string]any{
		"templating": map[string]any{
			"list": []any{
				map[string]any{"name": "empty-object", "current": map[string]any{}},
				map[string]any{"name": "empty-values", "current": map[string]any{"text": "", "value": []any{}}},
				map[string]any{"name": "persisted", "current": map[string]any{"text": "prod", "value": "prod"}},
				map[string]any{"name": "selected", "current": map[string]any{"selected": true, "text": "", "value": ""}},
			},
		},
	}

	lintRule := variableCurrentEmpty{}
	findings, err := lintRule.Check(context.Background(), dashboard.Dashboard{Root: root})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %#v", findings)
	}
}

func TestPanelRulesTraverseNestedPanels(t *testing.T) {
	root := map[string]any{
		"panels": []any{
			map[string]any{
				"type":    "row",
				"title":   "Row",
				"gridPos": map[string]any{"x": 0, "y": 0, "w": 24, "h": 1},
				"panels": []any{
					map[string]any{"gridPos": map[string]any{"x": 0, "y": 1, "w": 12, "h": 8}},
				},
			},
		},
	}

	titleFindings, err := (panelTitleRequired{}).Check(context.Background(), dashboard.Dashboard{Root: root})
	if err != nil {
		t.Fatalf("title Check returned error: %v", err)
	}
	if len(titleFindings) != 1 || titleFindings[0].Path != "$.panels[0].panels[0].title" {
		t.Fatalf("unexpected title findings: %#v", titleFindings)
	}

	typeFindings, err := (panelTypeRequired{}).Check(context.Background(), dashboard.Dashboard{Root: root})
	if err != nil {
		t.Fatalf("type Check returned error: %v", err)
	}
	if len(typeFindings) != 1 || typeFindings[0].Path != "$.panels[0].panels[0].type" {
		t.Fatalf("unexpected type findings: %#v", typeFindings)
	}
}

func TestPanelGridPositionValid(t *testing.T) {
	root := map[string]any{
		"panels": []any{
			map[string]any{"title": "valid", "type": "timeseries", "gridPos": map[string]any{"x": 0, "y": 0, "w": 12, "h": 8}},
			map[string]any{"title": "invalid", "type": "timeseries", "gridPos": map[string]any{"x": 20, "y": 0, "w": 8, "h": 8}},
		},
	}

	findings, err := (panelGridPositionValid{}).Check(context.Background(), dashboard.Dashboard{Root: root})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(findings) != 1 || findings[0].Path != "$.panels[1].gridPos" {
		t.Fatalf("unexpected findings: %#v", findings)
	}
}
