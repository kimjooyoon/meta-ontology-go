package languageutility

import "testing"

func TestUnknownEvidenceLowersResolutionAndFailsClosed(t *testing.T) {
	contract := fixtureContract()
	observation := fixtureObservation(contract)
	observation.Cells[0].State = StateUnknown
	observation.Cells[0].Step = "READ_CI_PLAN_REPORT"
	observation.Cells[0].Reason = "REPORT_UNREADABLE"
	report, err := Evaluate(contract, observation)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "LOWER_RESOLUTION" ||
		report.Reason != "UTILITY_EVIDENCE_UNKNOWN" || report.Summary.UnknownCells != 1 {
		t.Fatalf("unknown report = %#v", report)
	}
	cell := report.Cells[0]
	if cell.StageID != "SOURCE_PRESENT" || cell.Step != "READ_CI_PLAN_REPORT" ||
		cell.Reason != "REPORT_UNREADABLE" || cell.ClaimStatus != "OPEN" {
		t.Fatalf("unknown coordinate = %#v", cell)
	}
}

func TestRefutationDoesNotDischargeClaim(t *testing.T) {
	contract := fixtureContract()
	observation := fixtureObservation(contract)
	observation.Cells[0].State = StateRefuted
	observation.Cells[0].Reason = "SOURCE_DIGEST_MISMATCH"
	report, err := Evaluate(contract, observation)
	if err != nil {
		t.Fatal(err)
	}
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "EXACT" ||
		report.Summary.RefutedCells != 1 || report.Cells[0].ClaimStatus != "REFUTED" {
		t.Fatalf("refuted report = %#v", report)
	}
}
