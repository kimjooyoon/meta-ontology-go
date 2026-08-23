package languagesemanticbinding

func validateSemanticSummary(summary semanticSummary) error {
	counts := summary.Satisfied == 19 && summary.Total == 19 && summary.Executed == 19
	counts = counts && summary.NotSatisfied == 0 && summary.Unresolved == 0 && summary.ReadinessBPS == 10000
	models := summary.SourceModels == 14 && summary.NormalizedIRs == 14 && summary.SemanticReplays == 14
	models = models && summary.ProvenanceReplays == 14 && summary.EvidenceReplays == 14
	laws := summary.PresentationLaws == 1 && summary.CandidateAuthorityLaws == 1
	laws = laws && summary.DeterministicAuthorityLaws == 1 && summary.UpstreamRejections == 2
	guards := summary.UnregisteredGooo == 0 && summary.MissingRegistered == 0
	guards = guards && summary.StageOrderViolations == 0 && summary.EffectfulStages == 0 && summary.RegistryDrift == 0
	return require(counts && models && laws && guards, "semantic summary mismatch")
}
