package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

const localityEnvelopeSchema = "analyzer-locality/v1"

// LocalityEnvelope mirrors bidir.Locality's Touched/Affected vocabulary. The
// preserved fact set is analyzer-owned compatibility data for the future
// bidirectional adapter; it prevents partial observations from deleting base.
type LocalityEnvelope struct {
	SchemaVersion  string             `json:"schema_version"`
	BaseDigest     string             `json:"base_digest"`
	Touched        []semantic.ID      `json:"touched"`
	Affected       []semantic.ID      `json:"affected"`
	PreservedFacts []semantic.FactKey `json:"preserved_facts"`
	Digest         string             `json:"digest"`
}

// Canonical is order-independent and excludes Digest so it can be hashed.
func (e LocalityEnvelope) Canonical() string {
	var builder strings.Builder
	builder.WriteString(localityEnvelopeSchema)
	builder.WriteByte('\n')
	writeBindingField(&builder, e.SchemaVersion)
	writeBindingField(&builder, e.BaseDigest)
	for _, id := range sortedLocalityIDs(e.Touched) {
		writeBindingField(&builder, "touched")
		writeBindingField(&builder, id.String())
	}
	for _, id := range sortedLocalityIDs(e.Affected) {
		writeBindingField(&builder, "affected")
		writeBindingField(&builder, id.String())
	}
	for _, key := range sortedLocalityFacts(e.PreservedFacts) {
		writeBindingField(&builder, "preserved")
		writeBindingField(&builder, key.Subject.String())
		writeBindingField(&builder, key.Predicate.String())
		writeBindingField(&builder, key.Object.String())
	}
	return builder.String()
}

// StableHash is the deterministic digest of the locality closure envelope.
func (e LocalityEnvelope) StableHash() string {
	return semantic.StableHashString(e.Canonical())
}
