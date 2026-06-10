package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/hugomcfonseca/gdashlint/internal/lint"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

// GitHub writes GitHub Actions workflow command annotations.
func GitHub(writer io.Writer, result lint.Result) error {
	for _, finding := range result.Findings {
		level := githubLevel(finding.Severity)
		if _, err := fmt.Fprintf(
			writer,
			"::%s file=%s,title=%s::%s\n",
			level,
			escapeProperty(finding.File),
			escapeProperty(finding.RuleID),
			escapeData(fmt.Sprintf("%s (%s)", finding.Message, finding.Path)),
		); err != nil {
			return err
		}
	}
	return nil
}

func githubLevel(severity rule.Severity) string {
	switch severity {
	case rule.SeverityError:
		return "error"
	case rule.SeverityWarning:
		return "warning"
	default:
		return "notice"
	}
}

func escapeProperty(value string) string {
	value = strings.ReplaceAll(value, "%", "%25")
	value = strings.ReplaceAll(value, "\r", "%0D")
	value = strings.ReplaceAll(value, "\n", "%0A")
	value = strings.ReplaceAll(value, ":", "%3A")
	value = strings.ReplaceAll(value, ",", "%2C")
	return value
}

func escapeData(value string) string {
	value = strings.ReplaceAll(value, "%", "%25")
	value = strings.ReplaceAll(value, "\r", "%0D")
	value = strings.ReplaceAll(value, "\n", "%0A")
	return value
}
