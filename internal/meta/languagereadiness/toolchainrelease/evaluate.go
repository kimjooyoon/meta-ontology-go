package toolchainrelease

func Evaluate(corpus Corpus, corpusDigest string, evidence []PlatformEvidence,
	expectedHead, conceptDigest string, conceptBound bool) (Report, error) {
	summary := Summary{CasesTotal: len(corpus.Cases), CodeBindings: 6,
		MetricBindings: IndicatorCount, UseCaseBindings: 3}
	if conceptBound {
		summary.ConceptBindings = 1
	} else {
		summary.ConceptDrift = 1
	}
	grouped, unexpected := groupEvidence(evidence, expectedHead, &summary)
	summary.UnexpectedReceipts = unexpected
	observations := map[string]caseObservation{}
	for _, target := range corpus.Targets {
		observeTarget(target, grouped[target.ID], &summary, observations)
	}
	summary.OperatingSystems, summary.Architectures = uniquePlatforms(grouped)
	complete := summary.MissingReceipts == 0 && summary.DuplicateReceipts == 0 &&
		summary.UnexpectedReceipts == 0 && summary.PlatformReceipts == TargetCount
	observations["release-set-completeness"] = caseObservation{
		Observed: "3_OF_3_TARGETS", Digest: corpusDigest, Ready: complete,
	}
	checksums := summary.ChecksumEntries == TargetCount && summary.ChecksumDrift == 0
	observations["release-checksum-manifest"] = caseObservation{
		Observed: "3_SORTED_SHA256_ENTRIES", Digest: receiptSetDigest(evidence), Ready: checksums,
	}
	cases, satisfied := evaluateCases(corpus, observations)
	summary.CasesSatisfied = satisfied
	summary.CaseFailures = summary.CasesTotal - satisfied
	if summary.CasesTotal > 0 {
		summary.ReadinessBPS = satisfied * 10000 / summary.CasesTotal
	}
	proofs := buildProofs(corpusDigest, conceptDigest, evidence)
	if len(proofs) != 3 {
		summary.ProofFailures = 1
	}
	report := assembleReport(expectedHead, corpusDigest, conceptDigest, summary, cases, proofs)
	return finalizeReport(report)
}
