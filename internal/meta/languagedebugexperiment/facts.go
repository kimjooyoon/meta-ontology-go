package languagedebugexperiment

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/languagedebug"
)

func collectFacts(input Input) (facts, string) {
	if !validSHA(input.SubjectSHA) || !validDigest(input.ExecutableDigest) {
		return facts{}, "DEBUG_SUBJECT_UNKNOWN"
	}
	if input.First.Decision != languagedebug.DecisionPass || input.Second.Decision != languagedebug.DecisionPass {
		return facts{Unknowns: 1}, "DEBUG_DECISION_UNKNOWN"
	}
	if languagedebug.Validate(input.First) != nil || languagedebug.Validate(input.Second) != nil ||
		languagedebug.Validate(input.UnknownBreakpoint) != nil {
		return facts{}, "DEBUG_RECEIPT_INVALID"
	}
	result := facts{DebugReceipts: 2, Go127Runtimes: 2}
	positive := []languagedebug.Receipt{input.First, input.Second}
	digests := map[string]bool{}
	for _, receipt := range positive {
		if receipt.State == languagedebug.StatePaused {
			result.PausedSessions++
		}
		if receipt.CurrentEvent != nil && receipt.CurrentEvent.Kind == receipt.Breakpoint {
			result.BreakpointsReached++
			result.CurrentEvents++
		}
		result.TraceEvents += len(receipt.Trace)
		result.RemainingEvents += receipt.RemainingEvents
		result.RepositoryWrites += receipt.Effects.RepositoryWrites
		result.MutationAuthority = result.MutationAuthority || receipt.Effects.MutationAuthority
		digests[receipt.ExecutionDigest] = true
	}
	result.ExecutionDigestVariants = len(digests)
	result.SubjectCoherence = coherence(input.First, input.Second)
	if slices.Equal(input.First.NonClaims, languagedebug.CanonicalNonClaims()) &&
		slices.Equal(input.Second.NonClaims, languagedebug.CanonicalNonClaims()) {
		result.NonClaims = len(languagedebug.CanonicalNonClaims())
	}
	if input.UnknownBreakpoint.Decision == languagedebug.DecisionFailClosed &&
		input.UnknownBreakpoint.Reason == "DEBUG_BREAKPOINT_NOT_REACHED" {
		result.UnknownBreakpointRejections = 1
	}
	return result, ""
}

func coherence(first, second languagedebug.Receipt) int {
	if first.Filename == second.Filename && first.SourceDigest == second.SourceDigest &&
		first.SemanticDigest == second.SemanticDigest && first.ExecutionDigest == second.ExecutionDigest {
		return 2
	}
	return 0
}
