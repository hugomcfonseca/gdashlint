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

func TestRunFixInPlace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dashboard.json")
	writeDashboard(t, path, `{"title":"Example","uid":"example","tags":["team"],"editable":true,"refresh":"30s","templating":{"list":[{"current":{"text":"prod","value":"prod"}}]},"panels":[]}`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"fix", path, "--yes", "--fail-on", "none"}, strings.NewReader(""), &stdout, &stderr, BuildInfo{})

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr %q", code, stderr.String())
	}
	contents := readDashboard(t, path)
	for _, want := range []string{`"editable":false`, `"refresh":"1m"`, `"current":{}`} {
		if !strings.Contains(contents, want) {
			t.Fatalf("expected fixed dashboard to contain %s, got %s", want, contents)
		}
	}
	if !strings.Contains(stderr.String(), "Applied 3 fix(es) across 1 dashboard(s).") {
		t.Fatalf("expected fix summary, got %q", stderr.String())
	}
	if strings.Contains(stderr.String(), "Would apply") {
		t.Fatalf("expected no 'Would apply' in non-dry-run mode, got stderr %q", stderr.String())
	}
	if strings.Contains(stdout.String(), "simulated") {
		t.Fatalf("expected no 'simulated' label in non-dry-run mode, got stdout %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "After fixes:") {
		t.Fatalf("expected post-fix summary in stdout, got %q", stdout.String())
	}
}

func TestRunFixCopyMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dashboard.json")
	writeDashboard(t, path, `{"title":"Example","uid":"example","tags":["team"],"editable":true,"refresh":"1m","panels":[]}`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"fix", path, "--mode", "copy", "--yes", "--fail-on", "none"}, strings.NewReader(""), &stdout, &stderr, BuildInfo{})

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr %q", code, stderr.String())
	}
	original := readDashboard(t, path)
	if !strings.Contains(original, `"editable":true`) {
		t.Fatalf("expected original dashboard to be unchanged, got %s", original)
	}
	fixed := readDashboard(t, filepath.Join(dir, "dashboard.fixed.json"))
	if !strings.Contains(fixed, `"editable":false`) {
		t.Fatalf("expected copied dashboard to be fixed, got %s", fixed)
	}
}

func TestRunFixRejectsSymlinkInput(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.json")
	link := filepath.Join(dir, "dashboard.json")
	writeDashboard(t, target, `{"title":"Example","uid":"example","tags":["team"],"editable":true,"refresh":"1m","panels":[]}`)
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"fix", link, "--yes", "--fail-on", "none"}, strings.NewReader(""), &stdout, &stderr, BuildInfo{})

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "refusing to write through symlink") {
		t.Fatalf("expected symlink refusal, got %q", stderr.String())
	}
}

func TestRunFixCopyModeRejectsSymlinkDestination(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dashboard.json")
	target := filepath.Join(dir, "target.json")
	copyPath := filepath.Join(dir, "dashboard.fixed.json")
	writeDashboard(t, path, `{"title":"Example","uid":"example","tags":["team"],"editable":true,"refresh":"1m","panels":[]}`)
	writeDashboard(t, target, `{"title":"Other","uid":"other","tags":["team"],"editable":true,"refresh":"1m","panels":[]}`)
	if err := os.Symlink(target, copyPath); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"fix", path, "--mode", "copy", "--yes", "--fail-on", "none"}, strings.NewReader(""), &stdout, &stderr, BuildInfo{})

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "refusing to write through symlink") {
		t.Fatalf("expected symlink refusal, got %q", stderr.String())
	}
}

func TestRunFixCopyModeRejectsPathSuffix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dashboard.json")
	writeDashboard(t, path, `{"title":"Example","uid":"example","tags":["team"],"editable":true,"refresh":"1m","panels":[]}`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"fix", path, "--mode", "copy", "--suffix", "../escaped", "--fail-on", "none"}, strings.NewReader(""), &stdout, &stderr, BuildInfo{})

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "must be a filename suffix") {
		t.Fatalf("expected invalid suffix error, got %q", stderr.String())
	}
}

func TestRunFixCopyModeCreatesPrivateFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dashboard.json")
	writeDashboard(t, path, `{"title":"Example","uid":"example","tags":["team"],"editable":true,"refresh":"1m","panels":[]}`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"fix", path, "--mode", "copy", "--yes", "--fail-on", "none"}, strings.NewReader(""), &stdout, &stderr, BuildInfo{})

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr %q", code, stderr.String())
	}
	info, err := os.Stat(filepath.Join(dir, "dashboard.fixed.json"))
	if err != nil {
		t.Fatalf("stat fixed dashboard: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("expected copied dashboard mode 0600, got %o", got)
	}
}

func TestRunFixDryRunDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dashboard.json")
	writeDashboard(t, path, `{"title":"Example","uid":"example","tags":["team"],"editable":true,"refresh":"1m","panels":[]}`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"fix", path, "--dry-run", "--fail-on", "none"}, strings.NewReader(""), &stdout, &stderr, BuildInfo{})

	if code != 1 {
		t.Fatalf("expected exit code 1 when dry-run would apply fixes, got %d, stderr %q", code, stderr.String())
	}
	contents := readDashboard(t, path)
	if !strings.Contains(contents, `"editable":true`) {
		t.Fatalf("expected dry-run to leave original dashboard unchanged, got %s", contents)
	}
	if !strings.Contains(stderr.String(), "Would apply 1 fix(es) across 1 dashboard(s).") {
		t.Fatalf("expected dry-run summary, got %q", stderr.String())
	}
	if !strings.Contains(stdout.String(), "--- "+path) || !strings.Contains(stdout.String(), `+{"title":"Example","uid":"example","tags":["team"],"editable":false`) {
		t.Fatalf("expected dry-run diff, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "After simulated fixes: 0 finding(s) remain") {
		t.Fatalf("expected post-fix summary, got %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "No findings") {
		t.Fatalf("expected concise post-fix summary instead of full lint output, got %q", stdout.String())
	}
}

func TestRunFixPromptsForApproval(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dashboard.json")
	writeDashboard(t, path, `{"title":"Example","uid":"example","tags":["team"],"editable":true,"refresh":"1m","panels":[]}`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"fix", path, "--fail-on", "none"}, strings.NewReader("yes\n"), &stdout, &stderr, BuildInfo{})

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr %q", code, stderr.String())
	}
	contents := readDashboard(t, path)
	if !strings.Contains(contents, `"editable":false`) {
		t.Fatalf("expected approved fix to write dashboard, got %s", contents)
	}
	if !strings.Contains(stderr.String(), "Apply fixes? [y/N]") {
		t.Fatalf("expected approval prompt, got %q", stderr.String())
	}
}

func TestRunFixDeclineDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dashboard.json")
	writeDashboard(t, path, `{"title":"Example","uid":"example","tags":["team"],"editable":true,"refresh":"1m","panels":[]}`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"fix", path, "--fail-on", "none"}, strings.NewReader("no\n"), &stdout, &stderr, BuildInfo{})

	if code != 1 {
		t.Fatalf("expected exit code 1 for declined fixes, got %d, stderr %q", code, stderr.String())
	}
	contents := readDashboard(t, path)
	if !strings.Contains(contents, `"editable":true`) {
		t.Fatalf("expected declined fix to leave dashboard unchanged, got %s", contents)
	}
	if !strings.Contains(stdout.String(), "--- "+path) {
		t.Fatalf("expected diff before decline, got %q", stdout.String())
	}
}

func TestRunFixDryRunJSONIncludesFixes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dashboard.json")
	writeDashboard(t, path, `{"title":"Example","uid":"example","tags":["team"],"editable":true,"refresh":"1m","panels":[]}`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"fix", path, "--dry-run", "--format", "json", "--fail-on", "none"}, strings.NewReader(""), &stdout, &stderr, BuildInfo{})

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

func TestRunFixAppliesCustomRuleFix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dashboard.json")
	configPath := filepath.Join(dir, "gdashlint.yaml")
	writeDashboard(t, path, `{"title":"Example","uid":"example","tags":["team"],"editable":false,"refresh":"1m","panels":[]}`)
	writeDashboard(t, configPath, `version: 1
failOn: none
customRules:
  custom.timezone-required:
    type: required
    path: $.timezone
    message: dashboard timezone is required
    fix:
      action: setDefault
      value: browser
`)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"fix", path, "--config", configPath, "--yes", "--fail-on", "none"}, strings.NewReader(""), &stdout, &stderr, BuildInfo{})

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr %q", code, stderr.String())
	}
	contents := readDashboard(t, path)
	if !strings.Contains(contents, `"timezone":"browser"`) {
		t.Fatalf("expected custom fix to set timezone, got %s", contents)
	}
	if !strings.Contains(stderr.String(), "Applied 1 fix(es) across 1 dashboard(s).") {
		t.Fatalf("expected custom fix summary, got %q", stderr.String())
	}
}

func TestRunFixRejectsStdin(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run([]string{"fix", "-"}, strings.NewReader(`{"title":"Example"}`), &stdout, &stderr, BuildInfo{})

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "fix does not support stdin") {
		t.Fatalf("expected stdin fix error, got %q", stderr.String())
	}
}

func TestRunLintRejectsFixFlag(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"lint", "-", "--fix"}, strings.NewReader(`{"title":"Example"}`), &stdout, &stderr, BuildInfo{})

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "unknown flag: --fix") {
		t.Fatalf("expected unknown --fix flag error, got %q", stderr.String())
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
