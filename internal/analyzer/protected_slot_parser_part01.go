package analyzer

import (
	"bytes"
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
