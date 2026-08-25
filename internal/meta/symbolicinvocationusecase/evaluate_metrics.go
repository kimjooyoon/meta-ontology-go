package symbolicinvocationusecase

func buildIndicators(contract Contract, value facts) []Indicator {
	mutationAuthorities := 0
	if value.Effects.MutationAuthority {
		mutationAuthorities = 1
	}
	return []Indicator{
		indicator("user.validation-decisions", "OUTCOME", "COHERENCE", "sum-external-user-decisions", value.UserDecisions, contract.ExpectedAcceptedInstances+contract.ExpectedRejectedInstances),
		indicator("user.accepted-instances", "DRIVER", "FOUNDATION", "count-externally-accepted-instances", value.AcceptedInstances, contract.ExpectedAcceptedInstances),
		indicator("user.rejected-instances", "DRIVER", "REGRESSION", "count-externally-rejected-instances", value.RejectedInstances, contract.ExpectedRejectedInstances),
		indicator("guardrail.deterministic-replays", "GUARDRAIL", "FOUNDATION", "count-producer-replays", value.DeterministicReplays, contract.ExpectedDeterministicReplays),
		indicator("guardrail.repository-writes", "GUARDRAIL", "FOUNDATION", "sum-cross-boundary-writes", value.Effects.RepositoryWrites, contract.ExpectedRepositoryWrites),
		indicator("guardrail.mutation-authorities", "GUARDRAIL", "COHERENCE", "join-cross-boundary-authority", mutationAuthorities, contract.ExpectedMutationAuthorities),
	}
}

func indicator(id, class, proof, operation string, observed, expected int) Indicator {
	return Indicator{ID: id, Class: class, ProofChoice: proof, MetaOperation: operation,
		Observed: observed, Expected: expected, Satisfied: observed == expected}
}

func countIndicators(indicators []Indicator) Counter {
	result := Counter{Total: len(indicators)}
	for _, indicator := range indicators {
		if indicator.Satisfied {
			result.Satisfied++
		}
	}
	if result.Total > 0 {
		result.BasisPoints = result.Satisfied * 10000 / result.Total
	}
	return result
}

func buildViews(indicators []Indicator) []View {
	return []View{
		buildView("USER", "USER_VISIBLE", indicators[:3]),
		buildView("TOOL_AUTHOR", "TOOL_CONTRACT", indicators[:4]),
		buildView("GOVERNOR", "FULL_RECEIPT", indicators),
	}
}

func buildView(audience, resolution string, indicators []Indicator) View {
	view := View{Audience: audience, Resolution: resolution, Total: len(indicators)}
	for _, indicator := range indicators {
		view.IndicatorIDs = append(view.IndicatorIDs, indicator.ID)
		if indicator.Satisfied {
			view.Satisfied++
		}
	}
	if view.Total > 0 {
		view.BasisPoints = view.Satisfied * 10000 / view.Total
	}
	return view
}
