package output

import (
	"fmt"
	"io"

	"github.com/hugomcfonseca/gdashlint/internal/lint"
)

// Text writes human-readable output.
func Text(writer io.Writer, result lint.Result, sortMode string) error {
	if len(result.Findings) == 0 {
		_, err := fmt.Fprintf(writer, "No findings across %d dashboard(s).\n", result.Summary.Files)
		return err
	}

	if sortMode == "file" {
		lastFile := ""
		for _, finding := range result.Findings {
			if finding.File != lastFile {
				if lastFile != "" {
					fmt.Fprintln(writer)
				}
				fmt.Fprintf(writer, "%s\n", finding.File)
				lastFile = finding.File
			}
			fmt.Fprintf(writer, "  %-7s %-7s %-36s %-24s %s\n", finding.Severity, fixabilityLabel(finding.Fixable), finding.RuleID, finding.Path, finding.Message)
		}
	} else {
		for _, finding := range result.Findings {
			fmt.Fprintf(writer, "%-7s %-7s %-36s %-24s %-32s %s\n", finding.Severity, fixabilityLabel(finding.Fixable), finding.RuleID, finding.Path, finding.File, finding.Message)
		}
	}

	_, err := fmt.Fprintf(writer, "\nFound %d finding(s) in %d dashboard(s): %d error(s), %d warning(s), %d info(s).\n",
		result.Summary.Findings,
		result.Summary.Files,
		result.Summary.Errors,
		result.Summary.Warnings,
		result.Summary.Infos,
	)
	return err
}

func fixabilityLabel(fixable bool) string {
	if fixable {
		return "fixable"
	}
	return "manual"
}
