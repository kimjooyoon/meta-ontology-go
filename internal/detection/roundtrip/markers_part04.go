package roundtrip

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

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
