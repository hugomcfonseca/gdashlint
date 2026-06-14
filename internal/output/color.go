package output

import (
	"io"
	"os"
)

// ANSI color escape sequences used for diff output.
const (
	ansiReset = "\x1b[0m"
	ansiBold  = "\x1b[1m"
	ansiRed   = "\x1b[31m"
	ansiGreen = "\x1b[32m"
	ansiCyan  = "\x1b[36m"
)

// IsColorWriter reports whether w is an interactive terminal that supports
// ANSI colors. It returns false when:
//   - the NO_COLOR environment variable is set (https://no-color.org)
//   - TERM is set to "dumb"
//   - w is not an *os.File backed by a character device (e.g. a pipe or buffer)
func IsColorWriter(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
