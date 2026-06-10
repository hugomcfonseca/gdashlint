package builtin

import (
	"context"

	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

type panelTitleRequired struct{}

func (panelTitleRequired) Metadata() rule.Metadata {
	return rule.Metadata{
		ID:          "core.panel-title-required",
		Name:        "Panel title required",
		Description: "Panels should define non-empty titles.",
		Severity:    rule.SeverityWarning,
		Source:      rule.SourceBuiltin,
	}
}

func (r panelTitleRequired) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	metadata := r.Metadata()
	findings := make([]rule.Finding, 0)
	for _, panel := range collectPanels(dash.Root) {
		if !nonEmptyString(panel.object["title"]) {
			findings = append(findings, rule.Finding{
				RuleID:   metadata.ID,
				Severity: metadata.Severity,
				Message:  "panel title is required",
				Path:     panel.path + ".title",
			})
		}
	}
	return findings, nil
}
