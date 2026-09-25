package main

func countClaimDependencyKinds(report *claimDependencyReport) {
	for _, edge := range report.Edges {
		switch edge.Kind {
		case "REQUIRES":
			report.KindCounts.Requires++
		case "SUPPORTS":
			report.KindCounts.Supports++
		case "CONTRADICTS":
			report.KindCounts.Contradicts++
		case "FAILURE_ENTAILMENT":
			report.KindCounts.FailureEntailment++
		}
	}
}
