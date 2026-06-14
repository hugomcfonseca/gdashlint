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

func TestSetCreatesObjectParents(t *testing.T) {
	root := map[string]any{}
	updated, err := Set(root, "$.templating.current.value", "prod")
	if err != nil {
		t.Fatalf("Set returned error: %v", err)
	}
	values, err := Values(updated, "$.templating.current.value")
	if err != nil {
		t.Fatalf("Values returned error: %v", err)
	}
	if len(values) != 1 || values[0] != "prod" {
		t.Fatalf("unexpected value: %#v", values)
	}
}

func TestSetDefaultDoesNotOverwriteExistingValue(t *testing.T) {
	root := map[string]any{"refresh": "5m"}
	updated, changed, err := SetDefault(root, "$.refresh", "1m")
	if err != nil {
		t.Fatalf("SetDefault returned error: %v", err)
	}
	if changed {
		t.Fatalf("expected existing value to be unchanged")
	}
	if updated.(map[string]any)["refresh"] != "5m" {
		t.Fatalf("expected refresh to remain 5m, got %#v", updated)
	}
}

func TestValidateWritablePathRejectsWildcard(t *testing.T) {
	if err := ValidateWritablePath("$.panels[*].title"); err == nil {
		t.Fatalf("expected wildcard write path to be rejected")
	}
}
