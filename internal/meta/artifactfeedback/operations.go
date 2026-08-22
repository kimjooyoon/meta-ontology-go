package artifactfeedback

func CanonicalOperations() []MetaOperation {
	return []MetaOperation{
		{ID: "observe-operation-artifact-feedback", Activity: "ObserveOperationArtifactFeedback", ProofChoice: ProofFoundation},
		{ID: "join-cycle-artifact-feedback", Activity: "JoinCycleArtifactFeedback", ProofChoice: ProofCoherence},
		{ID: "measure-next-cycle-feedback", Activity: "MeasureNextCycleFeedback", ProofChoice: ProofCoherence},
		{ID: "select-next-meta-operation", Activity: "SelectNextMetaOperation", ProofChoice: ProofCoherence},
		{ID: "replay-operation-artifact-feedback", Activity: "ReplayOperationArtifactFeedback", ProofChoice: ProofRegression},
		{ID: "preserve-read-only-feedback", Activity: "PreserveReadOnlyFeedback", ProofChoice: ProofFoundation},
	}
}

