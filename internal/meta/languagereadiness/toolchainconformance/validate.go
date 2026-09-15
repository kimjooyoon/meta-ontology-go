package toolchainconformance

import "fmt"

func Validate(report Report, expectedHead string) error {
	if report.Schema != Schema {
		return fmt.Errorf("FAIL_CLOSED: unknown conformance report schema %q", report.Schema)
	}
	if report.Decision != DecisionPass || report.Resolution != ResolutionExact {
		return fmt.Errorf("FAIL_CLOSED: conformance decision %q/%q",
			report.Decision, report.Resolution)
	}
	if !validHead(expectedHead) || report.Source.ExpectedHeadSHA != expectedHead ||
		!report.Source.ObservationKnown {
		return fmt.Errorf("FAIL_CLOSED: conformance head observation is unknown")
	}
	summary := report.Summary
	if blockingCount(summary) != 0 ||
		summary.SurfacesSatisfied != ExpectedSurfaceCount ||
		summary.SurfacesTotal != ExpectedSurfaceCount ||
		summary.CasesSatisfied != ExpectedCaseCount ||
		summary.CasesTotal != ExpectedCaseCount ||
		summary.ExecutedCases != ExpectedCaseCount ||
		summary.IndicatorsSatisfied != ExpectedIndicatorCount ||
		summary.IndicatorsTotal != ExpectedIndicatorCount ||
		summary.ProofsPassed != ExpectedProofCount ||
		summary.ProofsTotal != ExpectedProofCount ||
		summary.HeadBindings != ExpectedSurfaceCount ||
		summary.TamperRejections != ExpectedTamperCount ||
		summary.TamperTotal != ExpectedTamperCount {
		return fmt.Errorf("FAIL_CLOSED: conformance summary is incomplete")
	}
	if len(report.Surfaces) != ExpectedSurfaceCount ||
		len(report.Cases) != ExpectedTamperCount+1 ||
		len(report.Indicators) != ExpectedMetricCount || len(report.Proofs) != 3 {
		return fmt.Errorf("FAIL_CLOSED: conformance evidence cardinality drift")
	}
	for _, indicator := range report.Indicators {
		if !indicator.Satisfied || indicator.Resolution != ResolutionExact {
			return fmt.Errorf("FAIL_CLOSED: conformance indicator %q failed", indicator.MetricID)
		}
	}
	for _, proof := range report.Proofs {
		if !proof.Passed {
			return fmt.Errorf("FAIL_CLOSED: conformance proof %q failed", proof.Choice)
		}
	}
	if report.RepositoryWrites != 0 || report.MutationAuthorized {
		return fmt.Errorf("FAIL_CLOSED: conformance evaluator claimed mutation")
	}
	digest := report.ReportDigest
	if !validDigest(digest) || seal(report).ReportDigest != digest {
		return fmt.Errorf("FAIL_CLOSED: conformance report digest mismatch")
	}
	return nil
}
