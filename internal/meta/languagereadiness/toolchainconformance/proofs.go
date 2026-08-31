package toolchainconformance

func buildProofs(summary Summary, source Source) []Proof {
	foundation := source.ObservationKnown && summary.RegistryDrift == 0 &&
		summary.ConceptDrift == 0 && summary.SurfacesTotal == ExpectedSurfaceCount
	coherence := blockingCount(summary) == 0 &&
		summary.SurfacesSatisfied == ExpectedSurfaceCount &&
		summary.CasesSatisfied == ExpectedCaseCount &&
		summary.IndicatorsSatisfied == ExpectedIndicatorCount
	regression := summary.ProofsPassed == ExpectedProofCount &&
		summary.TamperRejections == ExpectedTamperCount &&
		summary.TamperTotal == ExpectedTamperCount
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "bind-fixed-surface-corpus",
			EvidenceDigest: digestValue(struct {
				Registry string
				Concept  string
			}{source.RegistryDigest, source.ConceptArtifactDigest}), Passed: foundation},
		{Choice: "COHERENCE", MetaOperation: "join-exact-head-tool-receipts",
			EvidenceDigest: digestValue(struct {
				Surfaces, Cases, Indicators, Heads int
			}{summary.SurfacesSatisfied, summary.CasesSatisfied,
				summary.IndicatorsSatisfied, summary.HeadBindings}), Passed: coherence},
		{Choice: "REGRESSION", MetaOperation: "reject-bounded-receipt-drift",
			EvidenceDigest: digestValue(struct {
				Proofs, Rejections, Writes, Authorities int
			}{summary.ProofsPassed, summary.TamperRejections,
				summary.RepositoryWrites, summary.MutationAuthorities}), Passed: regression},
	}
}
