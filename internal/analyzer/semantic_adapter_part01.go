package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// SemanticAdapterInput supplies an analyzer result and an immutable semantic
// base. Base is normalized into a private transaction before any additions.
type SemanticAdapterInput struct {
	Base             semantic.IR
	Analysis         Result
	Policy           MappingPolicy
	Producer         semantic.ID
	EvidenceKind     semantic.EvidenceKind
	SourceDigest     string
	ToolchainDigest  string
	Registry         *Registry
	SlotObservations []ProtectedSlotObservation
}

// SemanticAdapterResult keeps unmapped observations and implementation detail
// outside the authoritative semantic graph. Candidates are added only through
// Graph.AddCandidate when their explicit mapping and endpoints are valid.
type SemanticAdapterResult struct {
	IR                              semantic.IR
	SourceDigest                    string
	PolicyDigest                    string
	ToolchainDigest                 string
	RegistryDigest                  string
	BindingDigest                   string
	Locality                        LocalityEnvelope
	ImplementationObservationDigest string
	SlotObservationDigest           string
	NormalizedDelta                 SemanticNormalizedDelta
	DeferredFacts                   []Fact
	DeferredCandidates              []Candidate
	// ShadowedCandidateEvidence retains a mapped candidate observation when
	// the same FactKey is already deterministic in the base graph. The
	// evidence is intentionally not added to IR: candidate evidence cannot
	// stand in for authoritative evidence, and the semantic IR rejects a
	// candidate evidence record without a candidate fact. Keeping it here
	// preserves the historical observation without changing authority.
	ShadowedCandidateEvidence  []semantic.Evidence
	ImplementationDetails      []ImplementationDetail
	ImplementationObservations []ImplementationObservation
	SlotObservations           []ProtectedSlotObservation
}
