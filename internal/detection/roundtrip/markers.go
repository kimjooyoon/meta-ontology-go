package roundtrip

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const (
	generatedStartPrefix = "//gooo:generated:start"
	generatedEndPrefix   = "//gooo:generated:end"
	slotStartPrefix      = "//gooo:slot:start"
	slotEndPrefix        = "//gooo:slot:end"
)

type generatedRegion struct {
	ID   semantic.ID
	Kind string
	Body []byte
}

type generatedFile struct {
	Regions map[semantic.ID]generatedRegion
}

type markerLine struct {
	Number     int
	Text       string
	Start, End int
}

type openRegion struct {
	ID           semantic.ID
	Kind         string
	ContentStart int
}

func parseGeneratedFile(source []byte) (generatedFile, error) {
	state := markerState{result: generatedFile{Regions: make(map[semantic.ID]generatedRegion)}, source: source}
	for _, line := range splitSourceLines(source) {
		if err := state.apply(line); err != nil {
			return generatedFile{}, err
		}
	}
	return state.finish()
}

type markerState struct {
	result generatedFile
	source []byte
	open   *openRegion
	slotID semantic.ID
}

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
	s.slotID = id
	return nil
}

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

func markerPrefix(line string) (string, string) {
	switch {
	case strings.HasPrefix(line, generatedStartPrefix):
		return generatedStartPrefix, "generated-start"
	case strings.HasPrefix(line, generatedEndPrefix):
		return generatedEndPrefix, "generated-end"
	case strings.HasPrefix(line, slotStartPrefix):
		return slotStartPrefix, "slot-start"
	case strings.HasPrefix(line, slotEndPrefix):
		return slotEndPrefix, "slot-end"
	default:
		return "", ""
	}
}

func validateMarkerAttributes(marker string, attributes map[string]string) error {
	allowed := map[string]struct{}{"id": {}}
	if marker == "generated-start" || marker == "generated-end" {
		allowed["kind"] = struct{}{}
	}
	for key := range attributes {
		if _, ok := allowed[key]; !ok {
			return fmt.Errorf("unknown %s attribute %q", marker, key)
		}
	}
	if attributes["id"] == "" {
		return fmt.Errorf("%s marker requires a non-empty id", marker)
	}
	if marker == "generated-start" || marker == "generated-end" {
		kind := attributes["kind"]
		if kind != "entity" && kind != "activity" {
			return fmt.Errorf("%s marker requires kind entity or activity", marker)
		}
	}
	return nil
}

func parseMarkerID(raw string) (semantic.ID, error) {
	return semantic.ParseIdentity(raw)
}

func canonicalRegionBody(source []byte) ([]byte, error) {
	var result []byte
	insideSlot := false
	for _, line := range splitSourceLines(source) {
		marker, _, ok, err := parseMarker(line.Text)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line.Number, err)
		}
		if !ok {
			if !insideSlot {
				result = append(result, source[line.Start:line.End]...)
			}
			continue
		}
		switch marker {
		case "slot-start":
			if insideSlot {
				return nil, fmt.Errorf("line %d: nested slot", line.Number)
			}
			insideSlot = true
			result = append(result, source[line.Start:line.End]...)
		case "slot-end":
			if !insideSlot {
				return nil, fmt.Errorf("line %d: slot ends without start", line.Number)
			}
			insideSlot = false
			result = append(result, source[line.Start:line.End]...)
		default:
			return nil, fmt.Errorf("line %d: generated marker is not nested legally", line.Number)
		}
	}
	if insideSlot {
		return nil, fmt.Errorf("unterminated slot")
	}
	return result, nil
}
