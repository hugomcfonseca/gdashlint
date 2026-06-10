package builtin

import (
	"context"

	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

type panelTypeRequired struct{}

func (panelTypeRequired) Metadata() rule.Metadata {
	return rule.Metadata{
		ID:          "core.panel-type-required",
		Name:        "Panel type required",
		Description: "Panels should define a non-empty type.",
		Severity:    rule.SeverityWarning,
		Source:      rule.SourceBuiltin,
	}
}

func (r panelTypeRequired) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	metadata := r.Metadata()
	findings := make([]rule.Finding, 0)
	for _, panel := range collectPanels(dash.Root) {
		if !nonEmptyString(panel.object["type"]) {
			findings = append(findings, rule.Finding{
				RuleID:   metadata.ID,
				Severity: metadata.Severity,
				Message:  "panel type is required",
				Path:     panel.path + ".type",
			})
		}
	}
	return findings, nil
}
