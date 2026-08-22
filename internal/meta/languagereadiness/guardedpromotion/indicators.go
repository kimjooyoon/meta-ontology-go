package guardedpromotion

func Indicators(source Source, summary Summary, coordinates []Coordinate) []Indicator {
	sourceBPS := coordinateBPS(coordinates[:8])
	boundaryDebt := coordinateDebt(coordinates[8:10])
	return []Indicator{
		indicator("promotion-readiness-bps", "OUTCOME", "COHERENCE",
			summary.ReadinessBPS, 10000),
		indicator("valid-predecessors", "DRIVER", "FOUNDATION",
			source.ValidCandidates, 1),
		indicator("source-integrity-bps", "DRIVER", "COHERENCE",
			sourceBPS, 10000),
		indicator("unmerged-boundary-debt", "GUARDRAIL", "REGRESSION",
			boundaryDebt, 0),
		indicator("ambiguous-predecessors", "GUARDRAIL", "REGRESSION",
			source.AmbiguousCandidates, 0),
		indicator("unresolved-evidence", "GUARDRAIL", "FOUNDATION",
			summary.Unresolved, 0),
		indicator("observer-writes", "GUARDRAIL", "FOUNDATION",
			source.RepositoryWrites, 0),
		indicator("mutation-authority", "GUARDRAIL", "FOUNDATION",
			boolInt(source.RepositoryMutationAuthorized), 0),
	}
}

func indicator(id, class, choice string, value, target int) Indicator {
	return Indicator{
		MetricID: "gooo.metric.language.autonomy-guarded-promotion-" + id + ".v1",
		Class: class, ProofChoice: choice,
		Producer: "internal/meta/languagereadiness/guardedpromotion",
		Consumer: "language-readiness", MetaOperation: "guard-readiness-promotion",
		Value: value, Target: target, Satisfied: value == target,
	}
}

func coordinateBPS(coordinates []Coordinate) int {
	return (len(coordinates) - coordinateDebt(coordinates)) * 10000 / len(coordinates)
}

func coordinateDebt(coordinates []Coordinate) int {
	debt := 0
	for _, coordinate := range coordinates {
		if coordinate.Status != statusSatisfied {
			debt++
		}
	}
	return debt
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
