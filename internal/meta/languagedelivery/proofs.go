package languagedelivery

func buildProofs(summary Summary, evidence string) []Proof {
	foundation := summary.SourceReceipts == summary.SourceReceiptsTotal && summary.MetaBindings == summary.MetaBindingsTotal
	coherence := summary.Coordinates.Unknown == 0 && summary.Coordinates.NotSatisfied == 0
	regression := summary.Effects.RepositoryWrites == 0 && !summary.Effects.MutationAuthority && summary.SelfMintedCredits == 0
	return []Proof{
		{Choice: ProofFoundation, Claim: "fixed contract and source artifacts are identity-bound",
			MetaOperation: "bind-delivery-foundation", EvidenceDigest: evidence, Passed: foundation},
		{Choice: ProofCoherence, Claim: "reader projections reduce the same obligation facts",
			MetaOperation: "reduce-reader-projections", EvidenceDigest: evidence, Passed: coherence},
		{Choice: ProofRegression, Claim: "unknown, self-minted, and effectful credit is rejected",
			MetaOperation: "guard-delivery-credit", EvidenceDigest: evidence, Passed: regression},
	}
}
