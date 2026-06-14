package dashboard

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// TextEdit describes one JSON value update that can be applied to dashboard source bytes.
type TextEdit struct {
	Path  string
	Value any
}

// ApplyTextEdits applies JSONPath-like value updates to source bytes while preserving
// unrelated object key order and formatting where possible.
func ApplyTextEdits(data []byte, edits []TextEdit) ([]byte, error) {
	updated := string(data)
	for _, edit := range edits {
		path, err := SJSONPath(edit.Path)
		if err != nil {
			return nil, err
		}
		updated, err = sjson.Set(updated, path, edit.Value)
		if err != nil {
			return nil, fmt.Errorf("set %s: %w", edit.Path, err)
		}
	}
	if !strings.HasSuffix(updated, "\n") {
		updated += "\n"
	}
	return []byte(updated), nil
}

// PathExistsInBytes reports whether a JSONPath-like path exists in source bytes.
func PathExistsInBytes(data []byte, path string) (bool, error) {
	sjsonPath, err := SJSONPath(path)
	if err != nil {
		return false, err
	}
	return gjson.GetBytes(data, sjsonPath).Exists(), nil
}

// SJSONPath converts gdashlint's writable JSONPath-like subset to sjson/gjson path syntax.
func SJSONPath(path string) (string, error) {
	compiled, err := CompilePath(path)
	if err != nil {
		return "", err
	}
	if err := compiled.validateWritable(); err != nil {
		return "", err
	}
	parts := make([]string, 0, len(compiled.tokens))
	for _, token := range compiled.tokens {
		if token.isIndex {
			parts = append(parts, strconv.Itoa(token.index))
			continue
		}
		parts = append(parts, escapeSJSONKey(token.key))
	}
	return strings.Join(parts, "."), nil
}

func escapeSJSONKey(key string) string {
	key = strings.ReplaceAll(key, `\\`, `\\\\`)
	key = strings.ReplaceAll(key, `.`, `\\.`)
	key = strings.ReplaceAll(key, `:`, `\\:`)
	return key
}
