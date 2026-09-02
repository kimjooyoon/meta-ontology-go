package integrationprogress

import "testing"

func TestCompletePortfolioClosesExactDenominator(t *testing.T) {
	report := Evaluate(completeObservation(), true)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != "COMPLETE" || report.Summary.ClosedCells != 150 ||
		report.Summary.MergedPullRequests != 30 || report.Summary.EvidencedMerges != 30 ||
		report.Summary.RunStartDelaySecondsTotal != 1800 || report.Summary.ExecutionSecondsTotal != 3600 ||
		report.Summary.QueuePressureBasisPoints != 7500 {
		t.Fatalf("complete report = %#v", report.Summary)
	}
}

func TestKnownUnmergedPullIsOpenNotUnknown(t *testing.T) {
	input := completeObservation()
	pull := fixturePull(&input, 550)
	pull.State, pull.MergedAt = "open", ""
	pull.RunSelection = "LATEST_OBSERVED"
	report := Evaluate(input, true)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != "PROGRESS_OBSERVED" || report.Summary.OpenCells != 2 ||
		report.Summary.UnknownCells != 0 || report.Summary.ClosedCells != 148 {
		t.Fatalf("open report = %#v", report.Summary)
	}
}

func TestRunObservationFailureLowersResolution(t *testing.T) {
	input := completeObservation()
	pull := fixturePull(&input, 550)
	pull.RunQueryFailure, pull.AuthoritativeRun = "GITHUB_PERMISSION_DENIED", nil
	report := Evaluate(input, true)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != "LOWER_RESOLUTION" || report.Summary.UnknownCells == 0 {
		t.Fatalf("lower-resolution report = %#v", report.Summary)
	}
}
