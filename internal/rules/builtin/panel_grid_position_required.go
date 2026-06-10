package builtin

import (
	"context"

	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

type panelGridPositionRequired struct{}

func (panelGridPositionRequired) Metadata() rule.Metadata {
	return rule.Metadata{
		ID:          "core.panel-grid-position-required",
		Name:        "Panel grid position required",
		Description: "Panels should define gridPos for stable dashboard layout.",
		Severity:    rule.SeverityWarning,
		Source:      rule.SourceBuiltin,
	}
}

func (r panelGridPositionRequired) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	metadata := r.Metadata()
	findings := make([]rule.Finding, 0)
	for _, panel := range collectPanels(dash.Root) {
		if _, ok := panel.object["gridPos"].(map[string]any); !ok {
			findings = append(findings, rule.Finding{
				RuleID:   metadata.ID,
				Severity: metadata.Severity,
				Message:  "panel gridPos is required",
				Path:     panel.path + ".gridPos",
			})
		}
	}
	return findings, nil
}
