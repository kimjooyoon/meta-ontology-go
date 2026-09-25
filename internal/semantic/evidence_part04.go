package semantic

import (
	"sort"
	"strings"
)

// ComparisonCanonical excludes host and source-location metadata. It is the
// cross-host provenance claim used to compare Go-hosted and gooo-hosted runs.
func (e Evidence) ComparisonCanonical() string {
	if normalized, err := e.Normalized(); err == nil {
		e = normalized
	}
	var b strings.Builder
	b.WriteString("evidence-comparison\t")
	writeCanonicalField(&b, e.ID.String())
	writeCanonicalField(&b, e.Kind.String())
	writeCanonicalField(&b, e.Status.String())
	writeCanonicalField(&b, e.Fact.Subject.String())
	writeCanonicalField(&b, e.Fact.Predicate.String())
	writeCanonicalField(&b, e.Fact.Object.String())
	writeCanonicalField(&b, e.Digest)
	return b.String()
}
func (e Evidence) StableHash() string {
	return StableHashString(e.Canonical())
}
func (e Evidence) ProvenanceHash() string {
	return StableHashString(e.ComparisonCanonical())
}
func sortEvidence(evidence []Evidence) {
	sort.Slice(evidence, func(i, j int) bool {
		if evidence[i].ID != evidence[j].ID {
			return evidence[i].ID < evidence[j].ID
		}
		return evidence[i].Canonical() < evidence[j].Canonical()
	})
}
