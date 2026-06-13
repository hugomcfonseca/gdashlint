package config

import (
	"fmt"
	"reflect"
)

// ValidateRuleOptions validates that provided options conform to expected types.
// This is called before passing options to WithOptions() implementations.
func ValidateRuleOptions(ruleID string, options map[string]any, schema map[string]string) error {
	if len(options) == 0 {
		return nil
	}

	if len(schema) == 0 {
		// If no schema provided, all options are rejected
		return fmt.Errorf("rule %q does not support options", ruleID)
	}

	for key, value := range options {
		expectedType, ok := schema[key]
		if !ok {
			return fmt.Errorf("rule %s: unsupported option %q", ruleID, key)
		}

		if err := validateOptionType(ruleID, key, expectedType, value); err != nil {
			return err
		}
	}

	return nil
}

// validateOptionType checks if a value matches the expected type.
func validateOptionType(ruleID, optionName, expectedType string, value any) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("rule %s: option %q must be a string, got %s", ruleID, optionName, reflect.TypeOf(value).String())
		}
	case "int":
		// YAML unmarshals to float64 by default for numbers
		if _, ok := value.(float64); !ok {
			if _, ok := value.(int); !ok {
				return fmt.Errorf("rule %s: option %q must be a number, got %s", ruleID, optionName, reflect.TypeOf(value).String())
			}
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("rule %s: option %q must be a boolean, got %s", ruleID, optionName, reflect.TypeOf(value).String())
		}
	case "[]string":
		list, ok := value.([]any)
		if !ok {
			return fmt.Errorf("rule %s: option %q must be a list, got %s", ruleID, optionName, reflect.TypeOf(value).String())
		}
		for i, item := range list {
			if _, ok := item.(string); !ok {
				return fmt.Errorf("rule %s: option %q[%d] must be a string, got %s", ruleID, optionName, i, reflect.TypeOf(item).String())
			}
		}
	case "[]any":
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("rule %s: option %q must be a list, got %s", ruleID, optionName, reflect.TypeOf(value).String())
		}
	default:
		return fmt.Errorf("rule %s: unknown option type %q for %q", ruleID, expectedType, optionName)
	}

	return nil
}
