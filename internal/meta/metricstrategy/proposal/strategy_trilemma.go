package proposal

import (
	"reflect"

	strategy "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

func validTrilemma(plan strategy.Plan) (bool, bool) {
	choices := []string{"FOUNDATION", "COHERENCE", "REGRESSION"}
	if !reflect.DeepEqual(plan.Policy.Choices, choices) || len(plan.Candidates) != len(choices) || len(plan.Bindings) == 0 {
		return false, false
	}
	for index, candidate := range plan.Candidates {
		if candidate.ProofChoice != choices[index] || candidate.EvidenceDigest == "" || len(candidate.MetaOperations) == 0 {
			return false, false
		}
	}
	for _, binding := range plan.Bindings {
		knownStatus := binding.Status == "SATISFIED" || binding.Status == "UNSATISFIED"
		if binding.MetaOperation == "" || binding.EvidenceDigest == "" || !knownStatus {
			return false, false
		}
	}
	for _, candidate := range plan.Candidates {
		if candidate.ProofChoice == plan.Selection.ProofChoice && candidate.EvidenceDigest == plan.Selection.CandidateDigest {
			known := map[string]bool{"REPAIR": true, "HOLD_FIXED_POINT": true, "RECONCILE": true}
			return known[plan.Selection.Decision], plan.Selection.Decision == "LOWER_RESOLUTION" || !known[plan.Selection.Decision]
		}
	}
	return false, true
}
