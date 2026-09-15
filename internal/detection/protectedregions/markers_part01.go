package protectedregions

import (
	"fmt"
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
