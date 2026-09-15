package impactgraph

import (
	"fmt"
	"strings"
	"unicode"
)

func edgeAlias(primary, first, second, field string) (string, error) {
	values := []string{primary, first, second}
	chosen := ""
	for _, value := range values {
		if value == "" {
			continue
		}
		if chosen != "" && chosen != value {
			return "", fmt.Errorf("conflicting %s aliases", field)
		}
		chosen = value
	}
	return chosen, nil
}
func validNodeKind(kind NodeKind) bool {
	switch kind {
	case NodeKindSource, NodeKindSemantic, NodeKindGoSymbol, NodeKindGoPackage,
		NodeKindGeneratedRegion, NodeKindObligation, NodeKindPressure:
		return true
	default:
		return false
	}
}
func validateID(value string) error {
	if value == "" || strings.TrimSpace(value) != value || strings.IndexFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}) >= 0 {
		return fmt.Errorf("stable ID must be non-empty and contain no whitespace or control characters")
	}
	return nil
}
func validDigest(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}
