package analyzer

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func parseProtectedSlots(source SourceFile) ([]protectedSlot, error) {
	var slots []protectedSlot
	var open *protectedSlotMarker
	for _, line := range sourceLines(source.Source) {
		marker, ok, err := parseProtectedSlotMarker(line.text)
		if err != nil {
			return nil, slotError(source.Filename, err.Error())
		}
		if !ok {
			continue
		}
		marker.start = line.start + bytes.Index(line.text, []byte("//gooo:slot:"))
		marker.lineEnd = line.end
		marker.next = line.next
		switch marker.kind {
		case "start":
			if open != nil {
				return nil, slotError(source.Filename, "nested slot start marker")
			}
			open = &marker
		case "end":
			if open == nil {
				return nil, slotError(source.Filename, "slot end marker has no start")
			}
			if open.id != marker.id {
				return nil, slotError(source.Filename, "slot end identity does not match start")
			}
			slot, err := completeProtectedSlot(source, *open, marker)
			if err != nil {
				return nil, err
			}
			slots = append(slots, slot)
			open = nil
		}
	}
	if open != nil {
		return nil, slotError(source.Filename, "slot start marker has no end")
	}
	return slots, nil
}

func completeProtectedSlot(source SourceFile, start, end protectedSlotMarker) (protectedSlot, error) {
	parsed, err := semantic.ParseIdentity(start.id)
	if err != nil {
		return protectedSlot{}, slotError(source.Filename, "slot identity is invalid")
	}
	span := Span{
		Filename: source.Filename,
		Start:    sourcePosition(source.Filename, source.Source, start.start),
		End:      sourcePosition(source.Filename, source.Source, end.lineEnd),
	}
	bodySpan := Span{
		Filename: source.Filename,
		Start:    sourcePosition(source.Filename, source.Source, start.next),
		End:      sourcePosition(source.Filename, source.Source, end.start),
	}
	if err := validateProtectedSpan(span); err != nil {
		return protectedSlot{}, slotError(source.Filename, err.Error())
	}
	if err := validateProtectedSpan(bodySpan); err != nil {
		return protectedSlot{}, slotError(source.Filename, err.Error())
	}
	return protectedSlot{
		SourceFile: source.Filename, SlotID: parsed.String(), Span: span,
		BodySpan: bodySpan, BodyDigest: semantic.StableHash(source.Source[start.next:end.start]),
	}, nil
}

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
