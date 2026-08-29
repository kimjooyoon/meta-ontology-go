package opentofuobservation

import "fmt"

func ValidateReport(report Report, expectedSubject, expectedContract string) error {
	if report.Schema != ReportSchema || report.ContractID != expectedContract || report.SubjectSHA != expectedSubject || report.MetaOperation != MetaOperation {
		return fmt.Errorf("report context is invalid")
	}
	if len(report.UserPaths) != len(fixedPaths) {
		return fmt.Errorf("report user-path denominator is invalid")
	}
	for index, path := range fixedPaths {
		if report.UserPaths[index] != path {
			return fmt.Errorf("report user-path identity is invalid")
		}
	}
	if len(report.Cells) != len(fixedCells) || report.Summary.CellsTotal != len(fixedCells) {
		return fmt.Errorf("report cell denominator is invalid")
	}
	seen := map[string]bool{}
	for index, cell := range report.Cells {
		spec := fixedCells[index]
		if cell.ID != spec.ID || cell.MetaOperation != spec.MetaOperation || cell.ProofChoice != spec.ProofChoice || cell.Indicator != spec.Indicator || seen[cell.ID] {
			return fmt.Errorf("report cell identity is invalid")
		}
		seen[cell.ID] = true
		if cell.Decision == DecisionPass && cell.State != "CLOSED" {
			return fmt.Errorf("closed cell state is invalid")
		}
	}
	if report.Summary.FoundationClosed != 4 || report.Summary.CoherenceClosed != 4 || report.Summary.RegressionClosed != 4 {
		return fmt.Errorf("proof denominator is invalid")
	}
	if report.Summary.ThreePaths != 3 || report.RepositoryWrites != 0 || report.LocalTestExecutions != 0 || report.PromotionAuthorized {
		return fmt.Errorf("boundary fields are invalid")
	}
	expected, err := sealedReportDigest(report)
	if err != nil || expected != report.ReportDigest || !validDigest(report.ReportDigest) {
		return fmt.Errorf("report digest is not sealed")
	}
	return validateDecision(report)
}

func validateDecision(report Report) error {
	refuted, unknown := 0, 0
	for _, cell := range report.Cells {
		if cell.Decision == DecisionRefuted { refuted++ }
		if cell.Decision == DecisionUnknown { unknown++ }
	}
	if refuted > 0 && report.Decision != DecisionRefuted { return fmt.Errorf("REFUTED precedence is missing") }
	if refuted == 0 && unknown > 0 && report.Decision != DecisionUnknown { return fmt.Errorf("UNKNOWN decision is missing") }
	if refuted == 0 && unknown == 0 && report.Decision != DecisionPass { return fmt.Errorf("PASS decision is missing") }
	return nil
}
