package builtin

import (
	"context"
	"fmt"

	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

type dashboardTagsRequired struct {
	min          int
	requiredTags []string
}

func (dashboardTagsRequired) Metadata() rule.Metadata {
	return rule.Metadata{
		ID:          "core.dashboard-tags-required",
		Name:        "Dashboard tags required",
		Description: "Dashboards should define at least one non-empty tag.",
		Severity:    rule.SeverityWarning,
		Source:      rule.SourceBuiltin,
	}
}

func (r dashboardTagsRequired) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	minimumTags := r.min
	if minimumTags == 0 {
		minimumTags = 1
	}
	values, err := dashboard.Values(dash.Root, "$.tags")
	if err != nil {
		return nil, err
	}
	tags := stringValues(values)
	if len(tags) < minimumTags {
		metadata := r.Metadata()
		return []rule.Finding{{RuleID: metadata.ID, Severity: metadata.Severity, Message: fmt.Sprintf("dashboard should define at least %d tag(s)", minimumTags), Path: "$.tags"}}, nil
	}
	missing := missingStrings(r.requiredTags, tags)
	if len(missing) > 0 {
		metadata := r.Metadata()
		return []rule.Finding{{RuleID: metadata.ID, Severity: metadata.Severity, Message: fmt.Sprintf("dashboard is missing required tag(s): %v", missing), Path: "$.tags"}}, nil
	}
	return nil, nil
}

func (r dashboardTagsRequired) WithOptions(options map[string]any) (rule.Rule, error) {
	minimumTags := r.min
	if minimumTags == 0 {
		minimumTags = 1
	}
	requiredTags := append([]string(nil), r.requiredTags...)
	for key, value := range options {
		switch key {
		case "min":
			number, ok := numberValue(value)
			if !ok || number < 1 || number != float64(int(number)) {
				return nil, fmt.Errorf("min must be a positive integer")
			}
			minimumTags = int(number)
		case "requiredTags":
			values, ok := value.([]any)
			if !ok {
				return nil, fmt.Errorf("requiredTags must be a list of strings")
			}
			requiredTags = make([]string, 0, len(values))
			for _, item := range values {
				text, ok := item.(string)
				if !ok || text == "" {
					return nil, fmt.Errorf("requiredTags must be a list of non-empty strings")
				}
				requiredTags = append(requiredTags, text)
			}
		default:
			return nil, fmt.Errorf("unsupported option %q", key)
		}
	}
	return dashboardTagsRequired{min: minimumTags, requiredTags: requiredTags}, nil
}

func stringValues(values []any) []string {
	if len(values) == 0 {
		return nil
	}
	items, ok := values[0].([]any)
	if !ok {
		return nil
	}
	strings := make([]string, 0, len(items))
	for _, item := range items {
		if nonEmptyString(item) {
			strings = append(strings, item.(string))
		}
	}
	return strings
}

func missingStrings(required []string, actual []string) []string {
	if len(required) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(actual))
	for _, value := range actual {
		seen[value] = true
	}
	missing := make([]string, 0)
	for _, value := range required {
		if !seen[value] {
			missing = append(missing, value)
		}
	}
	return missing
}
