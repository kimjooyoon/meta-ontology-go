package toolchainformatfix

func summarize(source Source, cases []CaseResult, cycle cycleEvidence, registryDrift int) Summary {
	summary := Summary{Total: FixedTotal, RegistryDrift: registryDrift,
		InMemoryApplications: cycle.applications, FixedPoints: cycle.fixedPoints,
		DirectWrites: cycle.directWrites}
	if source.ObservationKnown && validDigest(source.ExecutableDigest) {
		summary.BinaryBindings = 1
	}
	for _, result := range cases {
		summary.Invocations += result.Invocations
		summary.StructuredOutputs += result.StructuredOutput
		summary.StructuredPlans += result.StructuredPlan
		summary.RepositoryWrites += result.RepositoryWrites
		if result.Invocations == 2 {
			summary.Executed++
		}
		if result.Status == "SATISFIED" {
			summary.Satisfied++
			if result.Definition.Kind == "POSITIVE" {
				summary.PositivePaths++
			} else {
				summary.GuardrailRejections++
			}
		} else if result.Status == "UNRESOLVED" {
			summary.Unresolved++
		}
		if !result.ExitMatched && result.Invocations > 0 {
			summary.ExitMismatches++
		}
		if !result.OutputMatched && result.Invocations > 0 {
			summary.OutputMismatches++
		}
		if !result.ReplayMatched && result.Invocations > 0 {
			summary.ReplayMismatches++
		} else if result.ReplayMatched {
			summary.ReplayMatches++
		}
	}
	summary.ReadinessBPS = summary.Satisfied * 10000 / FixedTotal
	return summary
}
