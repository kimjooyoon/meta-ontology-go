package adapter

import (
	"testing"
)

func TestObserverWorkflowCaptureIsImmutable(t *testing.T) {
	receipt, request, _ := newCancelledReceipt(t)
	observer := newBareObserver(t, request)
	workflow := workflowForReceipt(receipt)
	evidence, err := newVerifiedWorkflowEvidence(workflow)
	if err != nil {
		t.Fatal(err)
	}
	if err := observer.captureVerifiedWorkflow(evidence); err != nil {
		t.Fatal(err)
	}
	assertReceiptOracleError(t, observer.captureVerifiedWorkflow(evidence), OracleNW003)
}
func TestPublicWorkflowCaptureCannotInjectVerifiedTuple(t *testing.T) {
	request := sampleRequest(StatusFail)
	observer := newBareObserver(t, request)
	assertReceiptOracleError(t, observer.CaptureUnverifiedWorkflow(verifiedTestWorkflow(request)), OracleNW003)
	fake := verifiedTestWorkflow(request)
	fake.Status = WorkflowEvidenceUnverified
	if err := observer.CaptureUnverifiedWorkflow(fake); err != nil {
		t.Fatal(err)
	}
	observation, err := observer.CaptureRejected(RejectionCancelled)
	if err != nil {
		t.Fatal(err)
	}
	assertReceiptOracleError(t, observation.VerifyNoWrite(request), OracleNW003)
}
func receiptWithWorkflowStatus(t *testing.T, status WorkflowEvidenceStatus) (ProvenanceReceipt, Request, NoWriteObservation) {
	t.Helper()
	receipt, request, _ := newCancelledReceipt(t)
	observer := newBareObserver(t, request)
	workflow := workflowForReceipt(receipt)
	workflow.Status = status
	if err := observer.CaptureUnverifiedWorkflow(workflow); err != nil {
		t.Fatal(err)
	}
	observation, err := observer.CaptureRejected(RejectionCancelled)
	if err != nil {
		t.Fatal(err)
	}
	return receipt, request, observation
}
func workflowForReceipt(receipt ProvenanceReceipt) WorkflowBinding {
	return WorkflowBinding{
		Status: WorkflowEvidenceVerified, Repository: receipt.Repository,
		BaseSHA: receipt.BaseSHA, HeadSHA: receipt.HeadSHA,
		EventRef: receipt.EventRef, CheckoutRef: receipt.CheckoutRef,
		Run: receipt.Run, Attempt: receipt.Attempt,
		ArtifactCount: receipt.ArtifactCount,
		Jobs:          append([]ReceiptJob{}, receipt.Jobs...),
	}
}
