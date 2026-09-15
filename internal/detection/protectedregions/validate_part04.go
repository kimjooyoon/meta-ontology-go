package protectedregions

import (
	"bytes"
	"fmt"
	"strings"
)

func parseMarker(line sourceLine, lineNumber int) (markerEvent, bool, error) {
	trimmed := strings.TrimSpace(line.text)
	for _, spec := range markerSpecs() {
		if !strings.HasPrefix(trimmed, spec.prefix) {
			continue
		}
		event := markerEvent{kind: spec.kind, boundary: spec.boundary, line: lineNumber, start: line.start, end: line.end, legacy: spec.legacy}
		if !hasMarkerPrefix(trimmed, spec.prefix) {
			return event, true, fmt.Errorf("invalid %s marker line boundary", spec.prefix)
		}
		rest := strings.TrimSpace(trimmed[len(spec.prefix):])
		id, semanticKind, err := markerAttributes(rest, spec)
		if err != nil {
			return event, true, err
		}
		event.id = id
		event.semanticKind = semanticKind
		return event, true, nil
	}
	return markerEvent{}, false, nil
}
func splitSourceLines(source []byte) []sourceLine {
	lines := make([]sourceLine, 0)
	for start := 0; start < len(source); {
		relativeEnd := bytes.IndexByte(source[start:], '\n')
		end := len(source)
		if relativeEnd >= 0 {
			end = start + relativeEnd + 1
		}
		textEnd := end
		if textEnd > start && source[textEnd-1] == '\n' {
			textEnd--
		}
		if textEnd > start && source[textEnd-1] == '\r' {
			textEnd--
		}
		lines = append(lines, sourceLine{start: start, end: end, text: string(source[start:textEnd])})
		start = end
	}
	return lines
}
