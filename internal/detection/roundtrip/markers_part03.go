package roundtrip

import (
	"fmt"
	"strings"
)

func (s *markerState) endSlot(line markerLine, attributes map[string]string) error {
	if s.slotID == "" {
		return fmt.Errorf("slot ends without start")
	}
	id, err := parseMarkerID(attributes["id"])
	if err != nil {
		return fmt.Errorf("slot ID: %w", err)
	}
	if id != s.slotID {
		return fmt.Errorf("slot %q closes as %q", s.slotID, id)
	}
	s.closedSlotIDs[s.slotID] = struct{}{}
	s.slotID = ""
	return nil
}
func (s *markerState) endRegion(line markerLine, attributes map[string]string) error {
	if s.open == nil {
		return fmt.Errorf("generated end without start")
	}
	if s.slotID != "" {
		return fmt.Errorf("generated region closes with open slot %q", s.slotID)
	}
	id, err := parseMarkerID(attributes["id"])
	if err != nil {
		return fmt.Errorf("generated region ID: %w", err)
	}
	if id != s.open.ID {
		return fmt.Errorf("region %q closes as %q", s.open.ID, id)
	}
	if attributes["kind"] != s.open.Kind {
		return fmt.Errorf("region %q closes with kind %q instead of %q", s.open.ID, attributes["kind"], s.open.Kind)
	}
	body, err := canonicalRegionBody(s.source[s.open.ContentStart:line.Start])
	if err != nil {
		return fmt.Errorf("region %q: %w", s.open.ID, err)
	}
	s.result.Regions[s.open.ID] = generatedRegion{ID: s.open.ID, Kind: s.open.Kind, Body: body}
	s.open = nil
	return nil
}
func (s *markerState) finish() (generatedFile, error) {
	if s.slotID != "" {
		return generatedFile{}, fmt.Errorf("unterminated slot %q", s.slotID)
	}
	if s.open != nil {
		return generatedFile{}, fmt.Errorf("unterminated generated region %q", s.open.ID)
	}
	return s.result, nil
}
func parseMarker(line string) (string, map[string]string, bool, error) {
	trimmed := strings.TrimSpace(line)
	prefix, marker := markerPrefix(trimmed)
	if prefix == "" {
		return "", nil, false, nil
	}
	if len(trimmed) > len(prefix) && trimmed[len(prefix)] != ' ' && trimmed[len(prefix)] != '\t' {
		return "", nil, false, fmt.Errorf("invalid %s marker line boundary", marker)
	}
	attributes, err := parseAttributes(strings.TrimSpace(trimmed[len(prefix):]))
	if err != nil {
		return "", nil, false, err
	}
	if err := validateMarkerAttributes(marker, attributes); err != nil {
		return "", nil, false, err
	}
	return marker, attributes, true, nil
}
