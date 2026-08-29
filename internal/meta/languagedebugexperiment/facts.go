package languagedebugexperiment

import "github.com/kimjooyoon/meta-ontology-go/internal/languagedebug"

func collectFacts(input Input) (facts, string) {
	value, reason := validateInput(input)
	if reason != "" {
		return value, reason
	}
	value = observedFacts(input)
	value.SubjectCoherence = coherence(input.First, input.Second)
	if canonicalNonClaims(input.First, input.Second) {
		value.NonClaims = len(languagedebug.CanonicalNonClaims())
	}
	if unknownBreakpointObserved(input.UnknownBreakpoint) {
		value.UnknownBreakpointRejections = 1
	}
	return value, ""
}

func coherence(first, second languagedebug.Receipt) int {
	if first.Filename == second.Filename && first.SourceDigest == second.SourceDigest &&
		first.SemanticDigest == second.SemanticDigest && first.ExecutionDigest == second.ExecutionDigest {
		return 2
	}
	return 0
}
