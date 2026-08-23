package languagepackageruntime

import "fmt"

func Validate(report Report, expectedHead string) error {
	if report.Schema != ReportSchema || report.Source.ExpectedHeadSHA != expectedHead ||
		report.ReportDigest == "" || report.ReportDigest != reportDigest(report) {
		return fmt.Errorf("package runtime report binding invalid")
	}
	if report.Decision != DecisionPass && report.Decision != DecisionClosed {
		return fmt.Errorf("package runtime decision unknown")
	}
	if report.Resolution != ResolutionExact && report.Resolution != ResolutionLower {
		return fmt.Errorf("package runtime resolution unknown")
	}
	if report.RepositoryWrites != 0 || report.MutationAuthorized {
		return fmt.Errorf("package runtime report has effects")
	}
	if report.Decision == DecisionPass {
		if report.Resolution != ResolutionExact || report.Summary.Total != FixedTotal ||
			report.Summary.Satisfied != FixedTotal || len(report.Cases) != FixedTotal ||
			!allIndicators(report.Indicators) || !allProofs(report.Proofs) || len(report.Stages) != 5 {
			return fmt.Errorf("package runtime exact evidence incomplete")
		}
	}
	return nil
}
