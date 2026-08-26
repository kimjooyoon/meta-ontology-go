package verticalsliceclosureshadow

import "fmt"

func Validate(report Report) error {
	if report.Schema != Schema || report.MetricID != MetricID ||
		report.MetaOperation != MetaOperation || report.EnforcementEffect != EnforcementNoEffect ||
		!validSHA(report.HeadSHA) || !validDigest(report.AssuranceDigest) ||
		!validDigest(report.DenominatorDigest) || !validDigest(report.ReportDigest) ||
		report.RepositoryWrites != 0 || report.PromotionApplied != 0 ||
		report.Summary.DenominatorTotal != officialTotal ||
		report.Summary.BoundariesTotal != boundaryTotal ||
		report.Summary.LinksTotal != linkTotal || len(report.Indicators) != 6 ||
		!validSeal(report) || !validIndicatorShape(report.Indicators) {
		return fmt.Errorf("vertical slice shadow report is incomplete")
	}
	if len(report.Boundaries) != 0 && len(report.Boundaries) != boundaryTotal {
		return fmt.Errorf("vertical slice boundary receipt count mismatch")
	}
	if report.Decision == DecisionShadowPass {
		return validateExact(report)
	}
	if report.Decision != DecisionFailClosed ||
		(report.Resolution != ResolutionLower && report.Resolution != ResolutionInvariant) ||
		report.Summary.ProjectedOperating != beforeOperating ||
		report.Summary.ProjectedCoverageBPS != beforeCoverageBPS {
		return fmt.Errorf("vertical slice failure did not preserve the baseline")
	}
	return nil
}

func validateExact(report Report) error {
	summary := report.Summary
	if report.Reason != ReasonShadowPass || report.Resolution != ResolutionExact ||
		report.AssuranceSubjectSHA != PredecessorSHA ||
		report.AssuranceDigest != AssuranceDigest ||
		report.DenominatorDigest != DenominatorDigest ||
		summary.BeforeOperating != beforeOperating ||
		summary.ProjectedOperating != projectedOperating ||
		summary.BeforeCoverageBPS != beforeCoverageBPS ||
		summary.ProjectedCoverageBPS != projectedCoverageBPS ||
		summary.BoundariesSatisfied != boundaryTotal || summary.UnknownBoundaries != 0 ||
		summary.BlockedBoundaries != 0 || summary.BoundaryCoverageBPS != 10000 ||
		summary.LinksSatisfied != linkTotal || summary.LinkCoverageBPS != 10000 ||
		summary.EvidenceAvailable != boundaryTotal ||
		summary.UnknownTopDecisions != 0 || summary.KnownFailures != 0 ||
		summary.ObservedRepositoryWrites != 0 || len(report.Boundaries) != boundaryTotal ||
		!allBoundariesExact(report.Boundaries) || !allIndicatorsSatisfied(report.Indicators) {
		return fmt.Errorf("vertical slice exact shadow contract failed")
	}
	return nil
}

func validSeal(report Report) bool {
	digest := report.ReportDigest
	report.ReportDigest = ""
	return digestJSON(report) == digest
}
