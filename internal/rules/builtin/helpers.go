package builtin

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hugomcfonseca/gdashlint/internal/dashboard"
)

type panelRef struct {
	object map[string]any
	path   string
}

func sourceFile(source dashboard.Source) string {
	if source.Path != "" {
		return source.Path
	}
	return source.Name
}

func dashboardObject(root any) (map[string]any, bool) {
	object, ok := root.(map[string]any)
	return object, ok
}

func collectPanels(root any) []panelRef {
	object, ok := dashboardObject(root)
	if !ok {
		return nil
	}
	panels, ok := object["panels"].([]any)
	if !ok {
		return nil
	}
	return collectPanelArray(panels, "$.panels")
}

func collectPanelArray(panels []any, basePath string) []panelRef {
	refs := make([]panelRef, 0, len(panels))
	for index, panel := range panels {
		object, ok := panel.(map[string]any)
		if !ok {
			continue
		}
		path := fmt.Sprintf("%s[%d]", basePath, index)
		refs = append(refs, panelRef{object: object, path: path})
		if nested, ok := object["panels"].([]any); ok {
			refs = append(refs, collectPanelArray(nested, path+".panels")...)
		}
	}
	return refs
}

func nonEmptyString(value any) bool {
	text, ok := value.(string)
	return ok && strings.TrimSpace(text) != ""
}

func numberValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case json.Number:
		number, err := typed.Float64()
		return number, err == nil
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	case string:
		number, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return number, err == nil
	default:
		return 0, false
	}
}

func isEmptyValue(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(typed) == ""
	case []any:
		if len(typed) == 0 {
			return true
		}
		for _, item := range typed {
			if !isEmptyValue(item) {
				return false
			}
		}
		return true
	case map[string]any:
		if len(typed) == 0 {
			return true
		}
		for key, item := range typed {
			if key == "selected" {
				selected, ok := item.(bool)
				if ok && !selected {
					continue
				}
			}
			if !isEmptyValue(item) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
