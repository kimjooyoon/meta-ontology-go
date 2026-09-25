package provenance

import (
	"fmt"
	"strings"
	"time"
)

func normalizeFreshness(freshness Freshness) (Freshness, error) {
	var err error
	freshness.ProducedAt, err = normalizeTimestamp(freshness.ProducedAt, "freshness.produced_at")
	if err != nil {
		return Freshness{}, err
	}
	if freshness.ValidUntil != "" {
		freshness.ValidUntil, err = normalizeTimestamp(freshness.ValidUntil, "freshness.valid_until")
		if err != nil {
			return Freshness{}, err
		}
		produced, _ := time.Parse(time.RFC3339Nano, freshness.ProducedAt)
		validUntil, _ := time.Parse(time.RFC3339Nano, freshness.ValidUntil)
		if !validUntil.After(produced) {
			return Freshness{}, fmt.Errorf("freshness.valid_until must be after freshness.produced_at")
		}
	}
	return freshness, nil
}
func normalizeTimestamp(value, field string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is required", field)
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return "", fmt.Errorf("%s: %w", field, err)
	}
	return parsed.UTC().Format(time.RFC3339Nano), nil
}
