package candidateleakageactivation

import (
	"encoding/json"

	eligibility "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageassurance/candidateleakageeligibility"
)

func validateEvidence(input Input) (eligibility.Report, int, int, int, string, string) {
	bindings := artifactBindings(input)
	rawExact := 0
	for _, binding := range bindings {
		if binding.Exact {
			rawExact++
		}
	}
	if !validSHA(input.SubjectSHA) || len(input.Assurance) == 0 || len(input.Eligibility) == 0 {
		return eligibility.Report{}, rawExact, 0, 0, ReasonUnavailable, ResolutionUnknown
	}
	if rawExact != len(bindings) {
		return eligibility.Report{}, rawExact, 0, 0, ReasonDigestMismatch, ResolutionInvariant
	}
	var assurance assuranceReport
	var report eligibility.Report
	if json.Unmarshal(input.Assurance, &assurance) != nil || json.Unmarshal(input.Eligibility, &report) != nil {
		return eligibility.Report{}, rawExact, 0, 0, ReasonUnavailable, ResolutionUnknown
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
	if report.Schema != "gooo/language-assurance-report/v1" || report.SubjectSHA != PredecessorSHA {
		return false
	}
	s := report.Summary
	if s.DenominatorTotal != 12 || s.Operating != 7 || s.NotImplemented != 5 || s.ImplementationCoverageBPS != 5833 || s.UnresolvedIndicators != 0 || s.ViolatedGuardrails != 0 || s.RepositoryWrites != 0 {
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

func validEligibility(report eligibility.Report) bool {
	copy := report
	copy.ReportDigest = ""
	if report.Schema != eligibility.ReportSchema || report.SubjectSHA != PredecessorSHA || report.EvidenceSubjectSHA != EvidenceSubjectSHA || report.Decision != eligibility.DecisionEligible || report.Resolution != eligibility.ResolutionExact || report.EnforcementEffect != "NO_EFFECT" || report.Reason != eligibility.ReasonEligible {
		return false
	}
	if report.DenominatorID != EligibilityDenominatorID || report.DenominatorDigest == "" || report.ReportDigest != EligibilityReportHash || digestValue(copy) != EligibilityReportHash || report.RepositoryWrites != 0 || report.PromotionApplied != 0 {
		return false
	}
	t, s := report.Transition, report.Summary
	if t.MetricID != MetricID || t.MetaOperation != MetaOperation || t.FromStatus != "NOT_IMPLEMENTED" || t.FromResolution != "NONE" || t.EligibleStatus != "OPERATING" || t.EligibleResolution != ResolutionExact {
		return false
	}
	return s.DenominatorTotal == 12 && s.BeforeOperating == 7 && s.AfterOperating == 8 && s.BeforeCoverageBPS == 5833 && s.AfterCoverageBPS == 6666 && s.CapsulesTotal == 2 && s.CapsulesExact == 2 && s.CapsuleCoverageBPS == 10000 && s.ShadowCasesTotal == 6 && s.ShadowCasesPassed == 6 && s.EligiblePaths == 1 && s.UnknownPaths == 0 && s.BlockedPaths == 0 && len(report.MetaOperations) == 5 && validEligibilityIndicators(report.Indicators)
}
