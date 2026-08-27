package integrationprogress

import "fmt"

func Validate(report Report) error {
	decision, reason, resolution := decide(report.Summary)
	if report.Schema != ReportSchema || report.Repository != Repository || !validSHA(report.ObserverHeadSHA) ||
		report.CohortID != CohortID || report.Decision != decision || report.Reason != reason ||
		report.Resolution != resolution || report.ObservationDigest == "" ||
		report.MetaProgramDigest != digestBytes(RenderProgram()) || report.RepositoryWrites != 0 ||
		report.PromotionAuthorized || len(report.Cells) != CellDenominator() || len(report.Indicators) != 12 ||
		len(report.Proofs) != 3 || report.Summary.PullRequestsTotal != len(PullNumbers()) ||
		report.Summary.CellsTotal != CellDenominator() {
		return fmt.Errorf("integration progress report contract mismatch")
	}
	if report.Summary.ClosedCells+report.Summary.OpenCells+report.Summary.UnknownCells+
		report.Summary.RefutedCells != report.Summary.CellsTotal {
		return fmt.Errorf("integration progress cell accounting mismatch")
	}
	if err := validateCells(report.Cells); err != nil {
		return err
	}
	digest := report.ReportDigest
	report.ReportDigest = ""
	if digest == "" || digestJSON(report) != digest {
		return fmt.Errorf("integration progress report digest mismatch")
	}
	return nil
}

func validateCells(cells []Cell) error {
	seen := make(map[string]bool, len(cells))
	for _, value := range cells {
		key := itoa(value.PullRequest) + "/" + value.Stage
		if seen[key] || value.Step == "" || value.Reason == "" {
			return fmt.Errorf("integration progress cell identity mismatch")
		}
		seen[key] = true
		if value.State != StateClosed && value.State != StateOpen &&
			value.State != StateUnknown && value.State != StateRefuted {
			return fmt.Errorf("integration progress cell state mismatch")
		}
	}
	return nil
}
