// Package builtin registers gdashlint's built-in dashboard lint rules.
package builtin

import (
	"context"

	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

type dashboardNotEditable struct{}

func (dashboardNotEditable) Metadata() rule.Metadata {
	return rule.Metadata{
		ID:          "core.dashboard-not-editable",
		Name:        "Dashboard not editable",
		Description: "Dashboards should not be editable when committed to source control.",
		Severity:    rule.SeverityError,
		Source:      rule.SourceBuiltin,
		Fixable:     true,
	}
}

func (r dashboardNotEditable) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	values, err := dashboard.Values(dash.Root, "$.editable")
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	editable, ok := values[0].(bool)
	if ok && !editable {
		return nil, nil
	}
	metadata := r.Metadata()
	return []rule.Finding{{RuleID: metadata.ID, Severity: metadata.Severity, Message: "dashboard must not be editable", Path: "$.editable"}}, nil
}

func (r dashboardNotEditable) Fix(_ context.Context, dash *dashboard.Dashboard) ([]rule.Fix, error) {
	object, ok := dashboardObject(dash.Root)
	if !ok {
		return nil, nil
	}
	editable, ok := object["editable"].(bool)
	if !ok || !editable {
		return nil, nil
	}
	path := "$.editable"
	value := false
	object["editable"] = value
	metadata := r.Metadata()
	return []rule.Fix{{RuleID: metadata.ID, File: sourceFile(dash.Source), Path: path, Description: "set editable to false", Operation: &rule.FixOperation{Path: path, Value: value}}}, nil
}
