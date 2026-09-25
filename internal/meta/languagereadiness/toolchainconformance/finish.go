package toolchainconformance

func finish(report Report) Report {
	report.Indicators = buildIndicators(report.Summary)
	report.Proofs = buildProofs(report.Summary, report.Source)
	return seal(report)
}
