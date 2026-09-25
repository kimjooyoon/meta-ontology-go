package artifactemit

func buildSymbolicValueReachability(artifact Artifact, contract SymbolicValueContract, subjectSHA string, analysis symbolicValueReachabilityAnalysis) SymbolicValueReachability {
	indicators := newSymbolicValueReachabilityIndicators(artifact, contract, subjectSHA, analysis)
	reachability := SymbolicValueReachability{
		Schema:     symbolicValueReachabilitySchema,
		SubjectSHA: subjectSHA,
		MetricID:   "gooo.metric.compiler.symbolic-value-reachability.v1",
		Decision:   "PASS",
		Resolution: "SCHEMA_VALUE_REACHABILITY_ONLY",
		Reason:     "SYMBOLIC_VALUE_REACHABILITY_COMPILED",
		Source: SymbolicValueReachabilitySource{
			ArtifactDigest: artifact.Digest,
			ContractDigest: contract.Digest,
		},
		Summary:    analysis.Summary,
		Rules:      analysis.Rules,
		Default:    analysis.Default,
		Indicators: indicators,
		Effects: SymbolicValueContractEffects{
			RepositoryWrites:  0,
			MutationAuthority: false,
		},
		PromotionCreditBPS: 0,
		NotClaimed: []string{
			"generic JSON Schema-to-value-contract entailment",
			"runtime branch frequency",
			"external user-path execution",
			"domain correctness",
			"production readiness",
		},
		Coordinates: symbolicValueCoordinates(indicators),
		Classes:     symbolicValueClasses(indicators),
		Views: []SymbolicValueContractView{
			symbolicValueView(indicators, "USER", "USER_VISIBLE"),
			symbolicValueView(indicators, "TOOL_AUTHOR", "TOOL_CONTRACT"),
			symbolicValueView(indicators, "GOVERNOR", "FULL_RECEIPT"),
		},
		Proofs: symbolicValueProofs(indicators),
	}
	if reachability.Coordinates.Satisfied != reachability.Coordinates.Total {
		reachability.Decision = "FAIL_CLOSED"
		reachability.Resolution = "INVARIANT_ONLY"
		reachability.Reason = "SYMBOLIC_VALUE_REACHABILITY_INCOMPLETE"
	}
	return reachability
}
