package generator

import "fmt"

type markerState struct {
	result parsedMarkers
	region *generatedRegion
	slot   *parsedSlot
}

func (s *markerState) apply(marker string, attrs map[string]string, line sourceLine, index int, source []byte) error {
	lineNumber := index + 1
	switch marker {
	case "generated-start":
		return s.startRegion(attrs, line, lineNumber)
	case "generated-end":
		return s.endRegion(attrs, line, index, lineNumber)
	case "slot-start":
		return s.startSlot(attrs, line, lineNumber)
	case "slot-end":
		return s.endSlot(attrs, line, index, lineNumber, source)
	default:
		return fmt.Errorf("generator: unknown marker %q on line %d", marker, lineNumber)
	}
}

func (s *markerState) startRegion(attrs map[string]string, line sourceLine, lineNumber int) error {
	if s.region != nil {
		return fmt.Errorf("generator: nested generated region on line %d", lineNumber)
	}
	id := attrs["id"]
	if id == "" {
		return fmt.Errorf("generator: generated region on line %d has no ID", lineNumber)
	}
	s.region = &generatedRegion{ID: id, Kind: attrs["kind"], Start: line.start, StartLine: lineNumber - 1}
	return nil
}

func (s *markerState) endRegion(attrs map[string]string, line sourceLine, index, lineNumber int) error {
	if s.region == nil {
		return fmt.Errorf("generator: generated end without start on line %d", lineNumber)
	}
	if s.slot != nil {
		return fmt.Errorf("generator: generated region %q closes with an open slot on line %d", s.region.ID, lineNumber)
	}
	if endID := attrs["id"]; endID != s.region.ID {
		return fmt.Errorf("generator: generated region %q closes as %q on line %d", s.region.ID, endID, lineNumber)
	}
	if endKind := attrs["kind"]; endKind != s.region.Kind {
		return fmt.Errorf("generator: generated region %q closes with kind %q instead of %q on line %d", s.region.ID, endKind, s.region.Kind, lineNumber)
	}
	s.region.End, s.region.EndLine = line.end, index
	for _, existing := range s.result.Regions {
		if existing.ID == s.region.ID {
			return fmt.Errorf("generator: duplicate generated region ID %q", s.region.ID)
		}
	}
	s.result.Regions = append(s.result.Regions, *s.region)
	s.region = nil
	return nil
}

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
