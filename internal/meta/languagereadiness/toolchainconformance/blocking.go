package toolchainconformance

func blockingCount(summary Summary) int {
	return summary.MissingSurfaces + summary.UnexpectedSurfaces +
		summary.SchemaMismatches + summary.HeadMismatches +
		summary.DecisionMismatches + summary.ResolutionDescents +
		summary.CaseMismatches + summary.IndicatorFailures +
		summary.ProofFailures + summary.Unresolved + summary.DigestFailures +
		summary.RegistryDrift + summary.ConceptDrift +
		summary.RepositoryWrites + summary.MutationAuthorities
}

func mergeAuthority(summary *Summary, counts conceptCounts) {
	summary.ConceptBindings = counts.ConceptBindings
	summary.CodeBindings = counts.CodeBindings
	summary.MetricBindings = counts.MetricBindings
	summary.UseCaseBindings = counts.UseCaseBindings
}
