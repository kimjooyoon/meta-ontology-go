package generator

import (
	"strings"
)

func splitSourceLines(source []byte) []sourceLine {
	if len(source) == 0 {
		return nil
	}
	var lines []sourceLine
	start := 0
	for start < len(source) {
		newline := strings.IndexByte(string(source[start:]), '\n')
		if newline < 0 {
			end := len(source)
			textEnd := end
			if textEnd > start && source[textEnd-1] == '\r' {
				textEnd--
			}
			lines = append(lines, sourceLine{start: start, end: end, text: string(source[start:textEnd])})
			break
		}
		end := start + newline + 1
		textEnd := end - 1
		if textEnd > start && source[textEnd-1] == '\r' {
			textEnd--
		}
		lines = append(lines, sourceLine{start: start, end: end, text: string(source[start:textEnd])})
		start = end
	}
	return lines
}
