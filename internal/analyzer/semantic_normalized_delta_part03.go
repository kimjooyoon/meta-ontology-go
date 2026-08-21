package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
	"strings"
)

// Canonical returns an order-independent representation of the typed delta.
func (d SemanticNormalizedDelta) Canonical() string {
	var builder strings.Builder
	builder.WriteString(d.SchemaVersion)
	builder.WriteByte('\n')
	signatures := append([]NormalizedSignatureFact(nil), d.SignatureFacts...)
	sort.Slice(signatures, func(i, j int) bool { return signatures[i].canonical() < signatures[j].canonical() })
	for _, fact := range signatures {
		builder.WriteString(fact.canonical())
	}
	candidates := append([]NormalizedCandidateFact(nil), d.CandidateFacts...)
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].canonical() < candidates[j].canonical() })
	for _, candidate := range candidates {
		builder.WriteString(candidate.canonical())
	}
	deferredFacts := append([]NormalizedDeferredFact(nil), d.DeferredFacts...)
	sort.Slice(deferredFacts, func(i, j int) bool { return deferredFacts[i].canonical() < deferredFacts[j].canonical() })
	for _, fact := range deferredFacts {
		builder.WriteString(fact.canonical())
	}
	observations := append([]ImplementationObservation(nil), d.DeferredImplementation...)
	sort.Slice(observations, func(i, j int) bool { return observations[i].Canonical() < observations[j].Canonical() })
	for _, observation := range observations {
		builder.WriteString("deferred\n")
		builder.WriteString(observation.Canonical())
	}
	details := append([]DeferredImplementationDetail(nil), d.DeferredDetails...)
	sort.Slice(details, func(i, j int) bool { return details[i].canonical() < details[j].canonical() })
	for _, detail := range details {
		builder.WriteString(detail.canonical())
	}
	slots := append([]ProtectedSlotObservation(nil), d.DeferredSlots...)
	sort.Slice(slots, func(i, j int) bool { return slots[i].Canonical() < slots[j].Canonical() })
	for _, slot := range slots {
		builder.WriteString("slot\n")
		builder.WriteString(slot.Canonical())
	}
	return builder.String()
}

// StableHash is the digest used to compare normalized deltas across runs.
func (d SemanticNormalizedDelta) StableHash() string {
	return semantic.StableHashString(d.Canonical())
}
func newSemanticNormalizedDelta(
	input SemanticAdapterInput, baseDigest string, result SemanticAdapterResult,
) (SemanticNormalizedDelta, error) {
	delta := SemanticNormalizedDelta{SchemaVersion: semanticNormalizedDeltaSchema}
	binding := DeltaBinding{
		SourceDigest: input.SourceDigest, BaseDigest: baseDigest,
		PolicyDigest: input.Policy.Digest(), ToolchainDigest: input.ToolchainDigest,
		RegistryDigest: result.RegistryDigest,
	}
	delta.SignatureFacts = normalizedSignatureFacts(input, result, binding)
	var err error
	delta.CandidateFacts, err = normalizedCandidateFacts(input, result, binding)
	if err != nil {
		return SemanticNormalizedDelta{}, err
	}
	delta.DeferredFacts = normalizedDeferredFacts(result.DeferredFacts, binding)
	delta.DeferredImplementation = append([]ImplementationObservation(nil), result.ImplementationObservations...)
	delta.DeferredDetails = deferredImplementationDetails(result, binding)
	delta.DeferredSlots = append([]ProtectedSlotObservation(nil), result.SlotObservations...)
	delta.Digest = delta.StableHash()
	return delta, validateDeltaShape(delta)
}
