package verticalsliceclosureshadow

func observeBoundary(id string, artifact artifactEnvelope) (int, string) {
	summary := artifact.Summary
	switch id {
	case "syntax":
		return summary.Satisfied, status(summary.Satisfied == 20 && summary.Total == 20 &&
			summary.Executed == 20 && summary.NotSatisfied == 0 && summary.Unresolved == 0 &&
			summary.ReadinessBPS == 10000)
	case "semantics":
		return summary.Satisfied, status(summary.Satisfied == 22 && summary.Total == 22 &&
			summary.Executed == 22 && summary.NotSatisfied == 0 && summary.Unresolved == 0 &&
			summary.ReadinessBPS == 10000 && summary.StageOrderViolations == 0 &&
			summary.EffectfulStages == 0 && summary.RegistryDrift == 0)
	case "binding":
		return summary.BoundCoordinates, status(summary.Coordinates == 12 &&
			summary.BoundCoordinates == 12 && summary.Unresolved == 0 &&
			summary.ReadinessCompleted == 21 && summary.ReadinessTotal == 24 &&
			summary.ReadinessBPS == 8750 && summary.SemanticSatisfied == 22 &&
			summary.SemanticTotal == 22 && summary.EffectfulStages == 0 &&
			summary.MutationAuthorities == 0)
	case "use-cases":
		return summary.Satisfied, status(summary.Satisfied == 3 && summary.Total == 3 &&
			summary.Executed == 3 && summary.NotSatisfied == 0 &&
			summary.Unresolved == 0 && summary.ReadinessBPS == 10000)
	case "toolchain":
		return summary.CasesSatisfied, status(summary.SurfacesSatisfied == 9 &&
			summary.SurfacesTotal == 9 && len(artifact.Surfaces) == 9 &&
			summary.CasesSatisfied == 165 && summary.CasesTotal == 165 &&
			summary.ExecutedCases == 165 && summary.CaseReadinessBPS == 10000 &&
			summary.IndicatorsSatisfied == summary.IndicatorsTotal &&
			summary.ProofsPassed == summary.ProofsTotal &&
			summary.TamperRejections == 13 && summary.TamperTotal == 13)
	case "release":
		return summary.CasesSatisfied, status(summary.CasesSatisfied == 20 &&
			summary.CasesTotal == 20 && len(artifact.Cases) == 20 &&
			summary.ReadinessBPS == 10000 && summary.PlatformReceipts == 3 &&
			summary.OperatingSystems == 3 && allReleaseCasesExact(artifact.Cases))
	default:
		return 0, StatusBlocked
	}
}

func status(satisfied bool) string {
	if satisfied {
		return StatusSatisfied
	}
	return StatusBlocked
}

func allReleaseCasesExact(cases []artifactCase) bool {
	for _, item := range cases {
		if item.Observed == "" || item.Observed != item.Expected {
			return false
		}
	}
	return true
}
