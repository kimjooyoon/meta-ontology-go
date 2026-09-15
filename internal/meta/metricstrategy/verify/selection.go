package metricstrategyverify

import (
	"strings"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify/intervention"
	strategy "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

func replaySelection(candidates []strategy.Candidate, projections []metric.Projection) strategy.Selection {
	for _, candidate := range candidates {
		if candidate.UnsatisfiedCount > 0 {
			if replayUnresolvedConcept(candidate) {
				return makeSelection(candidate, "LOWER_RESOLUTION", "lower-semantic-resolution", "CONCEPT_OPERATION_BINDING_UNKNOWN")
			}
			return makeSelection(candidate, "REPAIR", firstOperation(candidate), "FIRST_UNSATISFIED_CANONICAL_FAMILY")
		}
	}
	if replayZeroResiduals(projections) {
		return makeSelection(findCandidate(candidates, "REGRESSION"), "HOLD_FIXED_POINT", "terminate-at-fixed-point", "ALL_INDICATORS_SATISFIED_AND_RESIDUALS_ZERO")
	}
	return makeSelection(findCandidate(candidates, "COHERENCE"), "RECONCILE", "reconcile-metric-state", "FIXED_POINT_NOT_EVIDENCED")
}

func replayUnresolvedConcept(candidate strategy.Candidate) bool {
	for _, indicatorID := range candidate.IndicatorIDs {
		if strings.HasPrefix(indicatorID, "gooo.concept.unresolved-") {
			return true
		}
	}
	return false
}

func findCandidate(candidates []strategy.Candidate, choice string) strategy.Candidate {
	for _, candidate := range candidates {
		if candidate.ProofChoice == choice {
			return candidate
		}
	}
	return strategy.Candidate{ProofChoice: choice}
}

func makeSelection(candidate strategy.Candidate, decision, operation, reason string) strategy.Selection {
	return strategy.Selection{ProofChoice: candidate.ProofChoice, Decision: decision, MetaOperation: operation, Reason: reason, CandidateDigest: candidate.EvidenceDigest, SourceMetaOperations: append([]string(nil), candidate.MetaOperations...)}
}

func firstOperation(candidate strategy.Candidate) string {
	if len(candidate.MetaOperations) == 0 {
		return "fail-closed"
	}
	return candidate.MetaOperations[0]
}

func replayZeroResiduals(projections []metric.Projection) bool {
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
