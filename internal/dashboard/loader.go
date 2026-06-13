package dashboard

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Loader loads dashboard JSON documents from files, directories, and stdin.
type Loader struct {
	Stdin io.Reader
}

// Load loads dashboards from paths. A path of - reads one dashboard from stdin.
func (l Loader) Load(paths []string) ([]Dashboard, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("at least one dashboard path or - is required")
	}

	var dashboards []Dashboard
	for _, path := range paths {
		if path == "-" {
			dashboard, err := parse("<stdin>", "", true, l.Stdin)
			if err != nil {
				return nil, err
			}
			dashboards = append(dashboards, dashboard)
			continue
		}

		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		if info.IsDir() {
			loaded, err := loadDirectory(path)
			if err != nil {
				return nil, err
			}
			dashboards = append(dashboards, loaded...)
			continue
		}

		dashboard, err := loadFile(path)
		if err != nil {
			return nil, err
		}
		dashboards = append(dashboards, dashboard)
	}
	return dashboards, nil
}

func loadDirectory(root string) ([]Dashboard, error) {
	var files []string
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".json") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}
	sort.Strings(files)

	dashboards := make([]Dashboard, 0, len(files))
	for _, file := range files {
		dashboard, err := loadFile(file)
		if err != nil {
			return nil, err
		}
		dashboards = append(dashboards, dashboard)
	}
	return dashboards, nil
}

func loadFile(path string) (Dashboard, error) {
	// #nosec G304 -- gdashlint is a local CLI and intentionally reads user-provided dashboard paths.
	file, err := os.Open(path)
	if err != nil {
		return Dashboard{}, fmt.Errorf("open %s: %w", path, err)
	}
	dashboard, parseErr := parse(filepath.Base(path), path, false, file)
	if closeErr := file.Close(); closeErr != nil {
		return Dashboard{}, fmt.Errorf("close %s: %w", path, closeErr)
	}
	if parseErr != nil {
		return Dashboard{}, parseErr
	}
	return dashboard, nil
}

func parse(name string, path string, stdin bool, reader io.Reader) (Dashboard, error) {
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()

	var root any
	if err := decoder.Decode(&root); err != nil {
		return Dashboard{}, fmt.Errorf("parse %s: %w", name, err)
	}
	if _, ok := root.(map[string]any); !ok {
		return Dashboard{}, fmt.Errorf("parse %s: dashboard JSON must be an object", name)
	}
	if decoder.Decode(new(any)) != io.EOF {
		return Dashboard{}, fmt.Errorf("parse %s: dashboard input must contain exactly one JSON document", name)
	}
	return Dashboard{Source: Source{Name: name, Path: path, Stdin: stdin}, Root: root}, nil
}
