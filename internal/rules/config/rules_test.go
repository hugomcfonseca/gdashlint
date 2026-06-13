package configrules

import (
	"context"
	"testing"

	appconfig "github.com/hugomcfonseca/gdashlint/internal/config"
	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
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
