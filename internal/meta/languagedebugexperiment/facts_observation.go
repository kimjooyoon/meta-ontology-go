package languagedebugexperiment

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/languagedebug"
)

func observedFacts(input Input) facts {
	result := facts{DebugReceipts: 2, Go127Runtimes: len(input.RuntimeObservations), ReplayMatches: 1, ResourceObservations: len(input.RuntimeObservations)}
	positive := []languagedebug.Receipt{input.First, input.Second}
	digests := map[string]bool{}
	for _, receipt := range positive {
		if receipt.State == languagedebug.StatePaused { result.PausedSessions++ }
		if receipt.CurrentEvent != nil && receipt.CurrentEvent.Kind == receipt.Breakpoint { result.BreakpointsReached++; result.CurrentEvents++ }
		result.TraceEvents += len(receipt.Trace); result.RemainingEvents += receipt.RemainingEvents; result.RepositoryWrites += receipt.Effects.RepositoryWrites
		result.MutationAuthority = result.MutationAuthority || receipt.Effects.MutationAuthority; digests[receipt.ExecutionDigest] = true
	}
	result.ExecutionDigestVariants = len(digests)
	return result
}

func canonicalNonClaims(first, second languagedebug.Receipt) bool {
	claims := languagedebug.CanonicalNonClaims()
	return slices.Equal(first.NonClaims, claims) && slices.Equal(second.NonClaims, claims)
}

func unknownBreakpointObserved(receipt languagedebug.Receipt) bool {
	return receipt.Decision == languagedebug.DecisionFailClosed && receipt.Reason == "DEBUG_BREAKPOINT_NOT_REACHED"
}
