package output

import (
	"encoding/json"
	"io"

	"github.com/hugomcfonseca/gdashlint/internal/lint"
	"github.com/hugomcfonseca/gdashlint/internal/rule"
)

// JSON writes machine-readable JSON output.
func JSON(writer io.Writer, result lint.Result) error {
	if result.Findings == nil {
		result.Findings = []rule.Finding{}
	}
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
