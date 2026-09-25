package verticalsliceclosureactivation

func validEligibilityIndicators(indicators []eligibilityIndicator) bool {
	classes, proofs := map[string]int{}, map[string]int{}
	for _, item := range indicators {
		if !item.Satisfied || item.Resolution != ResolutionExact ||
			item.Producer != "verticalsliceclosureeligibility.Evaluate" ||
			item.Consumer != "language-assurance-activation-gate" || item.MetaOperation == "" {
			return false
		}
		classes[item.Class]++
		proofs[item.ProofChoice]++
	}
	return len(indicators) == 8 && classes["OUTCOME"] == 1 && classes["DRIVER"] == 4 && classes["GUARDRAIL"] == 3 &&
		proofs["FOUNDATION"] == 4 && proofs["COHERENCE"] == 3 && proofs["REGRESSION"] == 1
}

func validEligibilityMetaOperations(operations []eligibilityOperation) bool {
	proofs := map[string]int{}
	for _, operation := range operations {
		if operation.ID == "" {
			return false
		}
		proofs[operation.ProofChoice]++
	}
	return len(operations) == 6 && proofs["FOUNDATION"] == 3 && proofs["COHERENCE"] == 2 && proofs["REGRESSION"] == 1
}

func countSatisfied(indicators []eligibilityIndicator) int {
	count := 0
	for _, indicator := range indicators {
		if indicator.Satisfied {
			count++
		}
	}
	return count
}
