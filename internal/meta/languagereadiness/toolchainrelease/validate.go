package toolchainrelease

import "fmt"

func Validate(report Report, expectedHead string) error {
	if report.Decision == DecisionFailClosed {
		return fmt.Errorf("FAIL_CLOSED / TOOLCHAIN_RELEASE_REPORT_FAILED")
	}
	if report.Decision != DecisionPass {
		return fmt.Errorf("FAIL_CLOSED / TOOLCHAIN_RELEASE_DECISION_UNKNOWN")
	}
	if report.Resolution != ResolutionExact {
		return fmt.Errorf("FAIL_CLOSED / TOOLCHAIN_RELEASE_RESOLUTION_LOWERED")
	}
	if report.Schema != ReportSchema || report.HeadSHA != expectedHead {
		return fmt.Errorf("TOOLCHAIN_RELEASE_REPORT_IDENTITY_MISMATCH")
	}
	if report.Summary.CasesSatisfied != CaseCount ||
		report.Summary.CasesTotal != CaseCount ||
		report.Summary.ReadinessBPS != 10000 ||
		len(report.Cases) != CaseCount ||
		len(report.Indicators) != IndicatorCount ||
		len(report.Proofs) != 3 {
		return fmt.Errorf("TOOLCHAIN_RELEASE_REPORT_COUNT_MISMATCH")
	}
	for _, item := range report.Cases {
		if item.Status != CaseSatisfied {
			return fmt.Errorf("TOOLCHAIN_RELEASE_CASE_NOT_SATISFIED")
		}
	}
	for _, item := range report.Indicators {
		if !item.Satisfied {
			return fmt.Errorf("TOOLCHAIN_RELEASE_INDICATOR_NOT_SATISFIED")
		}
	}
	if guardrailTotal(report.Summary) != 0 || report.RepositoryWrites != 0 {
		return fmt.Errorf("TOOLCHAIN_RELEASE_GUARDRAIL_NONZERO")
	}
	expected, err := finalizeReport(report)
	if err != nil || expected.ReportDigest != report.ReportDigest {
		return fmt.Errorf("TOOLCHAIN_RELEASE_REPORT_DIGEST_MISMATCH")
	}
	return validateProofChoices(report.Proofs)
}
