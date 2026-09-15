package languagesemanticbinding

func validateSemanticSummary(summary semanticSummary) error {
	counts := summary.Satisfied == semanticCaseDenominator && summary.Total == semanticCaseDenominator && summary.Executed == semanticCaseDenominator
	counts = counts && summary.NotSatisfied == 0 && summary.Unresolved == 0 && summary.ReadinessBPS == 10000
	models := summary.SourceModels == semanticSourceDenominator && summary.NormalizedIRs == semanticSourceDenominator && summary.SemanticReplays == semanticSourceDenominator
	models = models && summary.ProvenanceReplays == semanticSourceDenominator && summary.EvidenceReplays == semanticSourceDenominator
	laws := summary.PresentationLaws == 1 && summary.CandidateAuthorityLaws == 1
	laws = laws && summary.DeterministicAuthorityLaws == 1 && summary.UpstreamRejections == 2
	guards := summary.UnregisteredGooo == 0 && summary.MissingRegistered == 0
	guards = guards && summary.StageOrderViolations == 0 && summary.EffectfulStages == 0 && summary.RegistryDrift == 0
	return require(counts && models && laws && guards, "semantic summary mismatch")
}
