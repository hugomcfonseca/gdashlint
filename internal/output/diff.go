package output

import (
	"bytes"
	"fmt"
	"io"
)

const diffContextLines = 3

// UnifiedDiff writes a compact line-oriented unified diff for a fixed dashboard preview.
// When color is true and the writer supports ANSI escapes, removed lines are red,
// added lines are green, and hunk headers are cyan.
func UnifiedDiff(writer io.Writer, fromPath string, toPath string, before []byte, after []byte, color bool) error {
	if bytes.Equal(before, after) {
		return nil
	}
	beforeLines := splitLines(before)
	afterLines := splitLines(after)
	changes := diffLines(beforeLines, afterLines)
	hunks := diffHunks(changes, diffContextLines)
	if len(hunks) == 0 {
		return nil
	}
	if color {
		if _, err := fmt.Fprintf(writer, "%s--- %s\n+++ %s%s\n", ansiBold, fromPath, toPath, ansiReset); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(writer, "--- %s\n+++ %s\n", fromPath, toPath); err != nil {
			return err
		}
	}
	for _, hunk := range hunks {
		if color {
			if _, err := fmt.Fprintf(writer, "%s@@ -%d,%d +%d,%d @@%s\n", ansiCyan, hunk.beforeStart, hunk.beforeCount, hunk.afterStart, hunk.afterCount, ansiReset); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(writer, "@@ -%d,%d +%d,%d @@\n", hunk.beforeStart, hunk.beforeCount, hunk.afterStart, hunk.afterCount); err != nil {
				return err
			}
		}
		for _, change := range hunk.changes {
			if err := writeChange(writer, change, color); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintln(writer)
	return err
}

func writeChange(writer io.Writer, change diffChange, color bool) error {
	if !color {
		_, err := fmt.Fprintf(writer, "%c%s\n", change.kind, change.line)
		return err
	}
	switch change.kind {
	case '-':
		_, err := fmt.Fprintf(writer, "%s-%s%s\n", ansiRed, change.line, ansiReset)
		return err
	case '+':
		_, err := fmt.Fprintf(writer, "%s+%s%s\n", ansiGreen, change.line, ansiReset)
		return err
	default:
		_, err := fmt.Fprintf(writer, " %s\n", change.line)
		return err
	}
}

type diffChange struct {
	kind        byte
	line        string
	beforeLine  int
	afterLine   int
	beforeCount int
	afterCount  int
}

type diffHunk struct {
	beforeStart int
	beforeCount int
	afterStart  int
	afterCount  int
	changes     []diffChange
}

func diffLines(before []string, after []string) []diffChange {
	lcs := make([][]int, len(before)+1)
	for index := range lcs {
		lcs[index] = make([]int, len(after)+1)
	}
	for beforeIndex := len(before) - 1; beforeIndex >= 0; beforeIndex-- {
		for afterIndex := len(after) - 1; afterIndex >= 0; afterIndex-- {
			if before[beforeIndex] == after[afterIndex] {
				lcs[beforeIndex][afterIndex] = lcs[beforeIndex+1][afterIndex+1] + 1
				continue
			}
			lcs[beforeIndex][afterIndex] = max(lcs[beforeIndex+1][afterIndex], lcs[beforeIndex][afterIndex+1])
		}
	}

	changes := make([]diffChange, 0, len(before)+len(after))
	beforeIndex := 0
	afterIndex := 0
	for beforeIndex < len(before) && afterIndex < len(after) {
		if before[beforeIndex] == after[afterIndex] {
			changes = append(changes, diffChange{kind: ' ', line: before[beforeIndex], beforeLine: beforeIndex + 1, afterLine: afterIndex + 1, beforeCount: 1, afterCount: 1})
			beforeIndex++
			afterIndex++
			continue
		}
		if lcs[beforeIndex+1][afterIndex] >= lcs[beforeIndex][afterIndex+1] {
			changes = append(changes, diffChange{kind: '-', line: before[beforeIndex], beforeLine: beforeIndex + 1, afterLine: afterIndex + 1, beforeCount: 1})
			beforeIndex++
			continue
		}
		changes = append(changes, diffChange{kind: '+', line: after[afterIndex], beforeLine: beforeIndex + 1, afterLine: afterIndex + 1, afterCount: 1})
		afterIndex++
	}
	for beforeIndex < len(before) {
		changes = append(changes, diffChange{kind: '-', line: before[beforeIndex], beforeLine: beforeIndex + 1, afterLine: afterIndex + 1, beforeCount: 1})
		beforeIndex++
	}
	for afterIndex < len(after) {
		changes = append(changes, diffChange{kind: '+', line: after[afterIndex], beforeLine: beforeIndex + 1, afterLine: afterIndex + 1, afterCount: 1})
		afterIndex++
	}
	return changes
}

func diffHunks(changes []diffChange, contextLines int) []diffHunk {
	changeIndexes := make([]int, 0)
	for index, change := range changes {
		if change.kind != ' ' {
			changeIndexes = append(changeIndexes, index)
		}
	}
	if len(changeIndexes) == 0 {
		return nil
	}

	hunks := make([]diffHunk, 0)
	hunkStart := max(changeIndexes[0]-contextLines, 0)
	hunkEnd := min(changeIndexes[0]+contextLines+1, len(changes))
	for _, changeIndex := range changeIndexes[1:] {
		start := max(changeIndex-contextLines, 0)
		end := min(changeIndex+contextLines+1, len(changes))
		if start <= hunkEnd {
			hunkEnd = max(hunkEnd, end)
			continue
		}
		hunks = append(hunks, newDiffHunk(changes[hunkStart:hunkEnd]))
		hunkStart = start
		hunkEnd = end
	}
	hunks = append(hunks, newDiffHunk(changes[hunkStart:hunkEnd]))
	return hunks
}

func newDiffHunk(changes []diffChange) diffHunk {
	hunk := diffHunk{changes: changes}
	for _, change := range changes {
		if hunk.beforeStart == 0 && change.beforeCount > 0 {
			hunk.beforeStart = change.beforeLine
		}
		if hunk.afterStart == 0 && change.afterCount > 0 {
			hunk.afterStart = change.afterLine
		}
		hunk.beforeCount += change.beforeCount
		hunk.afterCount += change.afterCount
	}
	if hunk.beforeStart == 0 {
		hunk.beforeStart = changes[0].beforeLine
	}
	if hunk.afterStart == 0 {
		hunk.afterStart = changes[0].afterLine
	}
	return hunk
}

func splitLines(data []byte) []string {
	data = bytes.TrimSuffix(data, []byte("\n"))
	if len(data) == 0 {
		return nil
	}
	parts := bytes.Split(data, []byte("\n"))
	lines := make([]string, len(parts))
	for index, part := range parts {
		lines[index] = string(part)
	}
	return lines
}
