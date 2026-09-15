package languagepackageexecution

func indicators(summary Summary) []Indicator {
	return []Indicator{
		metric("PACKAGE_FIXED_CASES", "OUTCOME", "reduce-fixed-package-cases", summary.CasesSatisfied, 5, summary.CasesSatisfied == 5),
		metric("PACKAGE_SOURCE_FILES", "OPERATION", "count-bound-source-files", summary.SourceFilesLoaded, 2, summary.SourceFilesLoaded == 2),
		metric("PACKAGE_EXECUTIONS", "OUTCOME", "count-package-executions", summary.PackageExecutions, 1, summary.PackageExecutions == 1),
		metric("PACKAGE_DETERMINISTIC_REPLAYS", "REGRESSION", "compare-receipt-digests", summary.DeterministicReplays, 1, summary.DeterministicReplays == 1),
		metric("PACKAGE_DIAGNOSTIC_REJECTIONS", "COHERENCE", "count-expected-rejections", summary.DiagnosticRejections, 3, summary.DiagnosticRejections == 3),
		metric("PACKAGE_EVENTS", "OPERATION", "count-observable-events", summary.EventsObserved, 7, summary.EventsObserved == 7),
		metric("PACKAGE_UNKNOWN_DECISIONS", "FOUNDATION", "reject-unknown-top-decisions", summary.UnknownDecisions, 0, summary.UnknownDecisions == 0),
		metric("PACKAGE_REPOSITORY_WRITES", "EFFECT", "sum-repository-writes", summary.RepositoryWrites, 0, summary.RepositoryWrites == 0),
		metric("PACKAGE_MUTATION_AUTHORITIES", "EFFECT", "sum-mutation-authorities", summary.MutationAuthorities, 0, summary.MutationAuthorities == 0),
	}
}

func metric(id, class, operation string, value, target int, satisfied bool) Indicator {
	return Indicator{ID: id, Class: class, MetaOperation: operation, Value: value, Target: target, Satisfied: satisfied}
}

func proofs(report Report) []Proof {
	foundation := report.Summary.CasesTotal == 5 && report.Summary.UnknownDecisions == 0
	coherence := report.Summary.DiagnosticRejections == 3 && report.Summary.RepositoryWrites == 0 && report.Summary.MutationAuthorities == 0
	regression := report.Summary.DeterministicReplays == 1
	return []Proof{
		{Choice: "FOUNDATION", Status: proofStatus(foundation), EvidenceDigest: report.FactsDigest},
		{Choice: "COHERENCE", Status: proofStatus(coherence), EvidenceDigest: report.FactsDigest},
		{Choice: "REGRESSION", Status: proofStatus(regression), EvidenceDigest: report.FactsDigest},
	}
}

func proofStatus(value bool) string {
	if value {
		return "PASS"
	}
	return "FAIL_CLOSED"
}
