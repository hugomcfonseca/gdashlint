package output

import (
	"fmt"
	"io"

	"github.com/hugomcfonseca/gdashlint/internal/lint"
)

// Render writes result in the requested format and sort order.
func Render(writer io.Writer, result lint.Result, format string, sortMode string) error {
	Sort(result.Findings, sortMode)
	switch format {
	case "text":
		return Text(writer, result, sortMode)
	case "json":
		return JSON(writer, result)
	case "github":
		return GitHub(writer, result)
	default:
		return fmt.Errorf("unknown output format %q", format)
	}
}
