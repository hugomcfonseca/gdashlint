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

func TestValidateRejectsInvalidCustomRuleFixAction(t *testing.T) {
	cfg := Default()
	cfg.CustomRules["custom.title"] = CustomRule{Type: "required", Path: "$.title", Fix: &CustomRuleFix{Action: "replace", Value: "Example"}}

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected invalid custom fix action error")
	}
}

func TestValidateRejectsWildcardCustomRuleFixPath(t *testing.T) {
	cfg := Default()
	cfg.CustomRules["custom.panel-title"] = CustomRule{Type: "required", Path: "$.panels[*].title", Fix: &CustomRuleFix{Action: "set", Path: "$.panels[*].title", Value: "Panel"}}

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected invalid wildcard fix path error")
	}
}

func TestValidateRejectsOneOfFixValueOutsideAllowedValues(t *testing.T) {
	cfg := Default()
	cfg.CustomRules["custom.refresh"] = CustomRule{Type: "oneOf", Path: "$.refresh", Values: []any{"1m", "5m"}, Fix: &CustomRuleFix{Action: "set", Value: "30s"}}

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected invalid oneOf fix value error")
	}
}

func TestValidateRejectsSetDefaultForMatchRule(t *testing.T) {
	cfg := Default()
	cfg.CustomRules["custom.title"] = CustomRule{Type: "match", Path: "$.title", Pattern: "^Team: .+", Fix: &CustomRuleFix{Action: "setDefault", Value: "Team: API"}}

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected incompatible match fix action error")
	}
}

func TestValidateRejectsForbiddenCustomRuleFix(t *testing.T) {
	cfg := Default()
	cfg.CustomRules["custom.editable"] = CustomRule{Type: "forbidden", Path: "$.editable", Fix: &CustomRuleFix{Action: "set", Value: false}}

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected forbidden rule fix error")
	}
}
