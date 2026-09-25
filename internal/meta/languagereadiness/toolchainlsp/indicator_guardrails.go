package toolchainlsp

func guardrailIndicators(s Summary, resolution string) []Indicator {
	values := []struct {
		name, proof string
		value       int
	}{
		{"missing-cases.guardrail.v1", "FOUNDATION", s.MissingCases},
		{"unexpected-cases.guardrail.v1", "FOUNDATION", s.UnexpectedCases},
		{"case-failures.guardrail.v1", "COHERENCE", s.CaseFailures},
		{"capability-gaps.guardrail.v1", "FOUNDATION", s.CapabilityGaps},
		{"unexpected-protocol-errors.guardrail.v1", "COHERENCE", s.UnexpectedProtocolErrors},
		{"diagnostic-gaps.guardrail.v1", "REGRESSION", s.DiagnosticGaps},
		{"nonstandard-wire-fields.guardrail.v1", "COHERENCE", s.NonstandardWireFields},
		{"stale-navigation-leaks.guardrail.v1", "REGRESSION", s.StaleNavigationLeaks},
		{"unknown-navigation-leaks.guardrail.v1", "REGRESSION", s.UnknownNavigationLeaks},
		{"fail-closed-navigation-leaks.guardrail.v1", "REGRESSION", s.FailClosedNavigationLeaks},
		{"unresolved.guardrail.v1", "COHERENCE", s.Unresolved},
		{"digest-failures.guardrail.v1", "REGRESSION", s.DigestFailures},
		{"corpus-drift.guardrail.v1", "FOUNDATION", s.CorpusDrift},
		{"concept-drift.guardrail.v1", "FOUNDATION", s.ConceptDrift},
		{"head-mismatches.guardrail.v1", "COHERENCE", s.HeadMismatches},
		{"proof-failures.guardrail.v1", "REGRESSION", s.ProofFailures},
		{"repository-writes.guardrail.v1", "REGRESSION", s.RepositoryWrites},
		{"mutation-authorities.guardrail.v1", "REGRESSION", s.MutationAuthorities},
	}
	result := make([]Indicator, 0, len(values))
	for _, value := range values {
		result = append(result, indicator(value.name, "GUARDRAIL", value.proof,
			value.value, 0, "less_or_equal", resolution))
	}
	return result
}
