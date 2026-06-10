package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"--version"}, strings.NewReader(""), &stdout, &stderr, BuildInfo{
		Version: "v0.0.0-test",
		Commit:  "abc123",
		Date:    "2026-06-10",
	})

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	want := "gdashlint v0.0.0-test (abc123, 2026-06-10)\n"
	if stdout.String() != want {
		t.Fatalf("unexpected stdout: got %q want %q", stdout.String(), want)
	}

	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr, got %q", stderr.String())
	}
}

func TestRunLintFromStdin(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"lint", "-", "--fail-on", "none"}, strings.NewReader(`{"title":"Example","uid":"example","tags":["team"],"editable":false,"refresh":"1m","panels":[]}`), &stdout, &stderr, BuildInfo{})

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "No findings") {
		t.Fatalf("expected no findings output, got %q", stdout.String())
	}
}

func TestRunLintFixInPlace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dashboard.json")
	writeDashboard(t, path, `{"title":"Example","uid":"example","tags":["team"],"editable":true,"refresh":"30s","templating":{"list":[{"current":{"text":"prod","value":"prod"}}]},"panels":[]}`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"lint", path, "--fix", "--fail-on", "none"}, strings.NewReader(""), &stdout, &stderr, BuildInfo{})

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr %q", code, stderr.String())
	}
	contents := readDashboard(t, path)
	for _, want := range []string{`"editable": false`, `"refresh": "1m"`, `"current": {}`} {
		if !strings.Contains(contents, want) {
			t.Fatalf("expected fixed dashboard to contain %s, got %s", want, contents)
		}
	}
	if !strings.Contains(stderr.String(), "Applied 3 fix(es).") {
		t.Fatalf("expected fix summary, got %q", stderr.String())
	}
}

func TestRunLintFixCopyMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dashboard.json")
	writeDashboard(t, path, `{"title":"Example","uid":"example","tags":["team"],"editable":true,"refresh":"1m","panels":[]}`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"lint", path, "--fix", "--fix-mode", "copy", "--fail-on", "none"}, strings.NewReader(""), &stdout, &stderr, BuildInfo{})

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr %q", code, stderr.String())
	}
	original := readDashboard(t, path)
	if !strings.Contains(original, `"editable":true`) {
		t.Fatalf("expected original dashboard to be unchanged, got %s", original)
	}
	fixed := readDashboard(t, filepath.Join(dir, "dashboard.fixed.json"))
	if !strings.Contains(fixed, `"editable": false`) {
		t.Fatalf("expected copied dashboard to be fixed, got %s", fixed)
	}
}

func TestRunLintFixDryRunDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dashboard.json")
	writeDashboard(t, path, `{"title":"Example","uid":"example","tags":["team"],"editable":true,"refresh":"1m","panels":[]}`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"lint", path, "--fix", "--dry-run", "--fail-on", "none"}, strings.NewReader(""), &stdout, &stderr, BuildInfo{})

	if code != 1 {
		t.Fatalf("expected exit code 1 when dry-run would apply fixes, got %d, stderr %q", code, stderr.String())
	}
	contents := readDashboard(t, path)
	if !strings.Contains(contents, `"editable":true`) {
		t.Fatalf("expected dry-run to leave original dashboard unchanged, got %s", contents)
	}
	if !strings.Contains(stderr.String(), "Would apply 1 fix(es).") {
		t.Fatalf("expected dry-run summary, got %q", stderr.String())
	}
	if !strings.Contains(stdout.String(), "No findings") {
		t.Fatalf("expected post-fix simulated findings, got %q", stdout.String())
	}
}

func TestRunLintFixDryRunJSONIncludesFixes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dashboard.json")
	writeDashboard(t, path, `{"title":"Example","uid":"example","tags":["team"],"editable":true,"refresh":"1m","panels":[]}`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"lint", path, "--fix", "--dry-run", "--format", "json", "--fail-on", "none"}, strings.NewReader(""), &stdout, &stderr, BuildInfo{})

	if code != 1 {
		t.Fatalf("expected exit code 1 when dry-run would apply fixes, got %d", code)
	}
	if stderr.Len() != 0 {
		t.Fatalf("expected empty stderr for json fix output, got %q", stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, `"fixes"`) || !strings.Contains(output, `"rule_id": "core.dashboard-not-editable"`) {
		t.Fatalf("expected json fixes in stdout, got %q", output)
	}
}

func TestRunLintFixRejectsStdin(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"lint", "-", "--fix"}, strings.NewReader(`{"title":"Example"}`), &stdout, &stderr, BuildInfo{})

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "--fix does not support stdin") {
		t.Fatalf("expected stdin fix error, got %q", stderr.String())
	}
}

func TestRunLintMissingTitleFails(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"lint", "-"}, strings.NewReader(`{"panels":[]}`), &stdout, &stderr, BuildInfo{})

	if code != 1 {
		t.Fatalf("expected exit code 1, got %d, stderr %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "core.dashboard-title-required") {
		t.Fatalf("expected title rule finding, got %q", stdout.String())
	}
}

func writeDashboard(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write dashboard: %v", err)
	}
}

func readDashboard(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read dashboard: %v", err)
	}
	return string(contents)
}
