package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func Compare(input Input) Comparison {
	oracle := Evaluate(input)
	baseline := EvaluateBaseline(input)
	comparison := Comparison{Oracle: oracle, Baseline: baseline, OutcomeMatch: oracle.Decision == baseline.Decision, ReasonMatch: oracle.Reason == baseline.Reason, LocalizationMatch: sameSurfaceSet(oracle.ChangedSurfaces, baseline.LocalizedSurfaces)}
	if oracle.Decision == DecisionPass {
		comparison.LocalizationMatch = len(baseline.LocalizedSurfaces) == 0
	}
	if comparison.OutcomeMatch && comparison.ReasonMatch && comparison.LocalizationMatch {
		comparison.Finding = "NO_UNIQUE_BENEFIT"
	} else {
		comparison.Finding = "UNIQUE_BENEFIT_NOT_ESTABLISHED"
	}
	return comparison
}

type baselineSemanticView struct {
	digest string
	facts  []string
}
type baselineRegistryView struct {
	bySurface map[string]CodeBinding
	bySymbol  map[string]CodeBinding
	digest    string
}

func baselineCounts(input Input) ObservationCounts {
	counts := ObservationCounts{RegistryBindings: uint64(len(input.Registry)), ReceiptRecords: uint64(len(input.Receipts)), PathEdges: uint64(len(input.Path.Edges)), PathClaims: uint64(len(input.Path.Claims)), PathEvidence: uint64(len(input.Path.Evidence)), ResourceReceipts: uint64(len(input.ResourceReceipts))}
	for _, change := range input.Changes {
		if change.BeforeDigest != change.AfterDigest {
			counts.ChangedCodeSurfaces++
		}
	}
	for _, edge := range input.Path.Edges {
		if edge.Kind == semantic.InferenceObservationCandidate {
			counts.CandidateObservations++
		}
		if edge.Kind == semantic.InferenceAcceptedLift {
			counts.AcceptedLifts++
		}
	}
	return counts
}
