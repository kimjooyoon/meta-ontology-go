package rollbackintegrityactivation

func validEligibility(report eligibilityReport) bool {
	if report.Schema != "gooo/rollback-integrity-eligibility/v1" || report.SubjectSHA != PredecessorSHA ||
		report.EvidenceSubjectSHA != EvidenceSubjectSHA || report.Decision != "ELIGIBLE" ||
		report.Resolution != ResolutionExact || report.EnforcementEffect != "NO_EFFECT" ||
		report.Reason != "ROLLBACK_INTEGRITY_PROMOTION_ELIGIBLE" ||
		report.DenominatorID != EligibilityDenominatorID || report.DenominatorDigest != EligibilityDenominatorDigest ||
		report.ReportDigest != EligibilityReportHash || report.RepositoryWrites != 0 || report.PromotionApplied != 0 {
		return false
	}
	t, s := report.Transition, report.Summary
	if t.MetricID != MetricID || t.MetaOperation != MetaOperation || t.FromStatus != "NOT_IMPLEMENTED" ||
		t.FromResolution != "NONE" || t.EligibleStatus != "OPERATING" || t.EligibleResolution != ResolutionExact {
		return false
	}
	if s.DenominatorTotal != 12 || s.BeforeOperating != 9 || s.AfterOperating != 10 ||
		s.BeforeCoverageBPS != 7500 || s.AfterCoverageBPS != 8333 ||
		s.CapsulesTotal != 3 || s.CapsulesExact != 3 || s.CapsuleCoverageBPS != 10000 ||
		s.ShadowCasesTotal != 7 || s.ShadowCasesPassed != 7 ||
		s.ShadowReplaysTotal != 2 || s.ShadowReplaysExact != 2 || s.ShadowReplayCoverageBPS != 10000 ||
		s.MetaOperationsRequired != 1 || s.MetaOperationsObserved != 1 || s.MetaOperationCoverageBPS != 10000 ||
		s.EligiblePaths != 1 || s.UnknownPaths != 0 || s.BlockedPaths != 0 {
		return false
	}
	if len(report.Artifacts) != 3 {
		return false
	}
	for _, artifact := range report.Artifacts {
		if !artifact.Exact {
			return false
		}
	}
	return validEligibilityIndicators(report.Indicators) && validEligibilityMetaOperations(report.MetaOperations)
}
