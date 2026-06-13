package output

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

// Rules writes rule metadata in the requested format.
func Rules(writer io.Writer, metadata []rule.Metadata, format string) error {
	if format == "json" {
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(metadata)
	}
	return RulesText(writer, metadata)
}

// RulesText writes human-readable rule metadata.
func RulesText(writer io.Writer, metadata []rule.Metadata) error {
	for _, meta := range metadata {
		fixable := "-"
		if meta.Fixable {
			fixable = "fixable"
		}
		if _, err := fmt.Fprintf(writer, "%-36s %-7s %-8s %-7s %s\n", meta.ID, meta.Severity, meta.Source, fixable, meta.Description); err != nil {
			return err
		}
	}
	return nil
}
