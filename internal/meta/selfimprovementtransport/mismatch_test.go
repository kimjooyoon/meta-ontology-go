package selfimprovementtransport

import "testing"

func TestKnownTransportMismatchFailsClosed(t *testing.T) {
	_, _, _, metadata, _ := fixture(t)
	report := evaluateFixture(t, metadata, digestBytes([]byte("different archive")))
	if err := ValidateReport(report); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionFailClosed || report.Metrics.FalseTotal != 1 ||
		report.Coordinate.Stage != "TRANSPORT" || report.Coordinate.Step != "verify-archive-digest" {
		t.Fatalf("known mismatch was not fail-closed: %+v", report)
	}
}
