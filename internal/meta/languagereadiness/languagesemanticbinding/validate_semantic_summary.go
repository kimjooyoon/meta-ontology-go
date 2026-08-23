package languagesemanticbinding

func validateSemanticSummary(summary semanticSummary) error {
	counts := summary.Satisfied == 18 && summary.Total == 18 && summary.Executed == 18
	counts = counts && summary.NotSatisfied == 0 && summary.Unresolved == 0 && summary.ReadinessBPS == 10000
	models := summary.SourceModels == 13 && summary.NormalizedIRs == 13 && summary.SemanticReplays == 13
	models = models && summary.ProvenanceReplays == 13 && summary.EvidenceReplays == 13
	laws := summary.PresentationLaws == 1 && summary.CandidateAuthorityLaws == 1
	laws = laws && summary.DeterministicAuthorityLaws == 1 && summary.UpstreamRejections == 2
	guards := summary.UnregisteredGooo == 0 && summary.MissingRegistered == 0
	guards = guards && summary.StageOrderViolations == 0 && summary.EffectfulStages == 0 && summary.RegistryDrift == 0
	return require(counts && models && laws && guards, "semantic summary mismatch")
}
