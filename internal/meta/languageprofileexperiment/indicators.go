package languageprofileexperiment

func buildIndicators(input Input, facts observedFacts) []Indicator {
	expectedSamples := input.Contract.Profiles * input.Contract.SamplesPerProfile
	resourceSummary := facts.wallObservations == expectedSamples && facts.wallMin > 0 &&
		facts.wallMin <= facts.wallMedian && facts.wallMedian <= facts.wallMax
	allocationSummary := facts.allocationObservations == expectedSamples && facts.allocMin > 0 &&
		facts.allocMin <= facts.allocMedian && facts.allocMedian <= facts.allocMax
	effects := facts.writes + boolInt(facts.mutation)
	return []Indicator{
		indicator("profile.receipts", "OUTCOME", "FOUNDATION", "produce-runner-profile-receipts", facts.profiles, 2),
		indicator("profile.samples", "OUTCOME", "FOUNDATION", "observe-fixed-profile-samples", facts.samples, expectedSamples),
		indicator("profile.successful-executions", "OUTCOME", "COHERENCE", "execute-profiled-language-activity", facts.successful, expectedSamples),
		indicator("profile.subject-coherence", "DRIVER", "COHERENCE", "bind-profile-subject-semantics", facts.sourceCoherence, 1),
		indicator("profile.execution-digest-variants", "DRIVER", "COHERENCE", "compare-profiled-execution-digests", facts.variants, 1),
		indicator("resource.wall-observations", "DRIVER", "FOUNDATION", "observe-runner-wall-time", facts.wallObservations, expectedSamples),
		indicator("resource.allocation-observations", "DRIVER", "FOUNDATION", "observe-runtime-total-allocation", facts.allocationObservations, expectedSamples),
		indicator("resource.wall-summary", "DRIVER", "REGRESSION", "order-wall-time-summary", boolInt(resourceSummary), 1),
		indicator("resource.allocation-summary", "DRIVER", "REGRESSION", "order-allocation-summary", boolInt(allocationSummary), 1),
		indicator("compiler.go127-executable", "DRIVER", "FOUNDATION", "bind-go127-profile-executable", boolInt(facts.executableBound && facts.go127Runtimes == 2), 1),
		indicator("counterexample.unknown-entry", "GUARDRAIL", "REGRESSION", "reject-unknown-profile-entry", facts.unknownRejections, 1),
		indicator("guardrail.effects", "GUARDRAIL", "REGRESSION", "deny-profiler-repository-effects", effects, 0),
		indicator("guardrail.non-claims", "GUARDRAIL", "FOUNDATION", "preserve-profile-non-claims", boolInt(facts.nonClaims), 1),
	}
}

func indicator(id, class, proof, operation string, observed, expected int) Indicator {
	return Indicator{ID: id, Class: class, ProofChoice: proof, MetaOperation: operation,
		Observed: observed, Expected: expected, Satisfied: observed == expected}
}
