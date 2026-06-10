package config

import (
	"os"
	"path/filepath"
)

var configNames = []string{".gdashlint.yaml", ".gdashlint.yml", "gdashlint.yaml", "gdashlint.yml"}

// Discover searches from start upward for a config file.
func Discover(start string) (string, bool, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", false, err
	}
	info, err := os.Stat(current)
	if err != nil {
		return "", false, err
	}
	if !info.IsDir() {
		current = filepath.Dir(current)
	}

	for {
		for _, name := range configNames {
			candidate := filepath.Join(current, name)
			info, err := os.Stat(candidate)
			if err == nil && !info.IsDir() {
				return candidate, true, nil
			}
			if err != nil && !os.IsNotExist(err) {
				return "", false, err
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", false, nil
		}
		current = parent
	}
}

// LoadDiscovered loads explicitPath or auto-discovers config from cwd.
func LoadDiscovered(explicitPath string) (Config, error) {
	if explicitPath != "" {
		return Load(explicitPath)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return Config{}, err
	}
	path, ok, err := Discover(cwd)
	if err != nil {
		return Config{}, err
	}
	if !ok {
		return Default(), nil
	}
	return Load(path)
}
