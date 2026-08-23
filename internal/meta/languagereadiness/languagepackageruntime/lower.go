package languagepackageruntime

func lowerResolution(source Source, reason string) Report {
	summary := Summary{Total: FixedTotal, Unresolved: FixedTotal, MetricBindings: FixedIndicators}
	report := Report{Schema: ReportSchema, Decision: DecisionClosed, Resolution: ResolutionLower,
		ReasonCode: reason, Source: source, Summary: summary}
	report.Indicators = indicators(summary, report.Resolution)
	report.Proofs, report.Stages = proofs(report), stages(source, summary)
	report.ReportDigest = reportDigest(report)
	return report
}

func allIndicators(indicators []Indicator) bool {
	if len(indicators) != FixedIndicators {
		return false
	}
	for _, indicator := range indicators {
		if !indicator.Satisfied {
			return false
		}
	}
	return true
}

func allProofs(proofs []Proof) bool {
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
