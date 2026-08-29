package verify

var operations = []operationSpec{
	{ID: "bind-exact-source-metrics", Activity: "BindExactSourceMetrics", ProofChoice: "FOUNDATION", InputEntity: "SourceMetrics", OutputEntity: "BoundMetrics", Mode: "BIND", Ordinal: 10},
	{ID: "exempt-project-root-topology", Activity: "ExemptProjectRootTopology", ProofChoice: "FOUNDATION", InputEntity: "BoundMetrics", OutputEntity: "RootPolicy", Mode: "EXEMPT", Ordinal: 20},
	{ID: "interpret-dimension-registry", Activity: "InterpretDimensionRegistry", ProofChoice: "FOUNDATION", InputEntity: "RootPolicy", OutputEntity: "DimensionRegistry", Mode: "INTERPRET", Ordinal: 30},
	{ID: "project-algebraic-root-state", Activity: "ProjectAlgebraicRootState", ProofChoice: "COHERENCE", InputEntity: "SourceMetrics", OutputEntity: "ProjectedMetrics", Mode: "PROJECT", Ordinal: 40},
	{ID: "observe-counterfactual-boundary", Activity: "ObserveCounterfactualBoundary", ProofChoice: "REGRESSION", InputEntity: "ProjectedMetrics", OutputEntity: "CounterfactualBoundary", Mode: "OBSERVE", Ordinal: 50},
	{ID: "preserve-repository-workspace", Activity: "PreserveRepositoryWorkspace", ProofChoice: "REGRESSION", InputEntity: "CounterfactualBoundary", OutputEntity: "WorkspaceReceipt", Mode: "ASSERT_NO_WRITE", Ordinal: 60},
	{ID: "replay-counterfactual", Activity: "ReplayCounterfactual", ProofChoice: "REGRESSION", InputEntity: "WorkspaceReceipt", OutputEntity: "VerificationEvidence", Mode: "REPLAY", Ordinal: 70},
	{ID: "terminate-at-fixed-point", Activity: "TerminateAtFixedPoint", ProofChoice: "REGRESSION", InputEntity: "VerificationEvidence", OutputEntity: "FixedPointReceipt", Mode: "TERMINATE", Ordinal: 80},
	{ID: "preserve-non-promoting-terminal", Activity: "PreserveNonPromotingTerminal", ProofChoice: "REGRESSION", InputEntity: "VerificationEvidence", OutputEntity: "FixedPointReceipt", Mode: "PRESERVE_NON_PROMOTING", Ordinal: 90},
}

func findOperation(id string) (operationSpec, bool) {
	for _, operation := range operations {
		if operation.ID == id {
			return operation, true
		}
	}
	return operationSpec{}, false
}
