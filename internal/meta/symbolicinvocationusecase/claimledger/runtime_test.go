package claimledger

import "testing"

func TestBoundRuntimeEvidenceDischargesMemoryClaim(t *testing.T) {
	claim := inScope("memory", "measurement.peak_rss_kib", "POSITIVE_INTEGER", "REGRESSION", "OBSERVE", "capture-peak-rss", nil)
	claim.Evidence.Source = "runtime"
	expected := ExpectedMetrics{
		FixedClaimTotal: 1, InScopeClaimTotal: 1, DischargedTotal: 1, DischargeBasisPoints: 10_000,
		ProofRoutes: ProofRouteCounts{Regression: 1}, ClaimSetDecision: "PASS", Resolution: "EXACT",
	}
	observation := []byte(`{"schema":"observation/v1"}`)
	runtime := runtimeEvidence("abc", digestBytes(observation), 2048)
	report, err := Project(testContract([]ClaimSpec{claim}, expected), observation, runtime, "abc")
	if err != nil {
		t.Fatal(err)
	}
	if report.Claims[0].Status != "DISCHARGED" || report.Inputs[1].Status != "VERIFIED" {
		t.Fatalf("claim=%s runtime=%s", report.Claims[0].Status, report.Inputs[1].Status)
	}
	if report.RuntimeDigest == "" || report.Conformance.Decision != "PASS" {
		t.Fatalf("digest=%s conformance=%s", report.RuntimeDigest, report.Conformance.Decision)
	}
}

func TestMismatchedRuntimeSubjectRemainsUnknown(t *testing.T) {
	claim := inScope("memory", "measurement.peak_rss_kib", "POSITIVE_INTEGER", "REGRESSION", "OBSERVE", "capture-peak-rss", nil)
	claim.Evidence.Source = "runtime"
	expected := ExpectedMetrics{
		FixedClaimTotal: 1, InScopeClaimTotal: 1, UnknownTotal: 1, OpenClaimTotal: 1,
		ProofRoutes: ProofRouteCounts{Regression: 1}, ClaimSetDecision: "FAIL_CLOSED", Resolution: "STAGE_LOCAL",
	}
	observation := []byte(`{}`)
	runtime := runtimeEvidence("wrong", digestBytes(observation), 2048)
	report, err := Project(testContract([]ClaimSpec{claim}, expected), observation, runtime, "abc")
	if err != nil {
		t.Fatal(err)
	}
	if report.Claims[0].Status != "UNKNOWN" || report.Claims[0].Reason != "RUNTIME_EVIDENCE_SUBJECT_MISMATCH" {
		t.Fatalf("status=%s reason=%s", report.Claims[0].Status, report.Claims[0].Reason)
	}
	if report.Evidence[0].Status != "REJECTED" || report.Metrics.FalsePromotionCount != 0 {
		t.Fatalf("evidence=%s false-promotions=%d", report.Evidence[0].Status, report.Metrics.FalsePromotionCount)
	}
}
