// Package output renders lint and fix results in user-facing formats.
package output

import (
	"fmt"
	"io"

	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

// FixSummary writes a concise human-readable summary of automatic remediations.
func FixSummary(writer io.Writer, fixes []rule.Fix, changedFiles int, dryRun bool) error {
	verb := "Applied"
	if dryRun {
		verb = "Would apply"
	}
	_, err := fmt.Fprintf(writer, "%s %d fix(es) across %d dashboard(s).\n\n", verb, len(fixes), changedFiles)
	return err
}

// PostFixSummary writes a concise human-readable summary of findings after fixes.
func PostFixSummary(writer io.Writer, files int, findings int, errors int, warnings int, infos int, dryRun bool) error {
	prefix := "After fixes"
	if dryRun {
		prefix = "After simulated fixes"
	}
	_, err := fmt.Fprintf(writer, "%s: %d finding(s) remain in %d dashboard(s): %d error(s), %d warning(s), %d info(s).\n", prefix, findings, files, errors, warnings, infos)
	return err
}
