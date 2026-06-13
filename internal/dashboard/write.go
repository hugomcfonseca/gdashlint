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
	if err := rejectSymlink(path); err != nil {
		return err
	}
	data, err := json.MarshalIndent(dash.Root, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func rejectSymlink(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to write through symlink %s", path)
	}
	return nil
}

// ValidateCopySuffix validates a suffix used to create sibling fixed files.
func ValidateCopySuffix(suffix string) error {
	if suffix == "" {
		return nil
	}
	if filepath.IsAbs(suffix) || filepath.Base(suffix) != suffix || strings.ContainsAny(suffix, `/\\`) {
		return fmt.Errorf("copy suffix %q must be a filename suffix, not a path", suffix)
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
