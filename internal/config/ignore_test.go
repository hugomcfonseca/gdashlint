package config

import (
	"testing"

	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

func TestApplyIgnoresMatchesRecursiveGlob(t *testing.T) {
	findings := []rule.Finding{
		{RuleID: "core.panel-title-required", File: "dashboards/experimental/api/service.json", Path: "$.panels[0].title"},
		{RuleID: "core.panel-title-required", File: "dashboards/prod/api.json", Path: "$.panels[0].title"},
	}

	kept := ApplyIgnores(findings, []Ignore{{
		Rule:  "core.panel-title-required",
		Paths: []string{"dashboards/experimental/**"},
	}})

	if len(kept) != 1 {
		t.Fatalf("expected 1 kept finding, got %d", len(kept))
	}
	if kept[0].File != "dashboards/prod/api.json" {
		t.Fatalf("unexpected kept finding: %#v", kept[0])
	}
}
