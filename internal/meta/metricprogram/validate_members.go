package metricprogram

import (
	"fmt"
	"slices"
)

func validateStrategyMembers(plan StrategyPlan, verification StrategyVerification) error {
	if !slices.IsSortedFunc(plan.Bindings, func(left, right StrategyBinding) int { return compareString(left.IndicatorID, right.IndicatorID) }) {
		return fmt.Errorf("strategy bindings are not canonical")
	}
	bindings := make(map[string]StrategyBinding, len(plan.Bindings))
	for _, binding := range plan.Bindings {
		operation, ok := findOperation(binding.MetaOperation)
		if binding.IndicatorID == "" || !ok || binding.Status != "SATISFIED" || !validDigest(binding.EvidenceDigest) {
			return fmt.Errorf("indicator %q is not resolved", binding.IndicatorID)
		}
		if _, duplicate := bindings[binding.IndicatorID]; duplicate {
			return fmt.Errorf("duplicate indicator %q", binding.IndicatorID)
		}
		if operation.ProofChoice != binding.Family || expectedTrilemma(binding.Family) != binding.Trilemma {
			return fmt.Errorf("indicator %q has an invalid proof binding", binding.IndicatorID)
		}
		bindings[binding.IndicatorID] = binding
	}
	return validateCandidates(plan, verification, bindings)
}

func expectedTrilemma(choice string) string {
	switch choice {
	case "FOUNDATION":
		return "AXIOM"
	case "COHERENCE":
		return "COHERENCE"
	case "REGRESSION":
		return "REGRESS"
	default:
		return ""
	}
}

func compareString(left, right string) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
