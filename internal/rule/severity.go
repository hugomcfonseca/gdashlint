package rule

import (
	"fmt"
	"strings"
)

// Severity describes the importance of a lint finding.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
	SeverityNone    Severity = "none"
)

// ParseSeverity parses a user-provided severity value.
func ParseSeverity(value string) (Severity, error) {
	switch Severity(strings.ToLower(strings.TrimSpace(value))) {
	case SeverityError:
		return SeverityError, nil
	case SeverityWarning:
		return SeverityWarning, nil
	case SeverityInfo:
		return SeverityInfo, nil
	case SeverityNone:
		return SeverityNone, nil
	default:
		return "", fmt.Errorf("invalid severity %q", value)
	}
}

// AllowsFailThreshold reports whether this severity can be used as a fail threshold.
func (s Severity) AllowsFailThreshold() bool {
	return s == SeverityError || s == SeverityWarning || s == SeverityInfo || s == SeverityNone
}

// Rank returns a numeric rank where higher values are more severe.
func (s Severity) Rank() int {
	switch s {
	case SeverityError:
		return 3
	case SeverityWarning:
		return 2
	case SeverityInfo:
		return 1
	default:
		return 0
	}
}

// FailsAt reports whether s meets threshold.
func (s Severity) FailsAt(threshold Severity) bool {
	if threshold == SeverityNone {
		return false
	}
	return s.Rank() >= threshold.Rank()
}
