package dashboard

import (
	"fmt"
	"strconv"
	"strings"
)

type pathToken struct {
	key      string
	index    int
	wildcard bool
	isIndex  bool
}

// ValidatePath validates a supported JSONPath-like path without evaluating it.
func ValidatePath(path string) error {
	_, err := parsePath(path)
	return err
}

// Exists reports whether at least one value exists at path.
func Exists(root any, path string) (bool, error) {
	values, err := Values(root, path)
	if err != nil {
		return false, err
	}
	return len(values) > 0, nil
}

// Values returns all values matching a simple JSONPath-like path.
// Supported syntax: $.field, $.field.nested, $.array[0], $.array[*].field.
func Values(root any, path string) ([]any, error) {
	tokens, err := parsePath(path)
	if err != nil {
		return nil, err
	}
	current := []any{root}
	for _, token := range tokens {
		next := make([]any, 0)
		for _, value := range current {
			matches := applyToken(value, token)
			next = append(next, matches...)
		}
		current = next
		if len(current) == 0 {
			return nil, nil
		}
	}
	return current, nil
}

func parsePath(path string) ([]pathToken, error) {
	if path == "$" {
		return nil, nil
	}
	if !strings.HasPrefix(path, "$.") {
		return nil, fmt.Errorf("path %q must start with $", path)
	}

	parts := strings.Split(strings.TrimPrefix(path, "$."), ".")
	tokens := make([]pathToken, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("invalid empty path segment in %q", path)
		}
		segmentTokens, err := parseSegment(part)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, segmentTokens...)
	}
	return tokens, nil
}

func parseSegment(segment string) ([]pathToken, error) {
	var tokens []pathToken
	for segment != "" {
		bracket := strings.Index(segment, "[")
		if bracket == -1 {
			tokens = append(tokens, pathToken{key: segment})
			return tokens, nil
		}
		if bracket > 0 {
			tokens = append(tokens, pathToken{key: segment[:bracket]})
		}
		end := strings.Index(segment[bracket:], "]")
		if end == -1 {
			return nil, fmt.Errorf("invalid array selector in %q", segment)
		}
		selector := segment[bracket+1 : bracket+end]
		if selector == "*" {
			tokens = append(tokens, pathToken{wildcard: true, isIndex: true})
		} else {
			index, err := strconv.Atoi(selector)
			if err != nil || index < 0 {
				return nil, fmt.Errorf("invalid array index %q", selector)
			}
			tokens = append(tokens, pathToken{index: index, isIndex: true})
		}
		segment = segment[bracket+end+1:]
		if segment != "" && !strings.HasPrefix(segment, "[") {
			return nil, fmt.Errorf("invalid path segment %q", segment)
		}
	}
	return tokens, nil
}

func applyToken(value any, token pathToken) []any {
	if token.isIndex {
		items, ok := value.([]any)
		if !ok {
			return nil
		}
		if token.wildcard {
			return items
		}
		if token.index >= len(items) {
			return nil
		}
		return []any{items[token.index]}
	}

	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	child, ok := object[token.key]
	if !ok {
		return nil
	}
	return []any{child}
}
