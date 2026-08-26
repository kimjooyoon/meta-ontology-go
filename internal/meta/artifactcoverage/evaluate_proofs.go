package artifactcoverage

func coverageProofs(report Report) []Proof {
	foundation := report.Summary.ExactHeadCoverageBasisPoints == 10000 && report.Summary.RepositoryWrites == 0
	coherence := report.Summary.CanonicalCoverageBasisPoints == 10000 && report.Summary.AmbiguousOperations == 0
	regression := report.Summary.ReplayBoundCoverageBasisPoints == 10000
	return []Proof{
		coverageProof(ProofFoundation, "bind-operation-artifact-foundation",
			"BindOperationArtifactFoundation", foundation, report.ProgramDigest, report.AuthorityDigest),
		coverageProof(ProofCoherence, "resolve-canonical-operation-artifact",
			"ResolveCanonicalOperationArtifact", coherence, report.ObservationDigest, report.ProgramDigest),
		coverageProof(ProofRegression, "replay-operation-artifact-coverage",
			"ReplayOperationArtifactCoverage", regression, report.ObservationDigest, report.ActionabilityDigest),
	}
}

func coverageProof(choice ProofChoice, operation, activity string, satisfied bool, evidence ...string) Proof {
	return Proof{Choice: choice, MetaOperation: operation, Activity: activity,
		Satisfied: satisfied, EvidenceDigest: digestJSON(append([]string{string(choice)}, evidence...))}
}
