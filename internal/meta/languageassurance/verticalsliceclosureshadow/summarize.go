package verticalsliceclosureshadow

func baselineSummary() Summary {
	return Summary{DenominatorTotal: officialTotal, BeforeOperating: beforeOperating,
		ProjectedOperating: beforeOperating, BeforeCoverageBPS: beforeCoverageBPS,
		ProjectedCoverageBPS: beforeCoverageBPS, BoundariesTotal: boundaryTotal,
		LinksTotal: linkTotal}
}

func summarize(boundaries []BoundaryResult) Summary {
	summary := baselineSummary()
	for _, result := range boundaries {
		if result.EvidenceAvailable {
			summary.EvidenceAvailable++
		}
		switch result.Status {
		case StatusSatisfied:
			summary.BoundariesSatisfied++
		case StatusUnknown:
			summary.UnknownBoundaries++
		case StatusBlocked:
			summary.BlockedBoundaries++
		}
		summary.LinksSatisfied += result.LinksSatisfied
		summary.ObservedRepositoryWrites += result.RepositoryWrites
		if result.UnknownTopDecision {
			summary.UnknownTopDecisions++
		}
		if result.KnownFailure {
			summary.KnownFailures++
		}
	}
	summary.BoundaryCoverageBPS = summary.BoundariesSatisfied * 10000 / boundaryTotal
	summary.LinkCoverageBPS = summary.LinksSatisfied * 10000 / linkTotal
	return summary
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
