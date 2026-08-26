package claimledger

import "testing"

func TestUnknownIsLocalizedAndCannotPass(t *testing.T) {
	claims := []ClaimSpec{
		inScope("schema", "schema", "EQUALS", "FOUNDATION", "OBSERVE", "decode-schema", rawString("observed/v1")),
		inScope("memory", "performance.peak_rss_kib", "POSITIVE_INTEGER", "COHERENCE", "OBSERVE", "capture-peak-rss", nil),
		excluded("human", "REGRESSION"),
	}
	expected := ExpectedMetrics{
		FixedClaimTotal: 3, InScopeClaimTotal: 2, DischargedTotal: 1, UnknownTotal: 1,
		ExcludedTotal: 1, OpenClaimTotal: 1, DischargeBasisPoints: 5000,
		ProofRoutes:      ProofRouteCounts{Foundation: 1, Coherence: 1, Regression: 1},
		ClaimSetDecision: "FAIL_CLOSED", Resolution: "STAGE_LOCAL",
	}
	report, err := Project(testContract(claims, expected), []byte(`{"schema":"observed/v1"}`), "abc")
	if err != nil {
		t.Fatal(err)
	}
	if report.Conformance.Decision != "PASS" || report.ClaimSet.Decision != "FAIL_CLOSED" {
		t.Fatalf("conformance=%s claim-set=%s", report.Conformance.Decision, report.ClaimSet.Decision)
	}
	if report.ClaimSet.Resolution != "STAGE_LOCAL" || report.Metrics.FalsePromotionCount != 0 {
		t.Fatalf("resolution=%s false-promotions=%d", report.ClaimSet.Resolution, report.Metrics.FalsePromotionCount)
	}
	if len(report.OpenClaimIDs) != 1 || report.OpenClaimIDs[0] != "memory" {
		t.Fatalf("open claims=%v", report.OpenClaimIDs)
	}
	if report.Claims[0].Status != "DISCHARGED" || report.Claims[1].Status != "UNKNOWN" {
		t.Fatalf("statuses=%s,%s", report.Claims[0].Status, report.Claims[1].Status)
	}
	if report.Claims[1].Coordinate.Step != "capture-peak-rss" || report.Claims[1].Reason != "memory_MISSING" {
		t.Fatalf("unknown coordinate=%+v reason=%s", report.Claims[1].Coordinate, report.Claims[1].Reason)
	}
}
