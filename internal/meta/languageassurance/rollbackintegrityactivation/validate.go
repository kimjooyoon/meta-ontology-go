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
