package claimledger

import "testing"

func TestMismatchedObservationDigestRemainsUnknown(t *testing.T) {
	claim := inScope("memory", "measurement.peak_rss_kib", "POSITIVE_INTEGER", "REGRESSION", "OBSERVE", "capture-peak-rss", nil)
	claim.Evidence.Source = "runtime"
	expected := ExpectedMetrics{
		FixedClaimTotal: 1, InScopeClaimTotal: 1, UnknownTotal: 1, OpenClaimTotal: 1,
		ProofRoutes: ProofRouteCounts{Regression: 1}, ClaimSetDecision: "FAIL_CLOSED", Resolution: "STAGE_LOCAL",
	}
	observation := []byte(`{}`)
	runtime := runtimeEvidence("abc", "sha256:not-the-observation", 2048)
	report, err := Project(testContract([]ClaimSpec{claim}, expected), observation, runtime, "abc")
	if err != nil {
		t.Fatal(err)
	}
	if report.Claims[0].Status != "UNKNOWN" || report.Claims[0].Reason != "RUNTIME_EVIDENCE_OBSERVATION_DIGEST_MISMATCH" {
		t.Fatalf("status=%s reason=%s", report.Claims[0].Status, report.Claims[0].Reason)
	}
	if report.Evidence[0].Status != "REJECTED" || report.Metrics.FalsePromotionCount != 0 {
		t.Fatalf("evidence=%s false-promotions=%d", report.Evidence[0].Status, report.Metrics.FalsePromotionCount)
	}
}
