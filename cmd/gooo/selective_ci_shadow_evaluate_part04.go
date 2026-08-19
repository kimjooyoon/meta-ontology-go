package main

import (
	"errors"
	analyzersci "github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"strings"
)

func shadowDecodeReason(err error) string {
	var snapshotErr *analyzersci.Error
	if errors.As(err, &snapshotErr) {
		return string(snapshotErr.Code)
	}
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "duplicate"):
		return "DUPLICATE_FIELD"
	case strings.Contains(message, "unknown field") || strings.Contains(message, "unknown"):
		return "UNKNOWN_FIELD"
	case strings.Contains(message, "trailing") || strings.Contains(message, "multiple"):
		return "TRAILING_DATA"
	case strings.Contains(message, "stale") || strings.Contains(message, "mismatch"):
		return "STALE_OR_MISMATCHED"
	default:
		return "MALFORMED"
	}
}
