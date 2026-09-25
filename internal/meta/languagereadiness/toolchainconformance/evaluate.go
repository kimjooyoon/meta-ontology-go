package toolchainconformance

func Evaluate(input Input) Report {
	corpus, registryDigest, registryErr := parseCorpus(input.RegistryRaw)
	if registryErr != nil {
		corpus = Corpus{Schema: CorpusSchema, Surfaces: fixedSurfaces,
			TamperCases: fixedTamperCases}
	}
	counts, conceptErr := inspectConcept(input.ConceptArtifact)
	summary, surfaces := inspectAll(corpus.Surfaces, input.Artifacts,
		input.ExpectedHeadSHA)
	if registryErr != nil {
		summary.RegistryDrift++
	}
	if conceptErr != nil {
		summary.ConceptDrift++
	} else {
		mergeAuthority(&summary, counts)
	}
	if !validHead(input.ExpectedHeadSHA) {
		summary.HeadMismatches++
	}
	cases := []CaseResult{canonicalCase(summary, input.Artifacts)}
	for _, tamper := range corpus.TamperCases {
		result := evaluateTamper(corpus.Surfaces, input, tamper)
		cases = append(cases, result)
		summary.TamperTotal++
		if result.Status == "SATISFIED" {
			summary.TamperRejections++
		}
	}
	report := Report{Schema: Schema, Source: Source{
		ExpectedHeadSHA: input.ExpectedHeadSHA, RegistryDigest: registryDigest,
		ConceptArtifactDigest: counts.ArtifactDigest, CatalogDigest: counts.CatalogDigest,
		ObservationKnown: registryErr == nil && conceptErr == nil &&
			validHead(input.ExpectedHeadSHA),
	}, Summary: summary, Surfaces: surfaces, Cases: cases}
	if blockingCount(summary) == 0 &&
		summary.TamperRejections == summary.TamperTotal {
		report.Decision, report.Resolution = DecisionPass, ResolutionExact
		report.ReasonCode = "ALL_TOOLCHAIN_SURFACES_CONFORM"
	} else {
		report.Decision, report.Resolution = DecisionFailClosed, ResolutionLower
		report.ReasonCode = "TOOLCHAIN_CONFORMANCE_UNKNOWN"
	}
	return finish(report)
}

func canonicalCase(summary Summary, artifacts map[string][]byte) CaseResult {
	decision, status := DecisionFailClosed, "NOT_SATISFIED"
	if blockingCount(summary) == 0 {
		decision, status = DecisionPass, "SATISFIED"
	}
	return CaseResult{ID: "canonical-surface-closure", Mutation: "NONE",
		Target: "ALL", ExpectedDecision: DecisionPass, ObservedDecision: decision,
		Status: status, EvidenceDigest: digestArtifacts(artifacts)}
}
