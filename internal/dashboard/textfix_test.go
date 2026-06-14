package dashboard

import (
	"strings"
	"testing"
)

func TestApplyTextEditsPreservesSiblingKeyOrder(t *testing.T) {
	input := []byte(`{
  "title": "Example",
  "uid": "example",
  "tags": ["team"],
  "editable": true,
  "refresh": "30s",
  "templating": {
    "list": [
      {
        "name": "env",
        "current": {
          "text": "prod",
          "value": "prod"
        }
      }
    ]
  },
  "panels": []
}
`)

	updated, err := ApplyTextEdits(input, []TextEdit{
		{Path: "$.editable", Value: false},
		{Path: "$.refresh", Value: "1m"},
		{Path: "$.templating.list[0].current", Value: map[string]any{}},
	})
	if err != nil {
		t.Fatalf("ApplyTextEdits returned error: %v", err)
	}
	text := string(updated)
	ordered := []string{
		`"title": "Example"`,
		`"uid": "example"`,
		`"tags": ["team"]`,
		`"editable": false`,
		`"refresh": "1m"`,
		`"templating"`,
		`"panels": []`,
	}
	last := -1
	for _, fragment := range ordered {
		index := strings.Index(text, fragment)
		if index == -1 {
			t.Fatalf("expected output to contain %s, got %s", fragment, text)
		}
		if index < last {
			t.Fatalf("expected %s to appear after previous fragment, got %s", fragment, text)
		}
		last = index
	}
	if !strings.Contains(text, `"current": {}`) {
		t.Fatalf("expected current to be cleared, got %s", text)
	}
}

func TestSJSONPathConvertsArraySelectors(t *testing.T) {
	path, err := SJSONPath("$.templating.list[0].current.value")
	if err != nil {
		t.Fatalf("SJSONPath returned error: %v", err)
	}
	if path != "templating.list.0.current.value" {
		t.Fatalf("unexpected path: %q", path)
	}
}
