package roundtrip

import (
	"fmt"
)

func (s *markerState) apply(line markerLine) error {
	marker, attributes, ok, err := parseMarker(line.Text)
	if err != nil {
		return fmt.Errorf("line %d: %w", line.Number, err)
	}
	if !ok {
		return nil
	}
	var applyErr error
	switch marker {
	case "generated-start":
		applyErr = s.startRegion(line, attributes)
	case "slot-start":
		applyErr = s.startSlot(line, attributes)
	case "slot-end":
		applyErr = s.endSlot(line, attributes)
	case "generated-end":
		applyErr = s.endRegion(line, attributes)
	}
	if applyErr != nil {
		return fmt.Errorf("line %d: %w", line.Number, applyErr)
	}
	return nil
}
func (s *markerState) startRegion(line markerLine, attributes map[string]string) error {
	if s.open != nil {
		return fmt.Errorf("nested generated region %q", s.open.ID)
	}
	id, err := parseMarkerID(attributes["id"])
	if err != nil {
		return fmt.Errorf("generated region ID: %w", err)
	}
	if _, exists := s.result.Regions[id]; exists {
		return fmt.Errorf("duplicate generated region %q", id)
	}
	if _, exists := s.closedSlotIDs[id]; exists {
		return fmt.Errorf("generated region %q collides with slot identity", id)
	}
	s.open = &openRegion{ID: id, Kind: attributes["kind"], ContentStart: line.End}
	return nil
}
func (s *markerState) startSlot(line markerLine, attributes map[string]string) error {
	if s.open == nil {
		return fmt.Errorf("slot starts outside generated region")
	}
	if s.slotID != "" {
		return fmt.Errorf("nested slot %q", s.slotID)
	}
	id, err := parseMarkerID(attributes["id"])
	if err != nil {
		return fmt.Errorf("slot ID: %w", err)
	}
	if id == s.open.ID {
		return fmt.Errorf("slot %q collides with generated region identity", id)
	}
	if _, exists := s.result.Regions[id]; exists {
		return fmt.Errorf("slot %q collides with generated region identity", id)
	}
	if _, exists := s.closedSlotIDs[id]; exists {
		return fmt.Errorf("duplicate slot %q", id)
	}
	s.slotID = id
	return nil
}
