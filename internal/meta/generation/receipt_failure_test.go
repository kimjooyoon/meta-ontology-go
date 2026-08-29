package generation

import (
	"strings"
	"testing"
)

func TestVerifyReceiptsPreservesRefutedPrecedenceAndUnknownFields(t *testing.T) {
	plan := actionableReceiptPlan()
	process := ProcessObservation{
		Command: []string{"<not-executed>", "extract-function"}, ExitCode: -1,
		RawStdoutDigest: "sha256:" + strings.Repeat("0", 64),
		StdoutDigest: "sha256:" + strings.Repeat("0", 64),
		RawStderrDigest: "sha256:" + strings.Repeat("0", 64),
		StderrDigest: "sha256:" + strings.Repeat("0", 64),
	}
	failures := []ObservationFailure{
		{
			ActionIndicatorID: plan.Selected[0].IndicatorID, Decision: "UNKNOWN",
			Stage: "execute-operation", Step: "run-selected-operation",
			Reason:        "INSTANCE_EVIDENCE_UNAVAILABLE",
			UnknownClass:  ReceiptUnknownClassDirectMissing,
			NextOperation: "restore-operation-evidence", BlockedBy: []string{}, Executor: process,
		},
		{
			ActionIndicatorID: plan.Selected[1].IndicatorID, Decision: "REFUTED",
			Stage: "evaluate-operation", Step: "validate-selected-operation",
			Reason: "KNOWN_CONTRADICTION", NextOperation: "report-counterexample",
			BlockedBy: []string{}, Executor: process,
		},
	}
	report := VerifyReceiptsWithFailures(plan, nil, failures)
	if report.Decision != ReceiptDecisionRefuted || report.Reason != ReceiptReasonRefutedOperation {
		t.Fatalf("refuted failure was not dominant: %+v", report)
	}
	var unknown ObservationFailure
	foundUnknown := false
	for _, failure := range report.Failures {
		if failure.Decision == "UNKNOWN" {
			unknown, foundUnknown = failure, true
			break
		}
	}
	if len(report.Failures) != 2 || !foundUnknown || unknown.Stage == "" || unknown.Step == "" ||
		unknown.Reason == "" || unknown.UnknownClass == "" || unknown.NextOperation == "" || unknown.BlockedBy == nil {
		t.Fatalf("unknown failure lost its six-field evidence: %+v", report.Failures)
	}
}

func TestVerifyReceiptsRejectsUnknownFailureWithoutSixFields(t *testing.T) {
	plan := sourcepolicyPlanForFailureTest()
	process := ProcessObservation{Command: []string{"<not-executed>"}, ExitCode: -1,
		RawStdoutDigest: "sha256:" + strings.Repeat("0", 64), StdoutDigest: "sha256:" + strings.Repeat("0", 64),
		RawStderrDigest: "sha256:" + strings.Repeat("0", 64), StderrDigest: "sha256:" + strings.Repeat("0", 64)}
	failure := ObservationFailure{ActionIndicatorID: plan.Selected[0].IndicatorID, Decision: "UNKNOWN",
		UnknownClass: ReceiptUnknownClassDirectMissing, NextOperation: "restore",
		Executor: process}
	report := VerifyReceiptsWithFailures(plan, nil, []ObservationFailure{failure})
	if report.Decision != ReceiptDecisionUnknown || report.Reason != ReceiptReasonUnknownIndicator || len(report.Failures) != 1 {
		t.Fatalf("incomplete unknown failure was accepted: %+v", report)
	}
}

func sourcepolicyPlanForFailureTest() Plan {
	return actionableReceiptPlan()
}
