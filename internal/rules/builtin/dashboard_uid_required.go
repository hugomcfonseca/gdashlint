package builtin

import (
	"context"

	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

type dashboardUIDRequired struct{}

func (dashboardUIDRequired) Metadata() rule.Metadata {
	return rule.Metadata{
		ID:          "core.dashboard-uid-required",
		Name:        "Dashboard UID required",
		Description: "Dashboards should define a stable uid.",
		Severity:    rule.SeverityWarning,
		Source:      rule.SourceBuiltin,
	}
}

func (r dashboardUIDRequired) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	values, err := dashboard.Values(dash.Root, "$.uid")
	if err != nil {
		return nil, err
	}
	if len(values) == 0 || !nonEmptyString(values[0]) {
		metadata := r.Metadata()
		return []rule.Finding{{RuleID: metadata.ID, Severity: metadata.Severity, Message: "dashboard uid is required", Path: "$.uid"}}, nil
	}
	return nil, nil
}
