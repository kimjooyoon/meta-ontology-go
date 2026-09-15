package languagedelivery

func Evaluate(contract Contract, manifest SourceManifest, evidence EvidenceSet, head string) (Report, error) {
	if err := ValidateContract(contract); err != nil {
		return Report{}, err
	}
	sources, decoded := inspectEvidence(manifest, evidence, head)
	results := make([]ObligationResult, 0, len(contract.Obligations))
	for _, item := range contract.Obligations {
		results = append(results, evaluateObligation(item, sources, decoded))
	}
	summary := summarize(results, sources, decoded)
	decision, resolution, reason := reportDecision(summary.Coordinates, summary.Effects)
	report := Report{
		Schema: ReportSchema, Decision: decision, Resolution: resolution, Reason: reason,
		SubjectSHA: head, ContractID: contract.ContractID,
		ContractDigest: digestValue(contract), ManifestDigest: digestValue(manifest),
		Summary: summary, Sources: sources, Obligations: results,
		NotClaimed: append([]string(nil), contract.NotClaimed...),
	}
	report.FactsDigest = factsDigest(report)
	report.Views = buildViews(results, decision, resolution, report.FactsDigest)
	report.Indicators = buildIndicators(summary)
	report.Proofs = buildProofs(summary, report.FactsDigest)
	report.Digest = digestValue(report)
	return report, nil
}

func reportDecision(coordinates Coordinates, effects EffectSummary) (string, string, string) {
	if coordinates.Unknown != 0 {
		return "FAIL_CLOSED", "LOWER_RESOLUTION", "LANGUAGE_DELIVERY_EVIDENCE_UNKNOWN"
	}
	if coordinates.NotSatisfied != 0 || effects.RepositoryWrites != 0 || effects.MutationAuthority {
		return "FAIL_CLOSED", "INVARIANT_ONLY", "LANGUAGE_DELIVERY_CONTRACT_VIOLATED"
	}
	if coordinates.NotImplemented != 0 {
		return "INCOMPLETE", "EXACT", "LANGUAGE_DELIVERY_KNOWN_GAPS"
	}
	return "PASS", "EXACT", "LANGUAGE_DELIVERY_CONTRACT_SATISFIED"
}
