package toolchainconformance

func guardrailIndicators(summary Summary) []Indicator {
	return []Indicator{
		guardrail("missing-surfaces", "FOUNDATION", summary.MissingSurfaces),
		guardrail("unexpected-surfaces", "FOUNDATION", summary.UnexpectedSurfaces),
		guardrail("schema-mismatches", "FOUNDATION", summary.SchemaMismatches),
		guardrail("head-mismatches", "COHERENCE", summary.HeadMismatches),
		guardrail("decision-mismatches", "COHERENCE", summary.DecisionMismatches),
		guardrail("resolution-descents", "COHERENCE", summary.ResolutionDescents),
		guardrail("case-mismatches", "COHERENCE", summary.CaseMismatches),
		guardrail("indicator-failures", "COHERENCE", summary.IndicatorFailures),
		guardrail("proof-failures", "REGRESSION", summary.ProofFailures),
		guardrail("unresolved", "COHERENCE", summary.Unresolved),
		guardrail("digest-failures", "REGRESSION", summary.DigestFailures),
		guardrail("registry-drift", "FOUNDATION", summary.RegistryDrift),
		guardrail("concept-drift", "FOUNDATION", summary.ConceptDrift),
		guardrail("repository-writes", "REGRESSION", summary.RepositoryWrites),
		guardrail("mutation-authorities", "REGRESSION", summary.MutationAuthorities),
	}
}

func guardrail(name, proof string, value int) Indicator {
	return metric("gooo.metric.toolchain.conformance-"+name+".guardrail.v1",
		"GUARDRAIL", proof, value, 0, "less_or_equal")
}
