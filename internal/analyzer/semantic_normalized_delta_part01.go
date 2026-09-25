package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

const semanticNormalizedDeltaSchema = "analyzer-semantic-delta/v1"

// DeltaBinding is the common identity tuple for one source-backed handoff.
// SourceDigest covers the exact generated source bytes; the other fields pin
// the semantic base and the producer contract used to interpret them.
type DeltaBinding struct {
	SourceDigest    string `json:"source_digest"`
	BaseDigest      string `json:"base_digest"`
	PolicyDigest    string `json:"policy_digest"`
	ToolchainDigest string `json:"toolchain_digest"`
	RegistryDigest  string `json:"registry_digest"`
}

func (b DeltaBinding) canonical() string {
	var builder strings.Builder
	writeBindingField(&builder, b.SourceDigest)
	writeBindingField(&builder, b.BaseDigest)
	writeBindingField(&builder, b.PolicyDigest)
	writeBindingField(&builder, b.ToolchainDigest)
	writeBindingField(&builder, b.RegistryDigest)
	return builder.String()
}
func (b DeltaBinding) complete() bool {
	return validDigest(b.SourceDigest) && validDigest(b.BaseDigest) &&
		validDigest(b.PolicyDigest) && validDigest(b.ToolchainDigest) &&
		validDigest(b.RegistryDigest)
}

// NormalizedSignatureFact is an authoritative, typed signature fact and its
// semantic evidence. It is the only delta member eligible for IR authority.
type NormalizedSignatureFact struct {
	Binding        DeltaBinding      `json:"binding"`
	SourceRelation Relation          `json:"source_relation"`
	Fact           semantic.Fact     `json:"fact"`
	Evidence       semantic.Evidence `json:"evidence"`
}

func (f NormalizedSignatureFact) canonical() string {
	var builder strings.Builder
	builder.WriteString("signature\n")
	builder.WriteString(f.Binding.canonical())
	writeBindingField(&builder, string(f.SourceRelation))
	builder.WriteString(f.Fact.Canonical())
	builder.WriteString(f.Evidence.Canonical())
	return builder.String()
}

// NormalizedCandidateFact records an unresolved source relation. Facts and
// evidence are populated only when an explicit policy mapped the options;
// the candidate remains separate from deterministic graph facts either way.
type NormalizedCandidateFact struct {
	Binding           DeltaBinding        `json:"binding"`
	ObservationDigest string              `json:"observation_digest"`
	SourceRelation    Relation            `json:"source_relation"`
	Origin            ObservationOrigin   `json:"origin"`
	Subject           semantic.ID         `json:"subject"`
	Options           []semantic.ID       `json:"options"`
	Facts             []semantic.Fact     `json:"facts"`
	Evidence          []semantic.Evidence `json:"evidence"`
	Span              semantic.Span       `json:"span"`
	Reason            string              `json:"reason"`
}
type NormalizedDeferredFact struct {
	Binding DeltaBinding `json:"binding"`
	Fact    Fact         `json:"fact"`
}
