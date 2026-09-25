package languagedebugexperiment

func buildIndicators(contract Contract, value facts) []Indicator {
	return []Indicator{
		indicator("debug.receipts", "OUTCOME", "FOUNDATION", "count-debug-receipts", value.DebugReceipts, contract.ExpectedDebugReceipts),
		indicator("debug.paused-sessions", "OUTCOME", "FOUNDATION", "count-paused-sessions", value.PausedSessions, contract.ExpectedPausedSessions),
		indicator("debug.breakpoints-reached", "OUTCOME", "FOUNDATION", "count-reached-breakpoints", value.BreakpointsReached, contract.ExpectedBreakpointsReached),
		indicator("debug.trace-events", "OPERATION", "FOUNDATION", "sum-visible-trace-events", value.TraceEvents, contract.ExpectedTraceEvents),
		indicator("debug.subject-coherence", "COHERENCE", "COHERENCE", "compare-debug-subjects", value.SubjectCoherence, contract.ExpectedSubjectCoherence),
		indicator("debug.execution-digest-variants", "REGRESSION", "REGRESSION", "count-execution-digest-variants", value.ExecutionDigestVariants, contract.ExpectedExecutionDigestVariants),
		indicator("debug.resource-observations", "OUTCOME", "REGRESSION", "observe-debug-runtime-resources", value.ResourceObservations, contract.ExpectedResourceObservations),
		indicator("debug.current-events", "OPERATION", "COHERENCE", "count-current-events", value.CurrentEvents, contract.ExpectedCurrentEvents),
		indicator("debug.remaining-events", "OPERATION", "COHERENCE", "sum-hidden-future-events", value.RemainingEvents, contract.ExpectedRemainingEvents),
		indicator("compiler.go127-executable", "FOUNDATION", "FOUNDATION", "bind-go127-debugger", value.Go127Runtimes, contract.ExpectedGo127Runtimes),
		indicator("counterexample.unknown-breakpoint", "REGRESSION", "REGRESSION", "count-unknown-breakpoint-rejections", value.UnknownBreakpointRejections, contract.ExpectedUnknownBreakpointRejections),
		indicator("guardrail.effects", "EFFECT", "FOUNDATION", "sum-debugger-effects", effectCount(value), 0),
		indicator("guardrail.non-claims", "FOUNDATION", "FOUNDATION", "count-explicit-non-claims", value.NonClaims, contract.ExpectedNonClaims),
	}
}

func indicator(id, class, proof, operation string, observed, expected int) Indicator {
	return Indicator{ID: id, Class: class, ProofChoice: proof, MetaOperation: operation,
		Observed: observed, Expected: expected, Satisfied: observed == expected}
}

func effectCount(value facts) int {
	count := value.RepositoryWrites
	if value.MutationAuthority {
		count++
	}
	return count
}
