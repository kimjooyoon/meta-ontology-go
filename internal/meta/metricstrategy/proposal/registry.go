package proposal

var registry = []CoordinateSpec{
	{ID: "exact-strategy-subject", Class: "DRIVER", ProofChoice: "FOUNDATION", MetaOperation: "bind-exact-strategy-subject"},
	{ID: "verified-strategy-receipt", Class: "DRIVER", ProofChoice: "FOUNDATION", MetaOperation: "verify-strategy-receipt"},
	{ID: "deterministic-strategy-replay", Class: "GUARDRAIL", ProofChoice: "REGRESSION", MetaOperation: "replay-strategy-plan"},
	{ID: "concept-governed-trilemma", Class: "DRIVER", ProofChoice: "COHERENCE", MetaOperation: "bind-concept-trilemma-selection"},
	{ID: "actionable-generation-plan", Class: "OUTCOME", ProofChoice: "COHERENCE", MetaOperation: "propose-independent-meta-operations"},
	{ID: "independent-action-groups", Class: "DRIVER", ProofChoice: "REGRESSION", MetaOperation: "select-independent-action-groups"},
	{ID: "executable-conformance-obligations", Class: "DRIVER", ProofChoice: "FOUNDATION", MetaOperation: "bind-executor-evaluator-receipts"},
	{ID: "read-only-non-authorizing-boundary", Class: "GUARDRAIL", ProofChoice: "FOUNDATION", MetaOperation: "preserve-proposal-boundary"},
}

func Registry() []CoordinateSpec {
	return append([]CoordinateSpec(nil), registry...)
}
