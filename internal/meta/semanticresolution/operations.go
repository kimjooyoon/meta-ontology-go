package semanticresolution

func CanonicalOperations() []MetaOperation {
	return []MetaOperation{
		{ID: "observe-semantic-conflict", Activity: "ObserveSemanticConflict", ProofChoice: ProofFoundation},
		{ID: "lower-semantic-resolution", Activity: "LowerSemanticResolution", ProofChoice: ProofCoherence},
		{ID: "select-coarse-recovery-operation", Activity: "SelectCoarseRecoveryOperation", ProofChoice: ProofCoherence},
		{ID: "replay-resolution-transition", Activity: "ReplayResolutionTransition", ProofChoice: ProofRegression},
		{ID: "preserve-resolution-descent-bound", Activity: "PreserveResolutionDescentBound", ProofChoice: ProofFoundation},
		{ID: "preserve-read-only-resolution", Activity: "PreserveReadOnlyResolution", ProofChoice: ProofFoundation},
	}
}
