package languagesourcebindingpromotion

import "bytes"

func summarize(cases []CaseResult, input Input) Summary {
	summary := Summary{CasesTotal: 5, ProducerDependencies: input.Independence.ProducerDependencies,
		PolicyReplays: boolInt(len(input.PolicyArtifact) > 0 && bytes.Equal(input.PolicyArtifact, input.PolicyReplayArtifact))}
	for _, item := range cases {
		if item.Status == "SATISFIED" {
			summary.CasesSatisfied++
		}
		promotion := item.Claims[2]
		if item.ID == "exact-promotion" && promotion.Status == "DISCHARGED" {
			summary.ExactPromotions++
			for _, claim := range item.Claims {
				if claim.Status == "DISCHARGED" {
					summary.ExactClaims++
				}
			}
		}
		if promotion.UnknownClass == "DEPENDENCY_BLOCKED" {
			summary.DependencyBlocked++
		}
		if promotion.Status == "REFUTED" {
			summary.LinkRefutations++
		}
		for _, claim := range item.Claims[:2] {
			if claim.UnknownClass == "DIRECT_MISSING" {
				summary.DirectUnknowns++
				break
			}
		}
	}
	return summary
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
