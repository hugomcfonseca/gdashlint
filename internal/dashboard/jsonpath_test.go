package dashboard

import "testing"

func TestValuesWildcard(t *testing.T) {
	root := map[string]any{
		"panels": []any{
			map[string]any{"title": "API"},
			map[string]any{"title": "DB"},
		},
	}

	values, err := Values(root, "$.panels[*].title")
	if err != nil {
		t.Fatalf("Values returned error: %v", err)
	}
	if len(values) != 2 || values[0] != "API" || values[1] != "DB" {
		t.Fatalf("unexpected values: %#v", values)
	}
}

func TestExistsMissingPath(t *testing.T) {
	root := map[string]any{"title": "API"}

	exists, err := Exists(root, "$.tags")
	if err != nil {
		t.Fatalf("Exists returned error: %v", err)
	}
	if exists {
		t.Fatalf("expected missing path")
	}
}

func TestCompiledPathValues(t *testing.T) {
	root := map[string]any{
		"panels": []any{
			map[string]any{"title": "API"},
			map[string]any{"title": "DB"},
		},
	}
	compiled, err := CompilePath("$.panels[*].title")
	if err != nil {
		t.Fatalf("CompilePath returned error: %v", err)
	}

	values := compiled.Values(root)
	if len(values) != 2 || values[0] != "API" || values[1] != "DB" {
		t.Fatalf("unexpected values: %#v", values)
	}
	if !compiled.Exists(root) {
		t.Fatalf("expected compiled path to exist")
	}
}
