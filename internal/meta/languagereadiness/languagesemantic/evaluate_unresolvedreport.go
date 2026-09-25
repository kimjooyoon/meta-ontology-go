package languagesemantic

func unresolvedReport(registry Registry, head, registryDigest, reason string) Report {
	results := make([]CaseResult, 0, len(registry.Cases))
	for _, definition := range registry.Cases {
		result := CaseResult{Definition: definition, Status: StatusUnresolved, Evidence: CaseEvidence{Error: reason}}
		result.Digest = caseDigest(result)
		results = append(results, result)
	}
	return buildReport(registry, results, Source{
		ExpectedHeadSHA:  head,
		ConceptID:        ConceptID,
		Producer:         "languagesemantic.Evaluate",
		Consumer:         "self-improvement-cycle",
		MetaOperation:    "prove-staged-semantic-model",
		RegistryDigest:   registryDigest,
		ObservationKnown: false,
		ConceptBound:     false,
	}, 0, expectedSources, 1)
}
