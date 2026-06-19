package dashboard

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// LineForPath returns the 1-based source line for a supported JSONPath-like path.
// It reports false when the path cannot be mapped exactly in the provided JSON.
func LineForPath(raw []byte, path string) (int, bool) {
	if len(raw) == 0 {
		return 0, false
	}
	if _, err := CompilePath(path); err != nil {
		return 0, false
	}

	locations := make(map[string]int)
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := collectSourceLines(decoder, "$", newLineIndex(raw), locations); err != nil {
		return 0, false
	}
	line, ok := locations[path]
	if !ok || line < 1 {
		return 0, false
	}
	return line, true
}

func collectSourceLines(decoder *json.Decoder, path string, lines []int, locations map[string]int) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}

	delim, ok := token.(json.Delim)
	if !ok {
		if path != "$" {
			locations[path] = lineAtOffset(lines, decoder.InputOffset())
		}
		return nil
	}

	switch delim {
	case '{':
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return nil
			}
			childPath := joinObjectPath(path, key)
			locations[childPath] = lineAtOffset(lines, decoder.InputOffset())
			if err := collectSourceLines(decoder, childPath, lines, locations); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	case '[':
		for index := 0; decoder.More(); index++ {
			childPath := path + "[" + strconv.Itoa(index) + "]"
			if err := collectSourceLines(decoder, childPath, lines, locations); err != nil {
				return err
			}
		}
		_, err := decoder.Token()
		return err
	default:
		return nil
	}
}

func joinObjectPath(parent string, key string) string {
	if parent == "$" {
		return "$." + key
	}
	return parent + "." + key
}

func newLineIndex(raw []byte) []int {
	lines := []int{0}
	for index, b := range raw {
		if b == '\n' {
			lines = append(lines, index+1)
		}
	}
	return lines
}

func lineAtOffset(lines []int, offset int64) int {
	if offset < 0 {
		return 1
	}
	position := int(offset)
	low, high := 0, len(lines)
	for low < high {
		mid := low + (high-low)/2
		if lines[mid] <= position {
			low = mid + 1
		} else {
			high = mid
		}
	}
	if low < 1 {
		return 1
	}
	return low
}
