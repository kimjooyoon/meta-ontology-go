package verticalsliceclosureactivation

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
	var eligibility eligibilityReport
	if json.Unmarshal(input.Assurance, &assurance) != nil || json.Unmarshal(input.Eligibility, &eligibility) != nil {
		return eligibilityReport{}, rawExact, 0, 0, ReasonUnavailable, ResolutionUnknown
	}
	if eligibility.Decision != "ELIGIBLE" {
		return eligibility, rawExact, 0, 0, ReasonEligibilityUnknown, ResolutionUnknown
	}
	if rawExact != len(bindings) {
		return eligibility, rawExact, 0, 0, ReasonDigestMismatch, ResolutionInvariant
	}
	if !validAssurance(assurance) {
		return eligibility, rawExact, 0, 0, ReasonAssuranceMismatch, ResolutionInvariant
	}
	if !validEligibility(eligibility) {
		return eligibility, rawExact, 1, 0, ReasonEligibilityInvalid, ResolutionInvariant
	}
	return eligibility, rawExact, 1, 1, "", ResolutionExact
}

func validAssurance(report assuranceReport) bool {
	if report.Schema != "gooo/language-assurance-report/v1" || report.SubjectSHA != PredecessorSHA ||
		report.DenominatorID != "gooo/language-assurance-denominator/v1" ||
		report.DenominatorDigest != "sha256:e5b266ceeaeb0757a40096fb661982a263370b1e08945dfedbe34f96eb237a02" ||
		report.AssuranceDecision != "PARTIAL" || report.CandidateDecision != "ALLOW_LIMITED" ||
		report.CandidateResolution != ResolutionExact || report.ReportDigest != AssuranceReportHash {
		return false
	}
	s := report.Summary
	if s.DenominatorTotal != 12 || s.Operating != 10 || s.NotImplemented != 2 || s.ImplementationCoverageBPS != 8333 ||
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
