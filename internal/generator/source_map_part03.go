package generator

import (
	"strings"
)

func findGeneratedFieldLine(source []byte, lines []sourceLine, region generatedRegion, next int, field Field) (bool, generatedFieldLine) {
	for lineIndex := next; lineIndex < len(lines); lineIndex++ {
		line := lines[lineIndex]
		if line.start < region.Start || line.end > region.End || !matchesGeneratedFieldLine(strings.TrimSpace(line.text), field) {
			continue
		}
		return true, generatedFieldLine{
			rangeValue: SourceRange{Start: positionAt(source, line.start), End: positionAt(source, line.end)},
			next:       lineIndex + 1,
		}
	}
	return false, generatedFieldLine{}
}
func matchesGeneratedFieldLine(line string, field Field) bool {
	parts := strings.Fields(line)
	return len(parts) == 2 && parts[0] == field.GoName && parts[1] == field.GoType
}
func rangeForOffsets(source []byte, start, end int) SourceRange {
	return SourceRange{Start: positionAt(source, start), End: positionAt(source, end)}
}
func positionAt(source []byte, offset int) Position {
	if offset < 0 {
		offset = 0
	}
	if offset > len(source) {
		offset = len(source)
	}
	line := 1
	column := 1
	for _, value := range source[:offset] {
		if value == '\n' {
			line++
			column = 1
			continue
		}
		column++
	}
	return Position{Offset: offset, Line: line, Column: column}
}
