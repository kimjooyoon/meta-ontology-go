package generator

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	generatedStartPrefix = "//gooo:generated:start"
	generatedEndPrefix   = "//gooo:generated:end"
	slotStartPrefix      = "//gooo:slot:start"
	slotEndPrefix        = "//gooo:slot:end"
)

type sourceLine struct {
	start int
	end   int
	text  string
}

type generatedRegion struct {
	ID        string
	Kind      string
	Start     int
	End       int
	StartLine int
	EndLine   int
	Slots     []parsedSlot
}

type parsedSlot struct {
	ID        string
	Start     int
	End       int
	StartLine int
	EndLine   int
	Body      []byte
}

type parsedMarkers struct {
	Regions []generatedRegion
	Slots   map[string]parsedSlot
}

func generatedMarker(prefix, id, kind string) string {
	if kind == "" {
		return prefix + " id=" + strconv.Quote(id)
	}
	return prefix + " id=" + strconv.Quote(id) + " kind=" + strconv.Quote(kind)
}

func slotMarker(prefix, id string) string {
	return prefix + " id=" + strconv.Quote(id)
}

func parseMarkers(source []byte) (parsedMarkers, error) {
	lines := splitSourceLines(source)
	result := parsedMarkers{Slots: make(map[string]parsedSlot)}
	var region *generatedRegion
	var slot *parsedSlot
	for index, line := range lines {
		marker, attrs, ok, err := parseMarker(line.text)
		if err != nil {
			return parsedMarkers{}, fmt.Errorf("generator: malformed marker on line %d: %w", index+1, err)
		}
		if !ok {
			continue
		}
		switch marker {
		case "generated-start":
			if region != nil {
				return parsedMarkers{}, fmt.Errorf("generator: nested generated region on line %d", index+1)
			}
			id := attrs["id"]
			if id == "" {
				return parsedMarkers{}, fmt.Errorf("generator: generated region on line %d has no ID", index+1)
			}
			region = &generatedRegion{
				ID:        id,
				Kind:      attrs["kind"],
				Start:     line.start,
				StartLine: index,
			}
		case "generated-end":
			if region == nil {
				return parsedMarkers{}, fmt.Errorf("generator: generated end without start on line %d", index+1)
			}
			if slot != nil {
				return parsedMarkers{}, fmt.Errorf("generator: generated region %q closes with an open slot on line %d", region.ID, index+1)
			}
			if endID := attrs["id"]; endID != "" && endID != region.ID {
				return parsedMarkers{}, fmt.Errorf("generator: generated region %q closes as %q on line %d", region.ID, endID, index+1)
			}
			region.End = line.end
			region.EndLine = index
			for _, existing := range result.Regions {
				if existing.ID == region.ID {
					return parsedMarkers{}, fmt.Errorf("generator: duplicate generated region ID %q", region.ID)
				}
			}
			result.Regions = append(result.Regions, *region)
			region = nil
		case "slot-start":
			if region == nil {
				return parsedMarkers{}, fmt.Errorf("generator: slot outside generated region on line %d", index+1)
			}
			if slot != nil {
				return parsedMarkers{}, fmt.Errorf("generator: nested slot in region %q on line %d", region.ID, index+1)
			}
			id := attrs["id"]
			if id == "" {
				return parsedMarkers{}, fmt.Errorf("generator: slot on line %d has no ID", index+1)
			}
			slot = &parsedSlot{ID: id, Start: line.end, StartLine: index}
		case "slot-end":
			if slot == nil {
				return parsedMarkers{}, fmt.Errorf("generator: slot end without start on line %d", index+1)
			}
			if endID := attrs["id"]; endID != "" && endID != slot.ID {
				return parsedMarkers{}, fmt.Errorf("generator: slot %q closes as %q on line %d", slot.ID, endID, index+1)
			}
			slot.End = line.start
			slot.EndLine = index
			slot.Body = append([]byte(nil), source[slot.Start:slot.End]...)
			if _, exists := result.Slots[slot.ID]; exists {
				return parsedMarkers{}, fmt.Errorf("generator: duplicate slot ID %q", slot.ID)
			}
			result.Slots[slot.ID] = *slot
			region.Slots = append(region.Slots, *slot)
			slot = nil
		}
	}
	if slot != nil {
		return parsedMarkers{}, fmt.Errorf("generator: unterminated slot %q", slot.ID)
	}
	if region != nil {
		return parsedMarkers{}, fmt.Errorf("generator: unterminated generated region %q", region.ID)
	}
	return result, nil
}

func parseMarker(line string) (string, map[string]string, bool, error) {
	trimmed := strings.TrimSpace(line)
	prefix := ""
	marker := ""
	switch {
	case strings.HasPrefix(trimmed, generatedStartPrefix):
		prefix, marker = generatedStartPrefix, "generated-start"
	case strings.HasPrefix(trimmed, generatedEndPrefix):
		prefix, marker = generatedEndPrefix, "generated-end"
	case strings.HasPrefix(trimmed, slotStartPrefix):
		prefix, marker = slotStartPrefix, "slot-start"
	case strings.HasPrefix(trimmed, slotEndPrefix):
		prefix, marker = slotEndPrefix, "slot-end"
	default:
		return "", nil, false, nil
	}
	if len(trimmed) > len(prefix) && trimmed[len(prefix)] != ' ' && trimmed[len(prefix)] != '\t' {
		return "", nil, false, nil
	}
	attrs, err := parseAttributes(strings.TrimSpace(trimmed[len(prefix):]))
	if err != nil {
		return "", nil, false, err
	}
	return marker, attrs, true, nil
}

func parseAttributes(input string) (map[string]string, error) {
	attrs := make(map[string]string)
	for index := 0; index < len(input); {
		for index < len(input) && (input[index] == ' ' || input[index] == '\t') {
			index++
		}
		if index == len(input) {
			break
		}
		keyStart := index
		for index < len(input) && isAttributeKeyChar(input[index]) {
			index++
		}
		if keyStart == index {
			return nil, fmt.Errorf("expected attribute name at byte %d", keyStart)
		}
		key := input[keyStart:index]
		if index >= len(input) || input[index] != '=' {
			return nil, fmt.Errorf("expected '=' after attribute %q", key)
		}
		index++
		if index >= len(input) || input[index] != '"' {
			return nil, fmt.Errorf("attribute %q must use a quoted value", key)
		}
		valueStart := index
		index++
		escaped := false
		for index < len(input) {
			if escaped {
				escaped = false
				index++
				continue
			}
			if input[index] == '\\' {
				escaped = true
				index++
				continue
			}
			if input[index] == '"' {
				index++
				break
			}
			index++
		}
		if index > len(input) || input[index-1] != '"' {
			return nil, fmt.Errorf("unterminated attribute %q", key)
		}
		value, err := strconv.Unquote(input[valueStart:index])
		if err != nil {
			return nil, fmt.Errorf("invalid attribute %q: %w", key, err)
		}
		if _, exists := attrs[key]; exists {
			return nil, fmt.Errorf("duplicate attribute %q", key)
		}
		attrs[key] = value
	}
	return attrs, nil
}

func isAttributeKeyChar(value byte) bool {
	return value == '_' || value == '-' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9'
}

func splitSourceLines(source []byte) []sourceLine {
	if len(source) == 0 {
		return nil
	}
	var lines []sourceLine
	start := 0
	for start < len(source) {
		newline := strings.IndexByte(string(source[start:]), '\n')
		if newline < 0 {
			end := len(source)
			textEnd := end
			if textEnd > start && source[textEnd-1] == '\r' {
				textEnd--
			}
			lines = append(lines, sourceLine{start: start, end: end, text: string(source[start:textEnd])})
			break
		}
		end := start + newline + 1
		textEnd := end - 1
		if textEnd > start && source[textEnd-1] == '\r' {
			textEnd--
		}
		lines = append(lines, sourceLine{start: start, end: end, text: string(source[start:textEnd])})
		start = end
	}
	return lines
}
