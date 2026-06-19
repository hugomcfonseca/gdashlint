package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hugomcfonseca/gdashlint/internal/lint"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

func TestGitHubAnnotationsIncludeLine(t *testing.T) {
	var buffer bytes.Buffer
	result := lint.Result{Findings: []rule.Finding{{
		RuleID:   "core.dashboard-title-required",
		Severity: rule.SeverityError,
		Message:  "dashboard title is required",
		File:     "dashboards/api.json",
		Path:     "$.title",
		Line:     163,
	}}}

	if err := GitHub(&buffer, result); err != nil {
		t.Fatalf("GitHub returned error: %v", err)
	}
	output := buffer.String()
	if !strings.Contains(output, "::error file=dashboards/api.json,line=163,title=core.dashboard-title-required::dashboard title is required ($.title)") {
		t.Fatalf("unexpected annotation output: %q", output)
	}
}

func TestGitHubAnnotationsFallbackToLineOne(t *testing.T) {
	var buffer bytes.Buffer
	result := lint.Result{Findings: []rule.Finding{{
		RuleID:   "core.unknown",
		Severity: rule.SeverityWarning,
		Message:  "unknown location",
		File:     "dashboards/api.json",
		Path:     "$.missing",
	}}}

	if err := GitHub(&buffer, result); err != nil {
		t.Fatalf("GitHub returned error: %v", err)
	}
	output := buffer.String()
	if !strings.Contains(output, "::warning file=dashboards/api.json,line=1,title=core.unknown::unknown location ($.missing)") {
		t.Fatalf("unexpected annotation output: %q", output)
	}
}

func TestGitHubAnnotationsSkipFileMetadataForStdin(t *testing.T) {
	var buffer bytes.Buffer
	result := lint.Result{Findings: []rule.Finding{{
		RuleID:   "core.dashboard-title-required",
		Severity: rule.SeverityError,
		Message:  "dashboard title is required",
		File:     "<stdin>",
		Path:     "$.title",
		Line:     1,
	}}}

	if err := GitHub(&buffer, result); err != nil {
		t.Fatalf("GitHub returned error: %v", err)
	}
	output := buffer.String()
	if strings.Contains(output, "file=") || strings.Contains(output, "line=") {
		t.Fatalf("stdin annotation should not include file or line metadata: %q", output)
	}
	if !strings.Contains(output, "::error title=core.dashboard-title-required::dashboard title is required ($.title)") {
		t.Fatalf("unexpected annotation output: %q", output)
	}
}

func TestGitHubAnnotationsEscapePropertiesAndData(t *testing.T) {
	var buffer bytes.Buffer
	result := lint.Result{Findings: []rule.Finding{{
		RuleID:   "custom:rule,one%two",
		Severity: rule.SeverityInfo,
		Message:  "message with %\nnewline",
		File:     "dashboards/team:api,prod%.json",
		Path:     "$.title",
		Line:     2,
	}}}

	if err := GitHub(&buffer, result); err != nil {
		t.Fatalf("GitHub returned error: %v", err)
	}
	output := buffer.String()
	if !strings.Contains(output, "file=dashboards/team%3Aapi%2Cprod%25.json,line=2,title=custom%3Arule%2Cone%25two") {
		t.Fatalf("annotation properties were not escaped: %q", output)
	}
	if !strings.Contains(output, "message with %25%0Anewline ($.title)") {
		t.Fatalf("annotation data was not escaped: %q", output)
	}
}
