package harness

import (
	"fmt"
	"strings"
)

const (
	ansiRed   = "\033[31m"
	ansiGreen = "\033[32m"
	ansiReset = "\033[0m"
	ansiCyan  = "\033[36m"

	diffContextLines = 3
	diffMaxLines     = 30
)

// UnifiedDiff produces a colored unified diff between oldContent and newContent.
// filename is used in the header. Output is truncated to diffMaxLines of diff body.
func UnifiedDiff(oldContent, newContent, filename string) string {
	oldLines := splitLines(oldContent)
	newLines := splitLines(newContent)

	firstDiff, lastDiff := diffBounds(oldLines, newLines)
	if firstDiff < 0 {
		return "" // identical
	}

	// context window
	ctxStart := firstDiff - diffContextLines
	if ctxStart < 0 {
		ctxStart = 0
	}

	oldEnd := lastDiff + diffContextLines + 1
	if oldEnd > len(oldLines) {
		oldEnd = len(oldLines)
	}
	newEnd := lastDiff + diffContextLines + 1
	if newEnd > len(newLines) {
		newEnd = len(newLines)
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s--- a/%s%s\n", ansiRed, filename, ansiReset))
	sb.WriteString(fmt.Sprintf("%s+++ b/%s%s\n", ansiGreen, filename, ansiReset))

	oldCount := oldEnd - ctxStart
	newCount := newEnd - ctxStart
	sb.WriteString(fmt.Sprintf("%s@@ -%d,%d +%d,%d @@%s\n",
		ansiCyan, ctxStart+1, oldCount, ctxStart+1, newCount, ansiReset))

	diffLines := 0
	truncated := 0

	// emit context before
	for i := ctxStart; i < firstDiff && i < len(oldLines); i++ {
		if diffLines >= diffMaxLines {
			truncated++
			continue
		}
		sb.WriteString(" " + oldLines[i] + "\n")
		diffLines++
	}

	// emit changed region
	// removed lines from old
	for i := firstDiff; i <= lastDiff && i < len(oldLines); i++ {
		if diffLines >= diffMaxLines {
			truncated++
			continue
		}
		sb.WriteString(fmt.Sprintf("%s-%s%s\n", ansiRed, oldLines[i], ansiReset))
		diffLines++
	}
	// added lines from new
	for i := firstDiff; i <= lastDiff && i < len(newLines); i++ {
		if diffLines >= diffMaxLines {
			truncated++
			continue
		}
		sb.WriteString(fmt.Sprintf("%s+%s%s\n", ansiGreen, newLines[i], ansiReset))
		diffLines++
	}

	// emit context after
	afterStart := lastDiff + 1
	for i := afterStart; i < oldEnd && i < len(oldLines); i++ {
		if diffLines >= diffMaxLines {
			truncated++
			continue
		}
		sb.WriteString(" " + oldLines[i] + "\n")
		diffLines++
	}

	if truncated > 0 {
		sb.WriteString(fmt.Sprintf("[... %d more lines]\n", truncated))
	}

	return sb.String()
}

// splitLines splits content into lines, handling trailing newline gracefully.
func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	// drop trailing empty element from trailing newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// diffBounds finds the first and last differing line indices.
// Returns (-1, -1) if contents are identical.
func diffBounds(oldLines, newLines []string) (first, last int) {
	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}
	if maxLen == 0 {
		return -1, -1
	}

	first = -1
	for i := 0; i < maxLen; i++ {
		var o, n string
		if i < len(oldLines) {
			o = oldLines[i]
		}
		if i < len(newLines) {
			n = newLines[i]
		}
		if o != n {
			if first < 0 {
				first = i
			}
			last = i
		}
	}

	return first, last
}
