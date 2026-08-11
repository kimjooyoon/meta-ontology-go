package roundtrip

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	generatedStart = "//gooo:generated:start"
	generatedEnd   = "//gooo:generated:end"
	slotStart      = "//gooo:slot:start"
	slotEnd        = "//gooo:slot:end"
)

type generatedRegion struct {
	ID           string
	Kind         string
	Body         []byte
	contentStart int
}

type generatedFile struct {
	Regions map[string]generatedRegion
}

func parseGeneratedFile(source []byte) (generatedFile, error) {
	result := generatedFile{Regions: make(map[string]generatedRegion)}
	var open *generatedRegion
	slotID := ""
	for _, line := range sourceLines(source) {
		marker, attributes, ok, err := parseMarker(line.text)
		if err != nil {
			return generatedFile{}, fmt.Errorf("line %d: %w", line.number, err)
		}
		if !ok {
			continue
		}
		switch marker {
		case "start":
			if open != nil {
				return generatedFile{}, fmt.Errorf("line %d: nested generated region %q", line.number, open.ID)
			}
			id := attributes["id"]
			if id == "" {
				return generatedFile{}, fmt.Errorf("line %d: generated region has no ID", line.number)
			}
			if err := validID(id); err != nil {
				return generatedFile{}, fmt.Errorf("line %d: invalid generated region ID: %w", line.number, err)
			}
			if _, exists := result.Regions[id]; exists {
				return generatedFile{}, fmt.Errorf("line %d: duplicate generated region %q", line.number, id)
			}
			open = &generatedRegion{ID: id, Kind: attributes["kind"], contentStart: line.end}
		case "slot-start":
			if open == nil {
				return generatedFile{}, fmt.Errorf("line %d: slot starts outside generated region", line.number)
			}
			if slotID != "" {
				return generatedFile{}, fmt.Errorf("line %d: nested slot %q", line.number, slotID)
			}
			slotID = attributes["id"]
			if slotID == "" {
				return generatedFile{}, fmt.Errorf("line %d: slot has no ID", line.number)
			}
			if err := validID(slotID); err != nil {
				return generatedFile{}, fmt.Errorf("line %d: invalid slot ID: %w", line.number, err)
			}
		case "slot-end":
			if slotID == "" {
				return generatedFile{}, fmt.Errorf("line %d: slot ends without start", line.number)
			}
			if id := attributes["id"]; id != "" && id != slotID {
				return generatedFile{}, fmt.Errorf("line %d: slot %q closes as %q", line.number, slotID, id)
			}
			slotID = ""
		case "end":
			if open == nil {
				return generatedFile{}, fmt.Errorf("line %d: generated end without start", line.number)
			}
			if slotID != "" {
				return generatedFile{}, fmt.Errorf("line %d: generated region closes with open slot %q", line.number, slotID)
			}
			if id := attributes["id"]; id != "" && id != open.ID {
				return generatedFile{}, fmt.Errorf("line %d: region %q closes as %q", line.number, open.ID, id)
			}
			var err error
			open.Body, err = canonicalRegionBody(source[open.contentStart:line.start])
			if err != nil {
				return generatedFile{}, fmt.Errorf("region %q: %w", open.ID, err)
			}
			result.Regions[open.ID] = *open
			open = nil
		}
	}
	if open != nil {
		return generatedFile{}, fmt.Errorf("unterminated generated region %q", open.ID)
	}
	return result, nil
}

func parseMarker(line string) (string, map[string]string, bool, error) {
	trimmed := strings.TrimSpace(line)
	prefix, marker := markerPrefix(trimmed)
	if prefix == "" {
		return "", nil, false, nil
	}
	if len(trimmed) > len(prefix) && trimmed[len(prefix)] != ' ' && trimmed[len(prefix)] != '\t' {
		return "", nil, false, nil
	}
	attributes, err := parseAttributes(strings.TrimSpace(trimmed[len(prefix):]))
	return marker, attributes, true, err
}

func markerPrefix(line string) (string, string) {
	switch {
	case strings.HasPrefix(line, generatedStart):
		return generatedStart, "start"
	case strings.HasPrefix(line, generatedEnd):
		return generatedEnd, "end"
	case strings.HasPrefix(line, slotStart):
		return slotStart, "slot-start"
	case strings.HasPrefix(line, slotEnd):
		return slotEnd, "slot-end"
	default:
		return "", ""
	}
}

func canonicalRegionBody(source []byte) ([]byte, error) {
	var result []byte
	open := false
	for _, line := range sourceLines(source) {
		marker, _, ok, err := parseMarker(line.text)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line.number, err)
		}
		if !ok {
			if !open {
				result = append(result, source[line.start:line.end]...)
			}
			continue
		}
		switch marker {
		case "slot-start":
			if open {
				return nil, fmt.Errorf("line %d: nested slot", line.number)
			}
			open = true
			result = append(result, source[line.start:line.end]...)
		case "slot-end":
			if !open {
				return nil, fmt.Errorf("line %d: slot ends without start", line.number)
			}
			open = false
			result = append(result, source[line.start:line.end]...)
		default:
			if !open {
				result = append(result, source[line.start:line.end]...)
			}
		}
	}
	if open {
		return nil, fmt.Errorf("unterminated slot")
	}
	return result, nil
}

func parseAttributes(input string) (map[string]string, error) {
	attributes := make(map[string]string)
	for index := 0; index < len(input); {
		index = skipSpaces(input, index)
		if index == len(input) {
			break
		}
		keyStart := index
		for index < len(input) && attributeChar(input[index]) {
			index++
		}
		if keyStart == index || index == len(input) || input[index] != '=' {
			return nil, fmt.Errorf("invalid marker attribute near byte %d", keyStart)
		}
		key := input[keyStart:index]
		index++
		if index == len(input) || input[index] != '"' {
			return nil, fmt.Errorf("marker attribute %q must be quoted", key)
		}
		valueStart := index
		index = quotedEnd(input, index)
		if index < 0 {
			return nil, fmt.Errorf("unterminated marker attribute %q", key)
		}
		value, err := strconv.Unquote(input[valueStart : index+1])
		if err != nil {
			return nil, fmt.Errorf("marker attribute %q: %w", key, err)
		}
		if _, exists := attributes[key]; exists {
			return nil, fmt.Errorf("duplicate marker attribute %q", key)
		}
		attributes[key] = value
		index++
	}
	return attributes, nil
}

func sourceLines(source []byte) []struct {
	number int
	text   string
	start  int
	end    int
} {
	parts := strings.Split(string(source), "\n")
	lines := make([]struct {
		number int
		text   string
		start  int
		end    int
	}, 0, len(parts))
	offset := 0
	for index, part := range parts {
		end := offset + len(part)
		if index < len(parts)-1 {
			end++
		}
		lines = append(lines, struct {
			number int
			text   string
			start  int
			end    int
		}{number: index + 1, text: strings.TrimSuffix(part, "\r"), start: offset, end: end})
		offset = end
	}
	return lines
}

func skipSpaces(value string, index int) int {
	for index < len(value) && (value[index] == ' ' || value[index] == '\t') {
		index++
	}
	return index
}

func quotedEnd(value string, start int) int {
	escaped := false
	for index := start + 1; index < len(value); index++ {
		if escaped {
			escaped = false
			continue
		}
		if value[index] == '\\' {
			escaped = true
			continue
		}
		if value[index] == '"' {
			return index
		}
	}
	return -1
}

func attributeChar(value byte) bool {
	return value == '_' || value == '-' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}
