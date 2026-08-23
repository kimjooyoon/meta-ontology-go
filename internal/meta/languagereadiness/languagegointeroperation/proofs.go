package languagegointeroperation

func proofs(summary Summary, registryDigest string) []Proof {
	foundation := summary.RegistryDrift == 0 && summary.ToolchainMatches == 1 &&
		summary.ConceptBindings == 1 && summary.CodeBindings == 6 &&
		summary.MetricBindings == 18 && summary.UseCaseBindings == 3
	coherence := summary.GeneratorProjections == 8 && summary.Go127Boundaries == 8 &&
		summary.CanonicalReplays == 16 && summary.TypeIdentityReplays == 16 &&
		summary.SourceMaps == 8 && summary.GenericMethods == 5 && summary.AliasNodes == 2
	regression := summary.GuardrailRejections == 8 && summary.NotSatisfied == 0 &&
		summary.Unresolved == 0 && summary.InvalidAcceptances == 0 &&
		summary.UnknownAcceptances == 0 && summary.ImportAcceptances == 0 && summary.EffectfulStages == 0
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "bind-versioned-go-interoperation-registry",
			EvidenceDigest: digestJSON(struct{ Registry string; Summary Summary }{registryDigest, summary}), Passed: foundation},
		{Choice: "COHERENCE", MetaOperation: "reify-project-normalize-and-replay-go-api",
			EvidenceDigest: digestJSON(struct{ Generator, Go127, Replay, Identity int }{8, 8, summary.CanonicalReplays, summary.TypeIdentityReplays}), Passed: coherence},
		{Choice: "REGRESSION", MetaOperation: "reject-invalid-unknown-and-ambient-authority",
			EvidenceDigest: digestJSON(struct{ Rejected, Invalid, Unknown, Imports int }{summary.GuardrailRejections, summary.InvalidAcceptances, summary.UnknownAcceptances, summary.ImportAcceptances}), Passed: regression},
	}
}

func allProofsPassed(proofs []Proof) bool {
	if len(proofs) != 3 {
		return false
	}
	for _, proof := range proofs {
		if !proof.Passed {
			return false
		}
	}
	return true
}
