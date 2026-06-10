package builtin

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

const minimumRefreshInterval = 60 * time.Second

var grafanaDurationPattern = regexp.MustCompile(`^([0-9]+)\s*(ms|s|m|h|d|w|M|y)$`)

type refreshMinInterval struct {
	minimum time.Duration
}

func (refreshMinInterval) Metadata() rule.Metadata {
	return rule.Metadata{
		ID:          "core.refresh-min-interval",
		Name:        "Refresh minimum interval",
		Description: "Dashboards must not allow automatic refresh intervals below 60 seconds.",
		Severity:    rule.SeverityError,
		Source:      rule.SourceBuiltin,
		Fixable:     true,
	}
}

func (r refreshMinInterval) Check(_ context.Context, dash dashboard.Dashboard) ([]rule.Finding, error) {
	minimum := r.minimum
	if minimum == 0 {
		minimum = minimumRefreshInterval
	}
	values, err := dashboard.Values(dash.Root, "$.refresh")
	if err != nil {
		return nil, err
	}
	if len(values) == 0 || isEmptyValue(values[0]) {
		return nil, nil
	}
	refresh, ok := values[0].(string)
	if !ok {
		return nil, nil
	}
	duration, ok := parseGrafanaDuration(refresh)
	if !ok || duration >= minimum {
		return nil, nil
	}
	metadata := r.Metadata()
	return []rule.Finding{{
		RuleID:   metadata.ID,
		Severity: metadata.Severity,
		Message:  fmt.Sprintf("dashboard refresh interval %q must be at least %s", refresh, formatDuration(minimum)),
		Path:     "$.refresh",
	}}, nil
}

func (r refreshMinInterval) Fix(_ context.Context, dash *dashboard.Dashboard) ([]rule.Fix, error) {
	minimum := r.minimum
	if minimum == 0 {
		minimum = minimumRefreshInterval
	}
	object, ok := dashboardObject(dash.Root)
	if !ok {
		return nil, nil
	}
	refresh, ok := object["refresh"].(string)
	if !ok || isEmptyValue(refresh) {
		return nil, nil
	}
	duration, ok := parseGrafanaDuration(refresh)
	if !ok || duration >= minimum {
		return nil, nil
	}
	object["refresh"] = formatDuration(minimum)
	metadata := r.Metadata()
	return []rule.Fix{{RuleID: metadata.ID, File: sourceFile(dash.Source), Path: "$.refresh", Description: fmt.Sprintf("set refresh to %s", formatDuration(minimum))}}, nil
}

func (r refreshMinInterval) WithOptions(options map[string]any) (rule.Rule, error) {
	minimum := r.minimum
	if minimum == 0 {
		minimum = minimumRefreshInterval
	}
	for key, value := range options {
		switch key {
		case "min":
			text, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("min must be a duration string")
			}
			duration, ok := parseGrafanaDuration(text)
			if !ok || duration <= 0 {
				return nil, fmt.Errorf("min must be a positive Grafana duration")
			}
			minimum = duration
		default:
			return nil, fmt.Errorf("unsupported option %q", key)
		}
	}
	return refreshMinInterval{minimum: minimum}, nil
}

func formatDuration(duration time.Duration) string {
	if duration%time.Hour == 0 {
		return fmt.Sprintf("%dh", int(duration/time.Hour))
	}
	if duration%time.Minute == 0 {
		return fmt.Sprintf("%dm", int(duration/time.Minute))
	}
	if duration%time.Second == 0 {
		return fmt.Sprintf("%ds", int(duration/time.Second))
	}
	return duration.String()
}

func parseGrafanaDuration(value string) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	matches := grafanaDurationPattern.FindStringSubmatch(value)
	if matches == nil {
		return 0, false
	}
	amount, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, false
	}
	unit := matches[2]
	switch unit {
	case "ms":
		return time.Duration(amount) * time.Millisecond, true
	case "s":
		return time.Duration(amount) * time.Second, true
	case "m":
		return time.Duration(amount) * time.Minute, true
	case "h":
		return time.Duration(amount) * time.Hour, true
	case "d":
		return time.Duration(amount) * 24 * time.Hour, true
	case "w":
		return time.Duration(amount) * 7 * 24 * time.Hour, true
	case "M":
		return time.Duration(amount) * 30 * 24 * time.Hour, true
	case "y":
		return time.Duration(amount) * 365 * 24 * time.Hour, true
	default:
		return 0, false
	}
}
