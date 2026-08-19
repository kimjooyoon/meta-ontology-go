package generator

import (
	"fmt"
	"strings"
)

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
		return "", nil, false, fmt.Errorf("invalid %s marker line boundary", marker)
	}
	attrs, err := parseAttributes(strings.TrimSpace(trimmed[len(prefix):]))
	if err != nil {
		return "", nil, false, err
	}
	if err := validateMarkerAttributes(marker, attrs); err != nil {
		return "", nil, false, err
	}
	return marker, attrs, true, nil
}
func validateMarkerAttributes(marker string, attrs map[string]string) error {
	allowed := map[string]struct{}{"id": {}}
	if marker == "generated-start" || marker == "generated-end" {
		allowed["kind"] = struct{}{}
	}
	for key := range attrs {
		if _, exists := allowed[key]; !exists {
			return fmt.Errorf("unknown %s attribute %q", marker, key)
		}
	}
	if attrs["id"] == "" {
		return fmt.Errorf("%s marker requires a non-empty id", marker)
	}
	if marker == "generated-start" || marker == "generated-end" {
		kind := attrs["kind"]
		if kind != "entity" && kind != "activity" {
			return fmt.Errorf("%s marker requires kind entity or activity", marker)
		}
	}
	return nil
}
