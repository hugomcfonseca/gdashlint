package dashboard

import "testing"

func TestLineForPathCommonDashboardPaths(t *testing.T) {
	raw := []byte(`{
  "title": "API Dashboard",
  "tags": ["api", "prod"],
  "refresh": "30s",
  "templating": {
    "list": [
      {
        "name": "env",
        "current": {"text": "prod", "value": "prod"}
      }
    ]
  },
  "panels": [
    {
      "title": "Requests",
      "gridPos": {"x": 0, "y": 0, "w": 12, "h": 8},
      "panels": [
        {
          "title": "Nested Requests",
          "gridPos": {"x": 0, "y": 8, "w": 12, "h": 8}
        }
      ]
    }
  ]
}`)

	tests := map[string]int{
		"$.title":                       2,
		"$.tags":                        3,
		"$.refresh":                     4,
		"$.templating.list[0].current":  9,
		"$.panels[0].title":             15,
		"$.panels[0].gridPos":           16,
		"$.panels[0].panels[0].title":   19,
		"$.panels[0].panels[0].gridPos": 20,
	}

	for path, want := range tests {
		got, ok := LineForPath(raw, path)
		if !ok {
			t.Fatalf("LineForPath(%q) did not find a line", path)
		}
		if got != want {
			t.Fatalf("LineForPath(%q) = %d, want %d", path, got, want)
		}
	}
}

func TestLineForPathMissingPath(t *testing.T) {
	if line, ok := LineForPath([]byte(`{"title":"API"}`), "$.refresh"); ok {
		t.Fatalf("LineForPath found unexpected missing path at line %d", line)
	}
}
