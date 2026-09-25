package artifactcoverage

func CanonicalOperations() []MetaOperation {
	return []MetaOperation{
		{ID: "bind-operation-artifact-foundation", Activity: "BindOperationArtifactFoundation",
			ProofChoice: ProofFoundation},
		{ID: "resolve-canonical-operation-artifact", Activity: "ResolveCanonicalOperationArtifact",
			ProofChoice: ProofCoherence},
		{ID: "measure-operation-artifact-coverage", Activity: "MeasureOperationArtifactCoverage",
			ProofChoice: ProofCoherence},
		{ID: "preserve-read-only-artifact-observation", Activity: "PreserveReadOnlyArtifactObservation",
			ProofChoice: ProofFoundation},
		{ID: "select-uncovered-operation", Activity: "SelectUncoveredOperation",
			ProofChoice: ProofCoherence},
		{ID: "replay-operation-artifact-coverage", Activity: "ReplayOperationArtifactCoverage",
			ProofChoice: ProofRegression},
	}
}
