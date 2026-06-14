package builtin

import (
	"context"
	"fmt"

	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

type variableCurrentEmpty struct{}

func (variableCurrentEmpty) Metadata() rule.Metadata {
	return rule.Metadata{
		ID:          "core.variable-current-empty",
		Name:        "Variable current values empty",
		Description: "Dashboard variables should not persist current selections in committed JSON.",
		Severity:    rule.SeverityWarning,
		Source:      rule.SourceBuiltin,
		Fixable:     true,
	}
}

func (r variableCurrentEmpty) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	values, err := dashboard.Values(dash.Root, "$.templating.list")
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	variables, ok := values[0].([]any)
	if !ok {
		return nil, nil
	}

	metadata := r.Metadata()
	findings := make([]rule.Finding, 0)
	for index, variable := range variables {
		object, ok := variable.(map[string]any)
		if !ok {
			continue
		}
		current, ok := object["current"]
		if !ok || isEmptyValue(current) {
			continue
		}
		findings = append(findings, rule.Finding{
			RuleID:   metadata.ID,
			Severity: metadata.Severity,
			Message:  "variable current value must be empty",
			Path:     fmt.Sprintf("$.templating.list[%d].current", index),
		})
	}
	return findings, nil
}

func (r variableCurrentEmpty) Fix(_ context.Context, dash *dashboard.Dashboard) ([]rule.Fix, error) {
	values, err := dashboard.Values(dash.Root, "$.templating.list")
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	variables, ok := values[0].([]any)
	if !ok {
		return nil, nil
	}
	metadata := r.Metadata()
	fixes := make([]rule.Fix, 0)
	for index, variable := range variables {
		object, ok := variable.(map[string]any)
		if !ok {
			continue
		}
		current, ok := object["current"]
		if !ok || isEmptyValue(current) {
			continue
		}
		path := fmt.Sprintf("$.templating.list[%d].current", index)
		value := map[string]any{}
		object["current"] = value
		fixes = append(fixes, rule.Fix{RuleID: metadata.ID, File: sourceFile(dash.Source), Path: path, Description: "clear variable current value", Operation: &rule.FixOperation{Path: path, Value: value}})
	}
	return fixes, nil
}
