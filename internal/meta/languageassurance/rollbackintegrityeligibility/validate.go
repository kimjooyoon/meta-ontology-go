package rollbackintegrityeligibility

import (
	"fmt"
	"reflect"
)

func bindings(input Input) []ArtifactBinding {
	return []ArtifactBinding{binding(input.Assurance), binding(input.ShadowReportA), binding(input.ShadowReportB)}
}

func binding(capsule Capsule) ArtifactBinding {
	observed := digestBytes(capsule.Payload)
	return ArtifactBinding{Name: capsule.Name, ArtifactID: capsule.ArtifactID,
		ArtifactDigest: capsule.ArtifactDigest, CapsuleDigest: capsule.CapsuleDigest,
		ObservedDigest: observed, Exact: observed == capsule.CapsuleDigest}
}

func exactDigests(input Input) bool {
	return input.Assurance.ArtifactID == AssuranceArtifactID && input.Assurance.ArtifactDigest == AssuranceArtifactDigest &&
		input.Assurance.CapsuleDigest == AssuranceCapsuleDigest && digestBytes(input.Assurance.Payload) == AssuranceCapsuleDigest &&
		input.ShadowReportA.ArtifactID == ShadowArtifactID && input.ShadowReportA.ArtifactDigest == ShadowArtifactDigest &&
		input.ShadowReportA.CapsuleDigest == ShadowReportACapsuleDigest &&
		digestBytes(input.ShadowReportA.Payload) == ShadowReportACapsuleDigest &&
		input.ShadowReportB.ArtifactID == ShadowArtifactID && input.ShadowReportB.ArtifactDigest == ShadowArtifactDigest &&
		input.ShadowReportB.CapsuleDigest == ShadowReportBCapsuleDigest &&
		digestBytes(input.ShadowReportB.Payload) == ShadowReportBCapsuleDigest
}

func validSemantics(assurance assuranceCapsule, shadowA, shadowB shadowReportCapsule) bool {
	return assurance.Schema == "gooo/language-assurance-report/v1" &&
		assurance.SubjectSHA == EvidenceSubjectSHA &&
		assurance.DenominatorDigest == "sha256:e5b266ceeaeb0757a40096fb661982a263370b1e08945dfedbe34f96eb237a02" &&
		assurance.AssuranceDecision == "PARTIAL" && assurance.CandidateDecision == "ALLOW_LIMITED" &&
		assurance.CandidateResolution == ResolutionExact &&
		assurance.Summary.DenominatorTotal == 12 && assurance.Summary.Operating == 9 &&
		assurance.Summary.NotImplemented == 3 && assurance.Summary.ImplementationCoverageBPS == 7500 &&
		assurance.Summary.UnknownTopDecisions == 0 && assurance.Summary.UnresolvedIndicators == 0 &&
		assurance.Summary.ViolatedGuardrails == 0 && assurance.Summary.RepositoryWrites == 0 &&
		assuranceBaseline(assurance) && reflect.DeepEqual(shadowA, shadowB) && validShadowReport(shadowA)
}

func assuranceBaseline(capsule assuranceCapsule) bool {
	matches := 0
	for _, obligation := range capsule.Obligations {
		if obligation.MetricID != MetricID {
			continue
		}
		matches++
		if obligation.Status != "NOT_IMPLEMENTED" || obligation.Resolution != "NONE" ||
			obligation.MetaOperation != nil && *obligation.MetaOperation != "" {
			return false
		}
	}
	return matches == 1
}

func validShadowReport(report shadowReportCapsule) bool {
	summary := report.Summary
	return report.Schema == "gooo/rollback-integrity-shadow/v1" && report.MetricID == MetricID &&
		report.MetaOperation == MetaOperation && report.Decision == "SHADOW_PASS" &&
		report.Resolution == ResolutionExact && report.EnforcementEffect == EffectNone &&
		report.AssuranceSubjectSHA == ShadowAssuranceSubjectSHA && report.EvidenceDigest == ShadowEvidenceDigest &&
		summary.DenominatorTotal == 12 && summary.BeforeOperating == 9 && summary.ProjectedOperating == 10 &&
		summary.BeforeCoverageBPS == 7500 && summary.ProjectedCoverageBPS == 8333 &&
		summary.CasesTotal == 7 && summary.CasesPassed == 7 && summary.CaseCoverageBPS == 10_000 &&
		summary.MetaReportsValid == 7 && summary.CoordinatesTotal == 70 && summary.TerminalCases == 2 &&
		summary.UnknownDecisionCases == 1 && summary.KnownRejectCases == 4 &&
		report.RepositoryWrites == 0 && report.PromotionApplied == 0 &&
		report.ReportDigest == ShadowReportDigest && preservedUnknownDecision(report.Cases)
}

func preservedUnknownDecision(cases []shadowCase) bool {
	if len(cases) != 7 {
		return false
	}
	unknownCases := 0
	for _, item := range cases {
		if !item.Passed {
			return false
		}
		if item.Name != "unknown-guard-decision" {
			continue
		}
		unknownCases++
		if item.ExpectedDecision != DecisionFailClosed || item.ActualDecision != DecisionFailClosed ||
			item.ExpectedResolution != "LOWER_RESOLUTION" || item.ActualResolution != "LOWER_RESOLUTION" ||
			item.ActualMode != "" || item.Unresolved != 3 || item.RepositoryWrites != 0 {
			return false
		}
	}
	return unknownCases == 1
}

func Validate(report Report, input Input) error {
	if report.Schema != ReportSchema || report.DenominatorID != DenominatorID || !reflect.DeepEqual(report, Evaluate(input)) {
		return fmt.Errorf("rollback integrity eligibility report does not replay")
	}
	return nil
}

func ValidateSuite(suite Suite, subjectSHA string) error {
	if suite.Schema != SuiteSchema || !reflect.DeepEqual(suite, RunSuite(subjectSHA)) {
		return fmt.Errorf("rollback integrity eligibility suite does not replay")
	}
	return nil
}
