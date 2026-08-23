package languagediagnosticprovenance

func proofs(summary Summary, registryDigest string) []Proof {
	foundation := summary.RegistryDrift == 0 && summary.ConceptDrift == 0 &&
		summary.ToolchainMatches == 1 && summary.ConceptBindings == 1 &&
		summary.CodeBindings == 8 && summary.MetricBindings == 18 &&
		summary.UseCaseBindings == 3 && summary.OrderedDiagnostics == 6
	coherence := summary.Traced == 10 && summary.PhysicalPositions == 10 &&
		summary.LogicalPositions == 10 && summary.SemanticBindings == 4 &&
		summary.LSPProjections == 10 && summary.CanonicalReplays == 10 &&
		summary.LineDirectiveRemaps == 1 && summary.TypeClassifications == 3 &&
		summary.ProvenanceSteps == 50
	regression := summary.GuardrailRejections == 8 &&
		summary.NotSatisfied == 0 && summary.Unresolved == 0 &&
		summary.UnknownAcceptances == 0 && summary.MissingMapAccepts == 0 &&
		summary.AmbiguousAccepts == 0 && summary.InvalidAcceptances == 0 &&
		summary.EffectfulStages == 0
	return []Proof{
		{
			Choice: "FOUNDATION",
			MetaOperation: "bind-versioned-diagnostic-provenance-registry",
			EvidenceDigest: digestJSON(struct {
				Registry string
				Summary  Summary
			}{registryDigest, summary}),
			Passed: foundation,
		},
		{
			Choice: "COHERENCE",
			MetaOperation: "trace-physical-logical-semantic-and-lsp-coordinates",
			EvidenceDigest: digestJSON(struct{ Traced, Semantic, Replayed int }{
				summary.Traced, summary.SemanticBindings, summary.CanonicalReplays}),
			Passed: coherence,
		},
		{
			Choice: "REGRESSION",
			MetaOperation: "reject-unknown-missing-ambiguous-and-invalid-provenance",
			EvidenceDigest: digestJSON(struct{ Rejected, Unknown, Missing, Ambiguous, Invalid int }{
				summary.GuardrailRejections, summary.UnknownAcceptances,
				summary.MissingMapAccepts, summary.AmbiguousAccepts,
				summary.InvalidAcceptances}),
			Passed: regression,
		},
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
