package rollbackintegrityshadow

import "fmt"

func Validate(report Report) error {
	if report.Schema != Schema || report.MetricID != MetricID ||
		report.MetaOperation != MetaOperation || report.EnforcementEffect != EnforcementNoEffect ||
		len(report.Indicators) != 6 || !validDigest(report.EvidenceDigest) ||
		!validDigest(report.ReportDigest) || report.RepositoryWrites != 0 ||
		report.PromotionApplied != 0 || !validSeal(report) {
		return fmt.Errorf("rollback integrity shadow report is incomplete")
	}
	if report.Decision == DecisionShadowPass {
		return validateExact(report)
	}
	if report.Decision != DecisionFailClosed ||
		(report.Resolution != ResolutionLower && report.Resolution != ResolutionInvariant) ||
		report.Summary.ProjectedOperating != report.Summary.BeforeOperating ||
		report.Summary.ProjectedCoverageBPS != report.Summary.BeforeCoverageBPS {
		return fmt.Errorf("rollback integrity failure did not preserve the baseline")
	}
	return nil
}

func validateExact(report Report) error {
	summary := report.Summary
	if report.Resolution != ResolutionExact || report.Reason != ReasonShadowPass ||
		report.AssuranceSubjectSHA != PredecessorSHA || report.EvidenceDigest != AssuranceDigest ||
		summary.DenominatorTotal != 12 || summary.BeforeOperating != 9 ||
		summary.ProjectedOperating != 10 || summary.BeforeCoverageBPS != 7500 ||
		summary.ProjectedCoverageBPS != 8333 || summary.CasesTotal != caseTotal ||
		summary.CasesPassed != caseTotal || summary.CaseCoverageBPS != 10000 ||
		summary.MetaReportsValid != caseTotal || summary.CoordinatesTotal != caseTotal*10 ||
		summary.TerminalCases != 2 || summary.UnknownDecisionCases != 1 ||
		summary.KnownRejectCases != 4 || len(report.Cases) != caseTotal ||
		!allCasesPass(report.Cases) || !allIndicatorsPass(report.Indicators) {
		return fmt.Errorf("rollback integrity exact shadow contract failed")
	}
	return nil
}

func validSeal(report Report) bool {
	digest := report.ReportDigest
	report.ReportDigest = ""
	return digestJSON(report) == digest
}

func allCasesPass(cases []CaseResult) bool {
	for _, result := range cases {
		if !result.Passed {
			return false
		}
	}
	return true
}

func allIndicatorsPass(indicators []Indicator) bool {
	for _, indicator := range indicators {
		if !indicator.Satisfied || indicator.MetaOperation != MetaOperation {
			return false
		}
	}
	return true
}
