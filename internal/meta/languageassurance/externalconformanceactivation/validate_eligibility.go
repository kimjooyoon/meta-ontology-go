package externalconformanceactivation

func validEligibility(report eligibilityReport) bool {
	if report.Schema != "gooo/external-conformance-eligibility/v1" || report.SubjectSHA != EligibilitySubjectSHA ||
		report.AssuranceSubjectSHA != EligibilityAssuranceSHA || report.Decision != "ELIGIBLE_SHADOW" ||
		report.Resolution != ResolutionExact || report.EnforcementEffect != "NO_EFFECT" ||
		report.Reason != "EXTERNAL_CONFORMANCE_ASSURANCE_ELIGIBLE_SHADOW" ||
		report.ReportDigest != EligibilityReportHash || report.RepositoryWrites != 0 ||
		report.OfficialMutationCount != 0 || report.PromotionApplied != 0 {
		return false
	}
	t := report.Transition
	if t.MetricID != MetricID || t.MetaOperation != EligibilityMetaOperation ||
		t.FromStatus != "NOT_IMPLEMENTED" || t.FromResolution != "NONE" ||
		t.EligibleStatus != "OPERATING" || t.EligibleResolution != ResolutionExact ||
		t.OfficialStatus != "NOT_IMPLEMENTED" || t.OfficialResolution != "NONE" {
		return false
	}
	return validEligibilitySummary(report.Summary) && validEligibilityBindings(report)
}

func validEligibilityBindings(report eligibilityReport) bool {
	if len(report.Artifacts) != 7 || len(report.Indicators) != 18 || len(report.Proofs) != 3 {
		return false
	}
	for _, artifact := range report.Artifacts {
		if !artifact.Exact {
			return false
		}
	}
	classes, proofs := map[string]int{}, map[string]int{}
	for _, item := range report.Indicators {
		if !item.Satisfied || item.Resolution != ResolutionExact || item.Producer != "assuranceeligibility.Evaluate" ||
			item.Consumer != "language-assurance-activation-gate" || item.MetaOperation == "" {
			return false
		}
		classes[item.Class]++
		proofs[item.ProofChoice]++
	}
	return classes["DRIVER"] == 7 && classes["OUTCOME"] == 4 && classes["GUARDRAIL"] == 7 &&
		proofs["FOUNDATION"] == 7 && proofs["COHERENCE"] == 4 && proofs["REGRESSION"] == 7 && validProofs(report.Proofs)
}

func validProofs(proofs []eligibilityProof) bool {
	want := map[string]int{"FOUNDATION": 7, "COHERENCE": 4, "REGRESSION": 7}
	for _, proof := range proofs {
		if proof.Status != "SATISFIED" || proof.Satisfied != want[proof.Choice] || proof.Total != want[proof.Choice] {
			return false
		}
		delete(want, proof.Choice)
	}
	return len(want) == 0
}
