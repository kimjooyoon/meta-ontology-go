package analyzer

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

// ReconcileWriteEffect is explicit because reconciliation is read-only. A
// rejected identity, binding, or observation can never authorize a write.
type ReconcileWriteEffect string

const ReconcileNoWrite ReconcileWriteEffect = "no-write"

// SemanticReconcileResult is a read-only comparison of an adapted source view
// with an independently declared semantic contract.
type SemanticReconcileResult struct {
	Comparison       semantic.IRComparison
	DeltaValid       bool
	AuthoritySafe    bool
	SourceMatch      bool
	PolicyMatch      bool
	ToolchainMatch   bool
	ObservationMatch bool
	Accepted         bool
	WriteEffect      ReconcileWriteEffect
	FailureCode      string
}

// ReconcileSemantic never mutates either IR. Acceptance requires semantic
// equality and all source/policy/toolchain bindings plus provenance equality;
// absent external provenance therefore remains fail-closed.
func ReconcileSemantic(
	observed SemanticAdapterResult, expected semantic.IR, sourceDigest, policyDigest,
	toolchainDigest, observationDigest string,
) SemanticReconcileResult {
	comparison := semantic.CompareIR(observed.IR, expected)
	deltaValid := normalizedDeltaValid(observed)
	authoritySafe := normalizedDeltaAuthoritySafe(observed)
	result := SemanticReconcileResult{
		Comparison:       comparison,
		DeltaValid:       deltaValid,
		AuthoritySafe:    authoritySafe,
		SourceMatch:      observed.SourceDigest == sourceDigest,
		PolicyMatch:      observed.PolicyDigest == policyDigest,
		ToolchainMatch:   observed.ToolchainDigest == toolchainDigest,
		ObservationMatch: observed.ImplementationObservationDigest == observationDigest && observationDigest != "",
		WriteEffect:      ReconcileNoWrite,
	}
	result.Accepted = deltaValid && authoritySafe && comparison.SemanticEqual && comparison.ProvenanceEqual &&
		result.SourceMatch && result.PolicyMatch && result.ToolchainMatch && result.ObservationMatch
	if !result.Accepted {
		result.FailureCode = reconcileFailureCode(result)
	}
	return result
}

func normalizedDeltaValid(result SemanticAdapterResult) bool {
	if result.NormalizedDelta.Digest == "" || result.NormalizedDelta.Digest != result.NormalizedDelta.StableHash() {
		return false
	}
	if result.SlotObservationDigest != protectedSlotObservationDigest(result.SlotObservations) ||
		result.SlotObservationDigest != protectedSlotObservationDigest(result.NormalizedDelta.DeferredSlots) ||
		len(result.SlotObservations) != len(result.NormalizedDelta.DeferredSlots) {
		return false
	}
	if result.ImplementationObservationDigest != implementationObservationDigest(
		result.ImplementationObservations, result.SlotObservations,
	) {
		return false
	}
	if err := validateDeltaShape(result.NormalizedDelta); err != nil {
		return false
	}
	memberCount := len(result.NormalizedDelta.SignatureFacts) + len(result.NormalizedDelta.CandidateFacts) +
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
		if !candidate.Binding.complete() {
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
	for _, observation := range result.NormalizedDelta.DeferredImplementation {
		binding := DeltaBinding{
			SourceDigest: observation.SourceDigest, BaseDigest: observation.BaseDigest,
			PolicyDigest: observation.PolicyDigest, ToolchainDigest: observation.ToolchainDigest,
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

func normalizedDeltaBindingsMatch(result SemanticAdapterResult) bool {
	var binding *DeltaBinding
	accept := func(candidate DeltaBinding) bool {
		if !candidate.complete() || candidate.SourceDigest != result.SourceDigest ||
			candidate.PolicyDigest != result.PolicyDigest || candidate.ToolchainDigest != result.ToolchainDigest {
			return false
		}
		if binding == nil {
			copyOf := candidate
			binding = &copyOf
			return true
		}
		return *binding == candidate
	}
	for _, fact := range result.NormalizedDelta.SignatureFacts {
		if !accept(fact.Binding) {
			return false
		}
	}
	for _, candidate := range result.NormalizedDelta.CandidateFacts {
		if !accept(candidate.Binding) {
			return false
		}
	}
	for _, observation := range result.NormalizedDelta.DeferredImplementation {
		if !validDigest(observation.SourceDigest) || observation.SourceDigest != result.SourceDigest ||
			observation.PolicyDigest != result.PolicyDigest || observation.ToolchainDigest != result.ToolchainDigest {
			return false
		}
		if binding == nil {
			binding = &DeltaBinding{
				SourceDigest: observation.SourceDigest, BaseDigest: observation.BaseDigest,
				PolicyDigest: observation.PolicyDigest, ToolchainDigest: observation.ToolchainDigest,
			}
		} else if binding.BaseDigest != observation.BaseDigest {
			return false
		}
	}
	for _, detail := range result.NormalizedDelta.DeferredDetails {
		if !accept(detail.Binding) {
			return false
		}
	}
	for _, slot := range result.NormalizedDelta.DeferredSlots {
		if !accept(DeltaBinding{
			SourceDigest: slot.SourceDigest, BaseDigest: slot.BaseDigest,
			PolicyDigest: slot.PolicyDigest, ToolchainDigest: slot.ToolchainDigest,
		}) {
			return false
		}
	}
	return binding != nil
}

func normalizedDeltaAuthoritySafe(result SemanticAdapterResult) bool {
	for _, fact := range result.NormalizedDelta.SignatureFacts {
		if !result.IR.Graph.HasFact(fact.Fact.Key()) {
			return false
		}
	}
	for _, candidate := range result.NormalizedDelta.CandidateFacts {
		for _, fact := range candidate.Facts {
			if result.IR.Graph.HasFact(fact.Key()) && !shadowedCandidateEvidenceMatches(result, fact.Key()) {
				return false
			}
			if !result.IR.Graph.HasCandidate(fact.Key()) && !shadowedCandidateEvidenceMatches(result, fact.Key()) {
				return false
			}
		}
	}
	return true
}

func shadowedCandidateEvidenceMatches(result SemanticAdapterResult, key semantic.FactKey) bool {
	for _, evidence := range result.ShadowedCandidateEvidence {
		if evidence.Fact == key && evidence.Status == semantic.FactCandidate {
			return true
		}
	}
	return false
}

func reconcileFailureCode(result SemanticReconcileResult) string {
	switch {
	case !result.DeltaValid:
		return "invalid-delta-binding"
	case !result.AuthoritySafe:
		return "candidate-or-deferred-promotion"
	case !result.Comparison.LeftValid || !result.Comparison.RightValid:
		return "invalid-semantic-ir"
	case !result.SourceMatch || !result.ObservationMatch:
		return "source-or-observation-mismatch"
	case !result.PolicyMatch || !result.ToolchainMatch:
		return "policy-or-toolchain-mismatch"
	case !result.Comparison.SemanticEqual:
		return "identity-mismatch"
	case !result.Comparison.ProvenanceEqual:
		return "provenance-mismatch"
	default:
		return "reconcile-rejected"
	}
}
