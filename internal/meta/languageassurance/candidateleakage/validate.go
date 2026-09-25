package candidateleakage

import (
	"fmt"
	"reflect"
)

func validateBoundary(input Input) string {
	if input.Schema != InputSchema || !validSHA(input.SubjectSHA) {
		return ReasonSubjectBindingMismatch
	}
	if input.Candidate.SubjectSHA != input.SubjectSHA ||
		input.Promotion.SubjectSHA != input.SubjectSHA || input.Official.SubjectSHA != input.SubjectSHA {
		return ReasonSubjectBindingMismatch
	}
	if !validDigest(input.Candidate.Digest) || !validDigest(input.Promotion.CandidateDigest) ||
		!validDigest(input.Promotion.EvidenceDigest) {
		return ReasonDigestBindingMismatch
	}
	if !knownCandidateDecision(input.Candidate.Decision) ||
		!knownPromotionDecision(input.Promotion.Decision) || !knownOfficialDecision(input.Official.Decision) {
		return ReasonDecisionUnknown
	}
	if input.Candidate.Resolution != ResolutionExact || input.Promotion.Resolution != ResolutionExact ||
		!knownOfficialResolution(input.Official.Status, input.Official.Resolution) {
		return ReasonResolutionUnknown
	}
	return ""
}

func knownCandidateDecision(value string) bool {
	return value == CandidateAllowLimited || value == CandidateBlock
}

func knownPromotionDecision(value string) bool {
	return value == PromotionAuthorized || value == PromotionDenied || value == PromotionFailClosed
}

func knownOfficialDecision(value string) bool {
	switch value {
	case OfficialAllow, OfficialPass, OfficialFixedPoint, OfficialAuthorized, OfficialBlock, OfficialFailClosed:
		return true
	default:
		return false
	}
}

func knownOfficialResolution(status, resolution string) bool {
	return (status == OfficialOperating && resolution == ResolutionExact) ||
		(status == OfficialNotImplemented && resolution == OfficialResolutionNone)
}

func Validate(report Report, input Input) error {
	if report.Schema != ReportSchema || report.DenominatorID != DenominatorID ||
		!validDigest(report.ReportDigest) {
		return fmt.Errorf("candidate leakage report contract mismatch")
	}
	if !reflect.DeepEqual(report, Evaluate(input)) {
		return fmt.Errorf("candidate leakage report does not replay")
	}
	return nil
}

func ValidateSuite(suite Suite, subjectSHA string) error {
	if suite.Schema != SuiteSchema || suite.DenominatorID != DenominatorID ||
		!validDigest(suite.SuiteDigest) {
		return fmt.Errorf("candidate leakage suite contract mismatch")
	}
	if !reflect.DeepEqual(suite, RunSuite(subjectSHA)) {
		return fmt.Errorf("candidate leakage suite does not replay")
	}
	return nil
}
