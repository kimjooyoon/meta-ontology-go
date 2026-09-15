package toolchainlsp

func summarize(observations map[string]observation, stats runtimeStats, concept ConceptBinding) Summary {
	summary := Summary{
		CasesTotal: len(caseContract), AdvertisedCapabilities: stats.Capabilities,
		ReadFeatures: stats.ReadFeatures, DiagnosticPaths: stats.DiagnosticPaths,
		NavigationPaths: stats.NavigationPaths, SymbolPaths: stats.SymbolPaths,
		SemanticTokenPaths: stats.SemanticTokenPaths, UTF16Replays: stats.UTF16Replays,
		TranscriptReplays: stats.TranscriptReplays, FailClosedPaths: stats.FailClosedPaths,
		ConceptBindings: 1, CodeBindings: len(concept.CodeBindings),
		MetricBindings: len(concept.MetricBindings), UseCaseBindings: concept.UseCaseBindings,
		CapabilityGaps: 8 - stats.Capabilities, UnexpectedProtocolErrors: stats.UnexpectedProtocolErrors,
		DiagnosticGaps: stats.DiagnosticGaps, NonstandardWireFields: stats.NonstandardWireFields,
		StaleNavigationLeaks: stats.StaleLeaks, UnknownNavigationLeaks: stats.UnknownLeaks,
		FailClosedNavigationLeaks: stats.FailClosedLeaks,
	}
	for _, expected := range caseContract {
		observed, ok := observations[expected.ID]
		if !ok {
			summary.MissingCases++
			continue
		}
		if observed.Satisfied {
			summary.CasesSatisfied++
		} else {
			summary.CaseFailures++
		}
		if expected.Group == "PROTOCOL" && observed.Satisfied {
			summary.ProtocolCases++
		}
		if expected.Group == "COUPLING" && observed.Satisfied {
			summary.CouplingCases++
		}
	}
	for id := range observations {
		found := false
		for _, expected := range caseContract {
			if id == expected.ID {
				found = true
				break
			}
		}
		if !found {
			summary.UnexpectedCases++
		}
	}
	summary.ReadinessBPS = summary.CasesSatisfied * 10000 / summary.CasesTotal
	return summary
}
