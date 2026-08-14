package query

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

// Canonical returns a stable query projection. Candidate status is included;
// candidate explanations are intentionally omitted from semantic identity.
func (graph Graph) Canonical() string {
	var builder strings.Builder
	if graph.binding != nil && graph.binding.namespace != "" {
		builder.WriteString("namespace\t")
		writeCanonicalField(&builder, graph.binding.namespace)
		builder.WriteByte('\n')
	}
	for _, fact := range graph.AllFacts() {
		builder.WriteString("fact\t")
		writeCanonicalField(&builder, fact.Status.String())
		writeCanonicalField(&builder, fact.Subject.String())
		writeCanonicalField(&builder, fact.Predicate.String())
		writeCanonicalField(&builder, fact.Object.String())
		builder.WriteByte('\n')
	}
	return builder.String()
}

// StableHash is the digest of the canonical query projection. It is a view
// fingerprint, not semantic authority and never replaces an IR hash.
func (graph Graph) StableHash() string {
	sum := sha256.Sum256([]byte(graph.Canonical()))
	return hex.EncodeToString(sum[:])
}

func writeCanonicalField(builder *strings.Builder, value string) {
	builder.WriteString(strconv.Quote(value))
	builder.WriteByte('\t')
}
