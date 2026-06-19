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
			"::%s %s::%s\n",
			level,
			githubAnnotationProperties(finding),
			escapeData(fmt.Sprintf("%s (%s)", finding.Message, finding.Path)),
		); err != nil {
			return err
		}
	}
	return nil
}

func githubAnnotationProperties(finding rule.Finding) string {
	properties := []string{}
	if isGitHubAnnotatableFile(finding.File) {
		line := finding.Line
		if line < 1 {
			line = 1
		}
		properties = append(properties, "file="+escapeProperty(finding.File), fmt.Sprintf("line=%d", line))
	}
	properties = append(properties, "title="+escapeProperty(finding.RuleID))
	return strings.Join(properties, ",")
}

func isGitHubAnnotatableFile(file string) bool {
	return file != "" && file != "-" && file != "<stdin>"
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
