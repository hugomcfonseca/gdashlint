package config

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

// ApplyIgnores removes findings suppressed by config ignore rules.
func ApplyIgnores(findings []rule.Finding, ignores []Ignore) []rule.Finding {
	if len(ignores) == 0 || len(findings) == 0 {
		return findings
	}
	kept := make([]rule.Finding, 0, len(findings))
	for _, finding := range findings {
		if ignored(finding, ignores) {
			continue
		}
		kept = append(kept, finding)
	}
	return kept
}

func ignored(finding rule.Finding, ignores []Ignore) bool {
	for _, ignore := range ignores {
		if ignore.Rule != finding.RuleID {
			continue
		}
		if !matchAny(ignore.Paths, finding.File) {
			continue
		}
		if len(ignore.JSONPaths) > 0 && !stringIn(ignore.JSONPaths, finding.Path) {
			continue
		}
		return true
	}
	return false
}

func matchAny(patterns []string, value string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, pattern := range patterns {
		if globMatch(pattern, value) || globMatch(pattern, filepath.Base(value)) {
			return true
		}
	}
	return false
}

func globMatch(pattern string, value string) bool {
	pattern = filepath.ToSlash(pattern)
	value = filepath.ToSlash(value)

	matched, err := filepath.Match(pattern, value)
	if err == nil && matched {
		return true
	}

	regex := globRegex(pattern)
	matched, err = regexp.MatchString(regex, value)
	return err == nil && matched
}

func globRegex(pattern string) string {
	var builder strings.Builder
	builder.WriteString("^")
	for index := 0; index < len(pattern); index++ {
		char := pattern[index]
		switch char {
		case '*':
			if index+1 < len(pattern) && pattern[index+1] == '*' {
				builder.WriteString(".*")
				index++
			} else {
				builder.WriteString("[^/]*")
			}
		case '?':
			builder.WriteString("[^/]")
		case '.', '+', '(', ')', '|', '^', '$', '{', '}', '[', ']', '\\':
			builder.WriteByte('\\')
			builder.WriteByte(char)
		default:
			builder.WriteByte(char)
		}
	}
	builder.WriteString("$")
	return builder.String()
}

func stringIn(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
