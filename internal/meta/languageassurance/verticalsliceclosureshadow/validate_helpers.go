package verticalsliceclosureshadow

func allBoundariesExact(boundaries []BoundaryResult) bool {
	for _, result := range boundaries {
		if result.Status != StatusSatisfied || result.Resolution != ResolutionExact ||
			result.Value != result.Target || result.LinksSatisfied != result.LinksTotal ||
			!result.EvidenceAvailable || result.UnknownTopDecision ||
			result.KnownFailure || result.RepositoryWrites != 0 {
			return false
		}
	}
	return true
}

func allIndicatorsSatisfied(indicators []Indicator) bool {
	for _, item := range indicators {
		if !item.Satisfied {
			return false
		}
	}
	return true
}

func validIndicatorShape(indicators []Indicator) bool {
	classes := map[string]int{}
	proofs := map[string]int{}
	for _, item := range indicators {
		if item.MetaOperation != MetaOperation ||
			item.Producer != "verticalsliceclosureshadow.Evaluate" ||
			item.Consumer != "language-assurance-promotion-gate" {
			return false
		}
		classes[item.Class]++
		proofs[item.ProofChoice]++
	}
	return classes["OUTCOME"] == 1 && classes["DRIVER"] == 2 &&
		classes["GUARDRAIL"] == 3 && proofs["FOUNDATION"] == 3 &&
		proofs["COHERENCE"] == 2 && proofs["REGRESSION"] == 1
}
