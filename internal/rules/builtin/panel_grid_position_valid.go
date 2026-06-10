package builtin

import (
	"context"
	"fmt"

	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

type panelGridPositionValid struct{}

func (panelGridPositionValid) Metadata() rule.Metadata {
	return rule.Metadata{
		ID:          "core.panel-grid-position-valid",
		Name:        "Panel grid position valid",
		Description: "Panel gridPos values should be valid for Grafana's 24-column grid.",
		Severity:    rule.SeverityError,
		Source:      rule.SourceBuiltin,
	}
}

func (r panelGridPositionValid) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	metadata := r.Metadata()
	findings := make([]rule.Finding, 0)
	for _, panel := range collectPanels(dash.Root) {
		gridPos, ok := panel.object["gridPos"].(map[string]any)
		if !ok {
			continue
		}
		if message := validateGridPos(gridPos); message != "" {
			findings = append(findings, rule.Finding{
				RuleID:   metadata.ID,
				Severity: metadata.Severity,
				Message:  message,
				Path:     panel.path + ".gridPos",
			})
		}
	}
	return findings, nil
}

func validateGridPos(gridPos map[string]any) string {
	w, ok := requiredNumber(gridPos, "w")
	if !ok || w <= 0 {
		return "panel gridPos.w must be greater than 0"
	}
	h, ok := requiredNumber(gridPos, "h")
	if !ok || h <= 0 {
		return "panel gridPos.h must be greater than 0"
	}
	x, ok := requiredNumber(gridPos, "x")
	if !ok || x < 0 {
		return "panel gridPos.x must be greater than or equal to 0"
	}
	y, ok := requiredNumber(gridPos, "y")
	if !ok || y < 0 {
		return "panel gridPos.y must be greater than or equal to 0"
	}
	if x+w > 24 {
		return fmt.Sprintf("panel gridPos.x + gridPos.w must be less than or equal to 24, got %.0f", x+w)
	}
	return ""
}

func requiredNumber(object map[string]any, key string) (float64, bool) {
	value, ok := object[key]
	if !ok {
		return 0, false
	}
	return numberValue(value)
}
