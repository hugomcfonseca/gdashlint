package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hugomcfonseca/gdashlint/internal/lint"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

func TestGitHubAnnotations(t *testing.T) {
	var buffer bytes.Buffer
	result := lint.Result{Findings: []rule.Finding{{
		RuleID:   "core.dashboard-title-required",
		Severity: rule.SeverityError,
		Message:  "dashboard title is required",
		File:     "dashboards/api.json",
		Path:     "$.title",
	}}}

	if err := GitHub(&buffer, result); err != nil {
		t.Fatalf("GitHub returned error: %v", err)
	}
	output := buffer.String()
	if !strings.Contains(output, "::error file=dashboards/api.json,title=core.dashboard-title-required::dashboard title is required ($.title)") {
		t.Fatalf("unexpected annotation output: %q", output)
	}
}
