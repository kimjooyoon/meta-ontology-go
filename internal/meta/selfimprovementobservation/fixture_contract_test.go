package selfimprovementobservation

func contractFixture(head, digest string) ContractReport {
	indicators := []ContractIndicator{
		{"foundation.head", "FOUNDATION", "PASS"},
		{"foundation.source", "FOUNDATION", "PASS"},
		{"foundation.semantic", "FOUNDATION", "PASS"},
		{"coherence.loop", "COHERENCE", "PASS"},
		{"coherence.executors", "COHERENCE", "PASS"},
		{"coherence.trilemma", "COHERENCE", "PASS"},
		{"coherence.read-only-language-observation", "COHERENCE", "PASS"},
		{"regression.replay", "REGRESSION", "PASS"},
	}
	return ContractReport{
		Schema: "gooo/self-improvement-contract/v1", CommitSHA: head,
		SemanticHash: digest, RegistryDigest: digest, Status: "PASS",
		Indicators: indicators,
	}
}

func reseal(report *SourceReport) {
	report.Digest = ""
	report.Digest = digestJSON(*report)
}
