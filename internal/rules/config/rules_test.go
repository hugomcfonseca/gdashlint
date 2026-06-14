package configrules

import (
	"context"
	"testing"

	appconfig "github.com/hugomcfonseca/gdashlint/internal/config"
	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

func TestRequiredRule(t *testing.T) {
	lintRule, err := New("custom.tags-required", appconfig.CustomRule{
		Type:     "required",
		Severity: "error",
		Path:     "$.tags",
		Message:  "tags required",
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	findings, err := lintRule.Check(context.Background(), dashboard.Dashboard{Root: map[string]any{"title": "API"}})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(findings) != 1 || findings[0].RuleID != "custom.tags-required" {
		t.Fatalf("unexpected findings: %#v", findings)
	}
}

func TestMatchRule(t *testing.T) {
	lintRule, err := New("custom.title-prefix", appconfig.CustomRule{
		Type:    "match",
		Path:    "$.title",
		Pattern: "^Team: .+",
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	findings, err := lintRule.Check(context.Background(), dashboard.Dashboard{Root: map[string]any{"title": "API"}})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected finding, got %#v", findings)
	}
}

func TestOneOfRule(t *testing.T) {
	lintRule, err := New("custom.timezone", appconfig.CustomRule{
		Type:   "oneOf",
		Path:   "$.timezone",
		Values: []any{"browser", "utc"},
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	findings, err := lintRule.Check(context.Background(), dashboard.Dashboard{Root: map[string]any{"timezone": "local"}})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected finding, got %#v", findings)
	}

	findings, err = lintRule.Check(context.Background(), dashboard.Dashboard{Root: map[string]any{"timezone": "utc"}})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %#v", findings)
	}
}

func TestCustomRuleFixSet(t *testing.T) {
	lintRule, err := New("custom.timezone", appconfig.CustomRule{
		Type:   "oneOf",
		Path:   "$.timezone",
		Values: []any{"browser", "utc"},
		Fix:    &appconfig.CustomRuleFix{Action: "set", Value: "browser"},
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if !lintRule.Metadata().Fixable {
		t.Fatalf("expected rule to be marked fixable")
	}
	fixable, ok := lintRule.(rule.FixableRule)
	if !ok {
		t.Fatalf("expected custom rule to implement FixableRule")
	}
	dash := dashboard.Dashboard{Source: dashboard.Source{Name: "dashboard.json"}, Root: map[string]any{"timezone": "local"}}

	fixes, err := fixable.Fix(context.Background(), &dash)
	if err != nil {
		t.Fatalf("Fix returned error: %v", err)
	}
	if len(fixes) != 1 || fixes[0].RuleID != "custom.timezone" || fixes[0].Path != "$.timezone" {
		t.Fatalf("unexpected fixes: %#v", fixes)
	}
	if dash.Root.(map[string]any)["timezone"] != "browser" {
		t.Fatalf("expected timezone to be fixed, got %#v", dash.Root)
	}
}

func TestCustomRuleFixSetDefault(t *testing.T) {
	lintRule, err := New("custom.timezone", appconfig.CustomRule{
		Type: "required",
		Path: "$.timezone",
		Fix:  &appconfig.CustomRuleFix{Action: "setDefault", Value: "browser"},
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	fixable := lintRule.(rule.FixableRule)
	dash := dashboard.Dashboard{Root: map[string]any{"title": "API"}}

	fixes, err := fixable.Fix(context.Background(), &dash)
	if err != nil {
		t.Fatalf("Fix returned error: %v", err)
	}
	if len(fixes) != 1 {
		t.Fatalf("expected one fix, got %#v", fixes)
	}
	if dash.Root.(map[string]any)["timezone"] != "browser" {
		t.Fatalf("expected timezone default, got %#v", dash.Root)
	}
}

func TestCustomRuleFixDoesNotApplyWhenRulePasses(t *testing.T) {
	lintRule, err := New("custom.timezone", appconfig.CustomRule{
		Type: "required",
		Path: "$.timezone",
		Fix:  &appconfig.CustomRuleFix{Action: "set", Value: "browser"},
	})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	fixable := lintRule.(rule.FixableRule)
	dash := dashboard.Dashboard{Root: map[string]any{"timezone": "utc"}}

	fixes, err := fixable.Fix(context.Background(), &dash)
	if err != nil {
		t.Fatalf("Fix returned error: %v", err)
	}
	if len(fixes) != 0 {
		t.Fatalf("expected no fixes, got %#v", fixes)
	}
	if dash.Root.(map[string]any)["timezone"] != "utc" {
		t.Fatalf("expected timezone to remain unchanged, got %#v", dash.Root)
	}
}
