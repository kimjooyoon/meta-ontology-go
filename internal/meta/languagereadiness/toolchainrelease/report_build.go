package toolchainrelease

func assembleReport(head, corpusDigest, conceptDigest string, summary Summary,
	cases []CaseResult, proofs []Proof) Report {
	decision, reason, resolution := DecisionPass, "TOOLCHAIN_CROSS_PLATFORM_RELEASE_READY", ResolutionExact
	if summary.CaseFailures != 0 || guardrailTotal(summary) != 0 {
		decision, reason, resolution = DecisionFailClosed, "TOOLCHAIN_CROSS_PLATFORM_RELEASE_NOT_READY", ResolutionInvariant
	}
	return Report{
		Schema: ReportSchema, Decision: decision, Reason: reason, Resolution: resolution,
		HeadSHA: head, CorpusDigest: corpusDigest, ConceptDigest: conceptDigest,
		Summary: summary, Cases: cases, Indicators: buildIndicators(summary),
		Proofs: proofs, RepositoryWrites: summary.RepositoryWrites,
	}
}

func finalizeReport(report Report) (Report, error) {
	report.ReportDigest = ""
	digest, err := digestValue(report)
	if err != nil {
		return Report{}, err
	}
	report.ReportDigest = digest
	return report, nil
}
