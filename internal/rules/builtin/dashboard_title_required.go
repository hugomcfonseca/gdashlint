package builtin

import (
	"context"

	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

type dashboardTitleRequired struct{}

func (dashboardTitleRequired) Metadata() rule.Metadata {
	return rule.Metadata{
		ID:          "core.dashboard-title-required",
		Name:        "Dashboard title required",
		Description: "Dashboards should define a non-empty title.",
		Severity:    rule.SeverityError,
		Source:      rule.SourceBuiltin,
	}
}

func (r dashboardTitleRequired) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	values, err := dashboard.Values(dash.Root, "$.title")
	if err != nil {
		return nil, err
	}
	if len(values) == 0 || !nonEmptyString(values[0]) {
		metadata := r.Metadata()
		return []rule.Finding{{
			RuleID:   metadata.ID,
			Severity: metadata.Severity,
			Message:  "dashboard title is required",
			Path:     "$.title",
		}}, nil
	}
	return nil, nil
}
