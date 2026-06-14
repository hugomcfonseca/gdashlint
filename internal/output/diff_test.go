package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestUnifiedDiffNoColor(t *testing.T) {
	before := []byte(`{"editable":true}` + "\n")
	after := []byte(`{"editable":false}` + "\n")

	var buf bytes.Buffer
	if err := UnifiedDiff(&buf, "a.json", "a.json", before, after, false); err != nil {
		t.Fatalf("UnifiedDiff returned error: %v", err)
	}

	out := buf.String()
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("expected no ANSI escapes in non-color mode, got %q", out)
	}
	if !strings.Contains(out, "--- a.json") || !strings.Contains(out, "+++ a.json") {
		t.Fatalf("expected file header, got %q", out)
	}
	if !strings.Contains(out, `-{"editable":true}`) || !strings.Contains(out, `+{"editable":false}`) {
		t.Fatalf("expected diff lines, got %q", out)
	}
}

func TestUnifiedDiffColor(t *testing.T) {
	before := []byte(`{"editable":true}` + "\n")
	after := []byte(`{"editable":false}` + "\n")

	var buf bytes.Buffer
	if err := UnifiedDiff(&buf, "a.json", "a.json", before, after, true); err != nil {
		t.Fatalf("UnifiedDiff returned error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, ansiRed) {
		t.Fatalf("expected red ANSI code for removed line, got %q", out)
	}
	if !strings.Contains(out, ansiGreen) {
		t.Fatalf("expected green ANSI code for added line, got %q", out)
	}
	if !strings.Contains(out, ansiCyan) {
		t.Fatalf("expected cyan ANSI code for hunk header, got %q", out)
	}
	if !strings.Contains(out, ansiReset) {
		t.Fatalf("expected ANSI reset after colored lines, got %q", out)
	}
	// Semantic content must still be present despite escapes.
	if !strings.Contains(out, `{"editable":true}`) || !strings.Contains(out, `{"editable":false}`) {
		t.Fatalf("expected diff content inside color codes, got %q", out)
	}
}

func TestUnifiedDiffIdentical(t *testing.T) {
	data := []byte(`{"editable":false}` + "\n")

	var buf bytes.Buffer
	if err := UnifiedDiff(&buf, "a.json", "a.json", data, data, true); err != nil {
		t.Fatalf("UnifiedDiff returned error: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected empty output for identical files, got %q", buf.String())
	}
}

func TestIsColorWriterBufferReturnsFalse(t *testing.T) {
	var buf bytes.Buffer
	if IsColorWriter(&buf) {
		t.Fatal("expected IsColorWriter to return false for bytes.Buffer")
	}
}
