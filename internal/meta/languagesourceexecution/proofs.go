package languagesourceexecution

func proofs(summary Summary, evidence string) []Proof {
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "execute-source-activity", EvidenceDigest: evidence,
			Passed: summary.SourceExecutions == 1 && summary.CasesSatisfied == summary.CasesTotal},
		{Choice: "COHERENCE", MetaOperation: "replay-source-execution-result", EvidenceDigest: evidence,
			Passed: summary.DeterministicReplays == 1 && summary.ExecutionEvents == 4},
		{Choice: "REGRESSION", MetaOperation: "reject-source-runtime-failure", EvidenceDigest: evidence,
			Passed: summary.DiagnosticRejections == 2 && summary.Unknowns == 0 &&
				summary.RepositoryWrites == 0 && summary.MutationAuthorities == 0},
	}
}
