package analyzer

import (
	"fmt"
)

func sourcePosition(filename string, source []byte, offset int) Position {
	line, column := 1, 1
	for index := 0; index < offset && index < len(source); index++ {
		if source[index] == '\n' {
			line++
			column = 1
			continue
		}
		column++
	}
	return Position{Offset: offset, Line: line, Column: column}
}
func slotError(filename, detail string) error {
	return adapterError(AdapterSlotConfig, "", filename, detail)
}
func validateProtectedSpan(span Span) error {
	if span.Start.Offset < 0 || span.End.Offset < span.Start.Offset {
		return fmt.Errorf("slot span is invalid")
	}
	return nil
}
