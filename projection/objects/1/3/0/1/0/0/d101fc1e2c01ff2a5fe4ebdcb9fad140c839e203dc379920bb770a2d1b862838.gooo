package analyzer

import (
	"bytes"
	"fmt"
	"strings"
)

func parseProtectedSlotMarker(line []byte) (protectedSlotMarker, bool, error) {
	trimmed := strings.TrimSpace(string(line))
	const prefix = "//gooo:slot:"
	if !strings.HasPrefix(trimmed, prefix) {
		return protectedSlotMarker{}, false, nil
	}
	rest := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
	fields := annotationFields(rest)
	if len(fields) == 0 || (fields[0] != "start" && fields[0] != "end") {
		return protectedSlotMarker{}, false, fmt.Errorf("slot marker kind is invalid")
	}
	marker := protectedSlotMarker{kind: fields[0]}
	for _, field := range fields[1:] {
		key, value, ok := strings.Cut(field, "=")
		if ok && strings.TrimSpace(key) == "id" {
			marker.id = strings.TrimSpace(value)
		}
	}
	if marker.id == "" {
		return protectedSlotMarker{}, false, fmt.Errorf("slot identity is missing")
	}
	return marker, true, nil
}

type sourceLine struct {
	text       []byte
	start, end int
	next       int
}

func sourceLines(source []byte) []sourceLine {
	var lines []sourceLine
	for start := 0; start < len(source); {
		end := bytes.IndexByte(source[start:], '\n')
		if end < 0 {
			end = len(source)
		} else {
			end += start
		}
		lineEnd := end
		if lineEnd > start && source[lineEnd-1] == '\r' {
			lineEnd--
		}
		next := end
		if next < len(source) {
			next++
		}
		lines = append(lines, sourceLine{source[start:lineEnd], start, lineEnd, next})
		if next == len(source) {
			break
		}
		start = next
	}
	if len(source) == 0 {
		return []sourceLine{{text: nil}}
	}
	return lines
}
