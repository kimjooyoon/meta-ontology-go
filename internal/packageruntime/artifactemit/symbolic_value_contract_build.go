package artifactemit

func buildSymbolicValueContract(input symbolicValueArtifactInput, subjectSHA string, acceptVector, rejectVector *symbolicValueVectorInput) SymbolicValueContract {
	indicators := newSymbolicValueContractIndicators(input)
	contract := SymbolicValueContract{
		Schema:               symbolicValueContractSchema,
		SubjectSHA:           subjectSHA,
		MetricID:             "gooo.metric.compiler.symbolic-value-contract.v1",
		Decision:             "PASS",
		Resolution:           "VALUE_CONTRACT_ONLY",
		Reason:               "SYMBOLIC_VALUE_SEMANTICS_COMPILED",
		SourceArtifactDigest: input.Digest,
		Rules:                newSymbolicValueContractRules(acceptVector, rejectVector),
		Default: SymbolicValueContractDefault{
			Decision:      "FAIL_CLOSED",
			Resolution:    "LOWER_RESOLUTION",
			Reason:        "SYMBOLIC_INVOCATION_VALUE_UNMATCHED",
			ProofChoice:   "REGRESSION",
			MetaOperation: "fail-closed-unmatched-symbolic-value",
		},
		Indicators: indicators,
		Effects: SymbolicValueContractEffects{
			RepositoryWrites:  0,
			MutationAuthority: false,
		},
		PromotionCreditBPS: 0,
		NotClaimed: []string{
			"default-policy external coverage",
			"effect execution",
			"arbitrary user input",
			"complete interpreter semantics",
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
	if contract.Coordinates.Satisfied != contract.Coordinates.Total {
		contract.Decision = "FAIL_CLOSED"
		contract.Resolution = "INVARIANT_ONLY"
		contract.Reason = "SYMBOLIC_VALUE_CONTRACT_INCOMPLETE"
	}
	return contract
}
