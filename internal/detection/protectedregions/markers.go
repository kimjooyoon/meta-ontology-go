package protectedregions

import (
	"fmt"
	"strconv"
	"strings"
)

type markerSpec struct {
	prefix   string
	kind     MarkerKind
	boundary MarkerBoundary
	legacy   bool
}

func markerSpecs() []markerSpec {
	return []markerSpec{
		{prefix: "//gooo:generated:start", kind: Generated, boundary: Start},
		{prefix: "//gooo:generated:end", kind: Generated, boundary: End},
		{prefix: "//gooo:slot:start", kind: Slot, boundary: Start},
		{prefix: "//gooo:slot:end", kind: Slot, boundary: End},
		{prefix: "//gooo:handwritten:start", kind: Handwritten, boundary: Start},
		{prefix: "//gooo:handwritten:end", kind: Handwritten, boundary: End},
		{prefix: "//gooo:protected:start", kind: Handwritten, boundary: Start},
		{prefix: "//gooo:protected:end", kind: Handwritten, boundary: End},
		{prefix: "//gooo:generated begin", kind: Generated, boundary: Start, legacy: true},
		{prefix: "//gooo:generated end", kind: Generated, boundary: End, legacy: true},
	}
}

func hasMarkerPrefix(line, prefix string) bool {
	if !strings.HasPrefix(line, prefix) {
		return false
	}
	return len(line) == len(prefix) || line[len(prefix)] == ' ' || line[len(prefix)] == '\t'
}

func markerAttributes(input string, spec markerSpec) (string, string, error) {
	if spec.legacy {
		value := strings.TrimSpace(input)
		if spec.boundary == End {
			if value != "" {
				return "", "", fmt.Errorf("legacy generated end marker has unexpected attributes")
			}
			return "", "", nil
		}
		if value == "" || len(strings.Fields(value)) != 1 {
			return "", "", fmt.Errorf("legacy generated marker has invalid ID")
		}
		return value, "", nil
	}
	attrs, err := parseAttributes(input)
	if err != nil {
		return "", "", err
	}
	for key := range attrs {
		if key != "id" && (spec.kind != Generated || key != "kind") {
			return "", "", fmt.Errorf("unknown %s marker attribute %q", spec.kind, key)
		}
	}
	semanticKind := attrs["kind"]
	if spec.kind == Generated && semanticKind != "" && semanticKind != "entity" && semanticKind != "activity" {
		return "", "", fmt.Errorf("generated marker requires kind entity or activity")
	}
	return attrs["id"], semanticKind, nil
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
		if keyStart == index || index >= len(input) || input[index] != '=' {
			return nil, fmt.Errorf("invalid marker attribute near byte %d", keyStart)
		}
		key := input[keyStart:index]
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
				break
			}
			index++
		}
		if index >= len(input) {
			return nil, fmt.Errorf("unterminated attribute %q", key)
		}
		index++
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
