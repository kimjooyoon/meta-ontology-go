package metricstrategy

import (
	"strings"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
)

func choose(candidates []Candidate, projections []metric.Projection, boundarySafe bool) Selection {
	for _, candidate := range candidates {
		if candidate.UnsatisfiedCount > 0 {
			if unresolvedConcept(candidate) {
				return selection(candidate, "LOWER_RESOLUTION", "lower-semantic-resolution", "CONCEPT_OPERATION_BINDING_UNKNOWN")
			}
			return selection(candidate, "REPAIR", firstOperation(candidate), "FIRST_UNSATISFIED_CANONICAL_FAMILY")
		}
	}
	if boundarySafe && zeroResiduals(projections) {
		return selection(candidateFor(candidates, "REGRESSION"), "HOLD_FIXED_POINT", "terminate-at-fixed-point", "ALL_INDICATORS_SATISFIED_AND_RESIDUALS_ZERO")
	}
	return selection(candidateFor(candidates, "COHERENCE"), "RECONCILE", "reconcile-metric-state", "FIXED_POINT_NOT_EVIDENCED")
}

func unresolvedConcept(candidate Candidate) bool {
	for _, indicatorID := range candidate.IndicatorIDs {
		if strings.HasPrefix(indicatorID, "gooo.concept.unresolved-") {
			return true
		}
	}
	return false
}

func candidateFor(candidates []Candidate, choice string) Candidate {
	for _, candidate := range candidates {
		if candidate.ProofChoice == choice {
			return candidate
		}
	}
	return Candidate{ProofChoice: choice}
}

func selection(candidate Candidate, decision, operation, reason string) Selection {
	return Selection{ProofChoice: candidate.ProofChoice, Decision: decision, MetaOperation: operation, Reason: reason, CandidateDigest: candidate.EvidenceDigest, SourceMetaOperations: append([]string(nil), candidate.MetaOperations...)}
}

func firstOperation(candidate Candidate) string {
	if len(candidate.MetaOperations) == 0 {
		return "fail-closed"
	}
	return candidate.MetaOperations[0]
}

func zeroResiduals(projections []metric.Projection) bool {
	if len(projections) == 0 {
		return false
	}
	for _, projection := range projections {
		if projection.Residual != 0 || projection.Status != "SATISFIED" {
			return false
		}
	}
	return true
}
