package domaincapability

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

// ScopeDigest identifies the declared capability boundary, not its observed quality.
func ScopeDigest(expectedCapabilityIDs []string) string {
	ids := normalizeIDs(expectedCapabilityIDs)
	var builder strings.Builder
	for _, id := range ids {
		builder.WriteString(strconv.Itoa(len(id)))
		builder.WriteByte(':')
		builder.WriteString(id)
		builder.WriteByte(0)
	}
	sum := sha256.Sum256([]byte(builder.String()))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// ScopeChanged prevents comparisons across different declared domain boundaries.
func ScopeChanged(previousDigest, currentDigest string) bool {
	return strings.TrimSpace(previousDigest) != strings.TrimSpace(currentDigest)
}