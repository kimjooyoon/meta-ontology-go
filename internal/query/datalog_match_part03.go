package query

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Canonical provides a stable receipt for replay and permutation tests.
func (result DatalogResult) Canonical() string {
	var builder strings.Builder
	for _, row := range result.Rows {
		builder.WriteString(datalogRowCanonical(row))
		builder.WriteByte('\n')
	}
	for _, fact := range result.Derived {
		builder.WriteString(datalogFactCanonical(fact))
		builder.WriteByte('\n')
	}
	if result.Complete {
		builder.WriteString("complete\n")
	} else {
		builder.WriteString("incomplete\n")
	}
	return builder.String()
}
func (result DatalogResult) StableHash() string {
	digest := sha256.Sum256([]byte(result.Canonical()))
	return hex.EncodeToString(digest[:])
}
