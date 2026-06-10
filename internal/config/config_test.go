package config

import "testing"

func TestValidateRejectsInvalidCustomRulePath(t *testing.T) {
	cfg := Default()
	cfg.CustomRules["custom.bad"] = CustomRule{Type: "required", Path: "title"}

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected invalid custom path error")
	}
}

func TestValidateRejectsInvalidIgnoreJSONPath(t *testing.T) {
	cfg := Default()
	cfg.Ignore = []Ignore{{Rule: "core.dashboard-title-required", JSONPaths: []string{"$.panels["}}}

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected invalid ignore jsonPath error")
	}
}

func TestValidateAllowsGithubOutput(t *testing.T) {
	cfg := Default()
	cfg.Output.Format = "github"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
}
