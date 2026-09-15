package toolchainlsp

func buildProofs(summary Summary, corpusDigest, conceptDigest string) []Proof {
	foundation := summary.CapabilityGaps == 0 && summary.CorpusDrift == 0 && summary.ConceptDrift == 0 && summary.HeadMismatches == 0
	coherence := summary.CaseFailures == 0 && summary.Unresolved == 0 && summary.NonstandardWireFields == 0
	regression := summary.StaleNavigationLeaks == 0 && summary.UnknownNavigationLeaks == 0 && summary.FailClosedNavigationLeaks == 0 && summary.RepositoryWrites == 0
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "bind-lsp-3.18-subset-corpus", EvidenceDigest: digestValue([]string{corpusDigest, conceptDigest}), Passed: foundation},
		{Choice: "COHERENCE", MetaOperation: "project-parser-state-and-coupling-evidence", EvidenceDigest: digestValue(summary), Passed: coherence},
		{Choice: "REGRESSION", MetaOperation: "withhold-navigation-under-uncertainty", EvidenceDigest: digestValue([]int{summary.FailClosedPaths, summary.StaleNavigationLeaks, summary.UnknownNavigationLeaks, summary.FailClosedNavigationLeaks}), Passed: regression},
	}
}
