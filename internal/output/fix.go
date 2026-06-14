// Package output renders lint and fix results in user-facing formats.
package output

import (
	"fmt"
	"io"

	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

// FixSummary writes a human-readable summary of automatic remediations.
func FixSummary(writer io.Writer, fixes []rule.Fix, dryRun bool) error {
	verb := "Applied"
	if dryRun {
		verb = "Would apply"
	}
	if _, err := fmt.Fprintf(writer, "%s %d fix(es).\n", verb, len(fixes)); err != nil {
		return err
	}
	for _, fix := range fixes {
		if _, err := fmt.Fprintf(writer, "  %s %s %s: %s\n", fix.File, fix.RuleID, fix.Path, fix.Description); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(writer)
	return err
}

// RemainingFindingsHeader writes the human-readable heading before post-fix findings.
func RemainingFindingsHeader(writer io.Writer) error {
	_, err := fmt.Fprintln(writer, "Remaining findings after fixes:")
	return err
}
