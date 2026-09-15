package artifactemit

func symbolicReaderRequestClasses(checks symbolicReaderRequestChecks) []SymbolicReaderRequestClass {
	return []SymbolicReaderRequestClass{
		{
			Class: "OUTCOME", Total: 3,
			Satisfied: symbolicReaderRequestCount(
				checks.ResolutionMatches, checks.IndicatorsSelected, checks.OutputSourceBound,
			),
		},
		{
			Class: "DRIVER", Total: 4,
			Satisfied: symbolicReaderRequestCount(
				checks.AudienceKnown, checks.ResolutionKnown,
				checks.ReaderPresent, checks.ReaderCountBound,
			),
		},
		{
			Class: "GUARDRAIL", Total: 5,
			Satisfied: symbolicReaderRequestCount(
				checks.SyntaxValid, checks.SingleProjectionOperation,
				checks.SourceContractValid, checks.SourceReadOnly,
				checks.SourceIndicatorsUnique,
			),
		},
	}
}

func symbolicReaderRequestProofs(checks symbolicReaderRequestChecks) []SymbolicReaderRequestProof {
	return []SymbolicReaderRequestProof{
		{
			ProofChoice: "FOUNDATION", Total: 5,
			Satisfied: symbolicReaderRequestCount(
				checks.SyntaxValid, checks.SingleProjectionOperation,
				checks.SourceContractValid, checks.AudienceKnown, checks.ResolutionKnown,
			),
		},
		{
			ProofChoice: "COHERENCE", Total: 4,
			Satisfied: symbolicReaderRequestCount(
				checks.ReaderPresent, checks.ReaderCountBound,
				checks.ResolutionMatches, checks.IndicatorsSelected,
			),
		},
		{
			ProofChoice: "REGRESSION", Total: 3,
			Satisfied: symbolicReaderRequestCount(
				checks.SourceReadOnly, checks.SourceIndicatorsUnique, checks.OutputSourceBound,
			),
		},
	}
}
