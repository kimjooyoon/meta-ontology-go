package generator

import (
	"fmt"
)

func (s *markerState) startSlot(attrs map[string]string, line sourceLine, lineNumber int) error {
	if s.region == nil {
		return fmt.Errorf("generator: slot outside generated region on line %d", lineNumber)
	}
	if s.slot != nil {
		return fmt.Errorf("generator: nested slot in region %q on line %d", s.region.ID, lineNumber)
	}
	id := attrs["id"]
	if id == "" {
		return fmt.Errorf("generator: slot on line %d has no ID", lineNumber)
	}
	s.slot = &parsedSlot{ID: id, RegionID: s.region.ID, RegionKind: s.region.Kind, Start: line.end, StartLine: lineNumber - 1}
	return nil
}
func (s *markerState) endSlot(attrs map[string]string, line sourceLine, index, lineNumber int, source []byte) error {
	if s.slot == nil {
		return fmt.Errorf("generator: slot end without start on line %d", lineNumber)
	}
	if endID := attrs["id"]; endID != s.slot.ID {
		return fmt.Errorf("generator: slot %q closes as %q on line %d", s.slot.ID, endID, lineNumber)
	}
	s.slot.End, s.slot.EndLine = line.start, index
	s.slot.Body = append([]byte(nil), source[s.slot.Start:s.slot.End]...)
	if _, exists := s.result.Slots[s.slot.ID]; exists {
		return fmt.Errorf("generator: duplicate slot ID %q", s.slot.ID)
	}
	s.result.Slots[s.slot.ID] = *s.slot
	s.region.Slots = append(s.region.Slots, *s.slot)
	s.slot = nil
	return nil
}
func (s *markerState) finish() (parsedMarkers, error) {
	if s.slot != nil {
		return parsedMarkers{}, fmt.Errorf("generator: unterminated slot %q", s.slot.ID)
	}
	if s.region != nil {
		return parsedMarkers{}, fmt.Errorf("generator: unterminated generated region %q", s.region.ID)
	}
	for _, region := range s.result.Regions {
		if _, exists := s.result.Slots[region.ID]; exists {
			return parsedMarkers{}, fmt.Errorf("generator: stable ID %q is used by region and slot", region.ID)
		}
	}
	return s.result, nil
}
