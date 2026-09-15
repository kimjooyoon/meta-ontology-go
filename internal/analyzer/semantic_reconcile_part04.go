package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func normalizedDeltaMembersValid(result SemanticAdapterResult) bool {
	memberCount := len(result.NormalizedDelta.SignatureFacts) + len(result.NormalizedDelta.CandidateFacts) +
		len(result.NormalizedDelta.DeferredFacts) +
		len(result.NormalizedDelta.DeferredImplementation) + len(result.NormalizedDelta.DeferredDetails) +
		len(result.NormalizedDelta.DeferredSlots)
	if memberCount == 0 ||
		!normalizedDeltaBindingsMatch(result) {
		return false
	}
	for _, fact := range result.NormalizedDelta.SignatureFacts {
		if !fact.Binding.complete() || fact.Fact.Validate() != nil || fact.Evidence.Validate() != nil ||
			fact.Evidence.Fact != fact.Fact.Key() || fact.Evidence.Span != fact.Fact.Span {
			return false
		}
	}
	for _, candidate := range result.NormalizedDelta.CandidateFacts {
		if !candidate.Binding.complete() || !validDigest(candidate.ObservationDigest) {
			return false
		}
		for _, fact := range candidate.Facts {
			if fact.Validate() != nil {
				return false
			}
		}
		for _, evidence := range candidate.Evidence {
			if evidence.Validate() != nil || !candidateFactKey(candidate, evidence.Fact) {
				return false
			}
		}
	}
	for _, fact := range result.NormalizedDelta.DeferredFacts {
		if !fact.Binding.complete() || !validSourceFact(fact.Fact) {
			return false
		}
	}
	for _, observation := range result.NormalizedDelta.DeferredImplementation {
		binding := DeltaBinding{
			SourceDigest: observation.SourceDigest, BaseDigest: observation.BaseDigest,
			PolicyDigest: observation.PolicyDigest, ToolchainDigest: observation.ToolchainDigest,
			RegistryDigest: observation.RegistryDigest,
		}
		if !binding.complete() || observation.Origin != OriginImplementation {
			return false
		}
	}
	for _, detail := range result.NormalizedDelta.DeferredDetails {
		if !validateDeferredImplementationDetail(detail) {
			return false
		}
	}
	for _, slot := range result.NormalizedDelta.DeferredSlots {
		if !validProtectedSlotObservation(slot) {
			return false
		}
	}
	return true
}
func candidateFactKey(candidate NormalizedCandidateFact, key semantic.FactKey) bool {
	for _, fact := range candidate.Facts {
		if fact.Key() == key && fact.Status == semantic.FactCandidate {
			return true
		}
	}
	return false
}
