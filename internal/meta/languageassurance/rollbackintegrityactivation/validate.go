package rollbackintegrityactivation

import "encoding/json"

func validateEvidence(input Input) (eligibilityReport, int, int, int, string, string) {
	bindings := artifactBindings(input)
	rawExact := 0
	for _, binding := range bindings {
		if binding.Exact {
			rawExact++
		}
	}
	if !validSHA(input.SubjectSHA) || len(input.Assurance) == 0 || len(input.Eligibility) == 0 {
		return eligibilityReport{}, rawExact, 0, 0, ReasonUnavailable, ResolutionUnknown
	}
	var assurance assuranceReport
	var report eligibilityReport
	if json.Unmarshal(input.Assurance, &assurance) != nil || json.Unmarshal(input.Eligibility, &report) != nil {
		return eligibilityReport{}, rawExact, 0, 0, ReasonUnavailable, ResolutionUnknown
	}
	if report.Decision != "ELIGIBLE" {
		return report, rawExact, 0, 0, ReasonEligibilityUnknown, ResolutionUnknown
	}
	if rawExact != len(bindings) {
		return report, rawExact, 0, 0, ReasonDigestMismatch, ResolutionInvariant
	}
	if !validAssurance(assurance) {
		return report, rawExact, 0, 0, ReasonAssuranceMismatch, ResolutionInvariant
	}
	if !validEligibility(report) {
		return report, rawExact, 1, 0, ReasonEligibilityNotExact, ResolutionInvariant
	}
	return report, rawExact, 1, 1, "", ResolutionExact
}

func validAssurance(report assuranceReport) bool {
	if report.Schema != "gooo/language-assurance-report/v1" || report.SubjectSHA != PredecessorSHA ||
		report.DenominatorID != "gooo/language-assurance-denominator/v1" || report.AssuranceDecision != "PARTIAL" ||
		report.CandidateDecision != "ALLOW_LIMITED" || report.CandidateResolution != ResolutionExact ||
		report.ReportDigest != "sha256:0644a5a29d16705d88b247141f3875c5ef3e7e6cc16c74ca94433e1958cc52dc" {
		return false
	}
	s := report.Summary
	if s.DenominatorTotal != 12 || s.Operating != 9 || s.NotImplemented != 3 || s.ImplementationCoverageBPS != 7500 ||
		s.UnresolvedIndicators != 0 || s.ViolatedGuardrails != 0 || s.RepositoryWrites != 0 || len(report.Obligations) != 12 {
		return false
	}
	matches := 0
	for _, item := range report.Obligations {
		if item.MetricID == MetricID && item.Status == "NOT_IMPLEMENTED" && item.Resolution == "NONE" && item.MetaOperation == "" {
			matches++
		}
	}
	return matches == 1
}

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
