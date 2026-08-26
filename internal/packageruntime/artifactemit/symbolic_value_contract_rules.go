package artifactemit

func newSymbolicValueContractRules(acceptVector, rejectVector *symbolicValueVectorInput) []SymbolicValueContractRule {
	return []SymbolicValueContractRule{
		{
			ID:            "complete-symbolic-invocation",
			Match:         SymbolicValueContractRuleMatch{Activity: "NON_EMPTY", Inputs: "NON_EMPTY"},
			Decision:      "READY",
			Resolution:    "VALUE_EXACT",
			Reason:        "SYMBOLIC_INVOCATION_VALUE_PROJECTED",
			ProofChoice:   acceptVector.ProofChoice,
			MetaOperation: acceptVector.MetaOperation,
		},
		{
			ID:            "missing-activity",
			Match:         SymbolicValueContractRuleMatch{Activity: "MISSING_OR_EMPTY", Inputs: "ANY"},
			Decision:      "FAIL_CLOSED",
			Resolution:    "LOWER_RESOLUTION",
			Reason:        "SYMBOLIC_INVOCATION_VALUE_INCOMPLETE",
			ProofChoice:   rejectVector.ProofChoice,
			MetaOperation: rejectVector.MetaOperation,
		},
	}
}
