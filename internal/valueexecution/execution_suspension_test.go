package valueexecution

import (
	"testing"
	"time"
)

func TestExecutionSuspensionRoundTripAndTamperBoundary(t *testing.T) {
	suspendedAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	execution := Execution{
		Phase:      ExecutionPhaseRunning,
		Conditions: executionRunningConditions(),
	}

	receipt, err := SuspendExecution(execution, "WAIT_FOR_EXTERNAL_OBSERVATION", suspendedAt)
	if err != nil {
		t.Fatalf("SuspendExecution() error = %v", err)
	}
	if receipt.Source != ExecutionSuspensionSource {
		t.Fatalf("receipt source = %q, want %q", receipt.Source, ExecutionSuspensionSource)
	}
	if receipt.ExecutionDigest == "" || receipt.ReceiptDigest == "" {
		t.Fatal("suspension receipt must carry both provenance digests")
	}

	resumed := receipt.Resume(suspendedAt.Add(time.Minute))
	if resumed.Status != ExecutionResumeStatusResumed {
		t.Fatalf("Resume() status = %q, want %q", resumed.Status, ExecutionResumeStatusResumed)
	}
	if resumed.Reason != "SUSPENSION_RECEIPT_OBSERVED" {
		t.Fatalf("Resume() reason = %q", resumed.Reason)
	}

	receipt.Reason = "TAMPERED"
	tampered := receipt.Resume(suspendedAt.Add(2 * time.Minute))
	if tampered.Status != ExecutionResumeStatusUnknown {
		t.Fatalf("tampered receipt status = %q, want %q", tampered.Status, ExecutionResumeStatusUnknown)
	}
	if tampered.Reason != "SUSPENSION_RECEIPT_DIGEST_MISMATCH" {
		t.Fatalf("tampered receipt reason = %q", tampered.Reason)
	}
}

func TestExecutionSuspensionRejectsTerminalExecution(t *testing.T) {
	_, err := SuspendExecution(
		Execution{Phase: ExecutionPhaseCompleted},
		"WAIT_FOR_EXTERNAL_OBSERVATION",
		time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC),
	)
	if err == nil {
		t.Fatal("SuspendExecution() error = nil, want terminal execution rejection")
	}
}
