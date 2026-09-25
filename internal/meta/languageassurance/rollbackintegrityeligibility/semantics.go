package rollbackintegrityeligibility

import "reflect"

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
