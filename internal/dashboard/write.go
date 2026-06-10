package dashboard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteFile writes a dashboard JSON document with stable indentation.
func WriteFile(path string, dash Dashboard) error {
	data, err := json.MarshalIndent(dash.Root, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// CopyPath returns the sibling copy path for a fixed dashboard.
func CopyPath(path string, suffix string) string {
	if suffix == "" {
		suffix = ".fixed"
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	return base + suffix + ext
}
