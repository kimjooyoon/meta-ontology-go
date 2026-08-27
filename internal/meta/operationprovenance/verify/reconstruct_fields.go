package verify

import (
	"fmt"
	"strings"
)

func parseComputed(value string) (map[string]string, string, error) {
	parts := strings.Split(value, "|")
	if len(parts) < 2 || (parts[0] != "metric" && parts[0] != "scenario") {
		return nil, "", fmt.Errorf("unsupported computes value")
	}
	fields := make(map[string]string, len(parts)-1)
	for _, part := range parts[1:] {
		key, raw, ok := strings.Cut(part, "=")
		if !ok || key == "" {
			return nil, "", fmt.Errorf("malformed computes field %q", part)
		}
		if _, exists := fields[key]; exists {
			return nil, "", fmt.Errorf("duplicate computes field %q", key)
		}
		fields[key] = raw
	}
	return fields, parts[0], nil
}
