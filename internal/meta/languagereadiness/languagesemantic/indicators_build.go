package languagesemantic

func buildReport(_ Registry, cases []CaseResult, source Source, unregistered, missing, registryDrift int) Report {
	summary := summarize(cases, unregistered, missing, registryDrift)
	resolution := ResolutionExact
	if !source.ObservationKnown || summary.Unresolved > 0 || unregistered > 0 || missing > 0 || registryDrift > 0 {
		resolution = ResolutionLower
	}
	indicators := buildIndicators(summary, resolution)
	decision, reason := DecisionPass, "SEMANTIC_MODEL_EXACTLY_PROVEN"
	for _, indicator := range indicators {
		if !indicator.Satisfied {
			decision = DecisionFailClosed
			reason = "SEMANTIC_MODEL_CONFORMANCE_FAILED"
			if resolution == ResolutionLower {
				reason = "SEMANTIC_MODEL_EVIDENCE_UNKNOWN"
			}
			break
		}
	}
	proofs := buildProofs(summary, source, resolution)
	report := Report{
		Schema:             ReportSchema,
		Decision:           decision,
		Resolution:         resolution,
		ReasonCode:         reason,
		Source:             source,
		Summary:            summary,
		Cases:              cases,
		Indicators:         indicators,
		Proofs:             proofs,
		RepositoryWrites:   0,
		MutationAuthorized: false,
	}
	finalizeReport(&report)
	return report
}
