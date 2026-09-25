package roundtrip

import (
	"bytes"
)

func splitSourceLines(source []byte) []markerLine {
	if len(source) == 0 {
		return nil
	}
	lines := make([]markerLine, 0, bytes.Count(source, []byte{'\n'})+1)
	for start := 0; start < len(source); {
		newline := bytes.IndexByte(source[start:], '\n')
		end := len(source)
		if newline >= 0 {
			end = start + newline + 1
		}
		textEnd := end
		if textEnd > start && source[textEnd-1] == '\n' {
			textEnd--
		}
		if textEnd > start && source[textEnd-1] == '\r' {
			textEnd--
		}
		lines = append(lines, markerLine{Number: len(lines) + 1, Text: string(source[start:textEnd]), Start: start, End: end})
		if newline < 0 {
			break
		}
		start = end
	}
	return lines
}
func skipSpaces(value string, index int) int {
	for index < len(value) && (value[index] == ' ' || value[index] == '\t') {
		index++
	}
	return index
}
func attributeKeyChar(value byte) bool {
	return value == '_' || value == '-' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}
