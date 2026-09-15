package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
	case !result.RegistryMatch:
		return "registry-mismatch"
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
