package verticalsliceclosureactivation

func validEligibility(report eligibilityReport) bool {
	if report.Schema != "gooo/vertical-slice-closure-eligibility/v1" || report.SubjectSHA != PredecessorSHA ||
		report.AssuranceSubjectSHA != "145b81c8bb8e4b1eb46cb10af0ea21a6b6be51b5" ||
		report.ShadowEvidenceHead != "5a6f3535ed6984fa8ed4bd9806638f246d9c0263" ||
		report.Decision != "ELIGIBLE" || report.Resolution != ResolutionExact || report.EnforcementEffect != "NO_EFFECT" ||
		report.Reason != "VERTICAL_SLICE_CLOSURE_PROMOTION_ELIGIBLE" ||
		report.AssuranceDenominatorID != "gooo/language-assurance-denominator/v1" ||
		report.AssuranceDenominatorDigest != "sha256:e5b266ceeaeb0757a40096fb661982a263370b1e08945dfedbe34f96eb237a02" ||
		report.ShadowDenominatorDigest != "sha256:6b4b3793133313d430d4c53792baf04cb17fc0fd9ac5592aeb33bc01c0ad6962" ||
		report.ReportDigest != EligibilityReportHash || report.RepositoryWrites != 0 || report.PromotionApplied != 0 { return false }
	t, s := report.Transition, report.Summary
	if t.MetricID != MetricID || t.MetaOperation != EligibilityMetaOperation || t.FromStatus != "NOT_IMPLEMENTED" ||
		t.FromResolution != "NONE" || t.EligibleStatus != "OPERATING" || t.EligibleResolution != ResolutionExact ||
		t.OfficialStatus != "NOT_IMPLEMENTED" || t.OfficialResolution != "NONE" { return false }
	if s.DenominatorTotal != 12 || s.BeforeOperating != 10 || s.EligibleOperating != 11 || s.OfficialOperating != 10 ||
		s.BeforeCoverageBPS != 8333 || s.EligibleCoverageBPS != 9166 || s.OfficialCoverageBPS != 8333 ||
		s.CapsulesTotal != 2 || s.CapsulesExact != 2 || s.CapsuleCoverageBPS != 10000 ||
		s.BoundariesTotal != 6 || s.BoundariesSatisfied != 6 || s.LinksTotal != 12 || s.LinksSatisfied != 12 ||
		s.EligiblePaths != 1 || s.UnknownPaths != 0 || s.BlockedPaths != 0 || s.ObservedRepositoryWrites != 0 { return false }
	if len(report.Artifacts) != 2 { return false }
	for _, artifact := range report.Artifacts { if !artifact.Exact { return false } }
	return validEligibilityIndicators(report.Indicators) && validEligibilityMetaOperations(report.MetaOperations)
}
