package output

import (
	"sort"

	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

// Sort sorts findings in-place.
func Sort(findings []rule.Finding, mode string) {
	sort.SliceStable(findings, func(i, j int) bool {
		left := findings[i]
		right := findings[j]
		if mode == "file" && left.File != right.File {
			return left.File < right.File
		}
		if left.Severity != right.Severity {
			return left.Severity.Rank() > right.Severity.Rank()
		}
		if left.File != right.File {
			return left.File < right.File
		}
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		return left.RuleID < right.RuleID
	})
}
