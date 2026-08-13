package adapter

import (
	"strings"
	"testing"
)

func TestProvenanceReceiptObservedCloseNoWrite(t *testing.T) {
	receipt, request, observation := newClosedReceipt(t)
	if err := receipt.ValidateObservedNoWrite(request, &observation); err != nil {
		t.Fatal(err)
	}
	if observation.Reason != RejectionClosed || receipt.WriteEffect != ReceiptWriteEffectNone {
		t.Fatalf("close receipt was not bound to no-write evidence: %+v %+v", observation, receipt)
	}
}

func TestProvenanceReceiptRejectsMissingWorkflowEvidence(t *testing.T) {
	receipt, request, _ := newCancelledReceipt(t)
	observer := newBareObserver(t, request)
	observation, err := observer.CaptureRejected(RejectionCancelled)
	if err != nil {
		t.Fatal(err)
	}
	assertReceiptOracleError(t, receipt.ValidateObservedNoWrite(request, &observation), OracleNW001)
}

func TestProvenanceReceiptRejectsUnverifiedWorkflowEvidence(t *testing.T) {
	receipt, request, observation := receiptWithWorkflowStatus(t, WorkflowEvidenceUnverified)
	assertReceiptOracleError(t, receipt.ValidateObservedNoWrite(request, &observation), OracleNW003)
}

func TestProvenanceReceiptRejectsFakeAndStaleWorkflowTuple(t *testing.T) {
	receipt, request, observation := newCancelledReceipt(t)
	receipt.Repository = "attacker/repository"
	assertReceiptOracleError(t, receipt.ValidateObservedNoWrite(request, &observation), OracleNW002)

	receipt, request, observation = newCancelledReceipt(t)
	receipt.Binding.Fixture = "relabeled-fixture"
	assertReceiptOracleError(t, receipt.ValidateObservedNoWrite(request, &observation), OracleID001)

	receipt, request, observation = newCancelledReceipt(t)
	receipt.HeadSHA = strings.Repeat("c", 40)
	for index := range receipt.Jobs {
		receipt.Jobs[index].HeadSHA = receipt.HeadSHA
	}
	assertReceiptOracleError(t, receipt.ValidateObservedNoWrite(request, &observation), OracleNW002)
}

func TestProvenanceReceiptRejectsObserverWorkflowRelabel(t *testing.T) {
	receipt, request, observation := newCancelledReceipt(t)
	observation.Workflow.Repository = "relabeled/repository"
	assertReceiptOracleError(t, receipt.ValidateObservedNoWrite(request, &observation), OracleNW003)
}

func TestProvenanceReceiptRejectsRunBindingRelabel(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	receipt.Run = "relabeled-run"
	assertReceiptOracleError(t, receipt.Validate(), OracleID001)
}

func TestProvenanceReceiptRejectsMissingOrZeroArtifact(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	receipt.ArtifactCount = 0
	assertReceiptOracleError(t, receipt.Validate(), OracleNW003)

	request := sampleRequest(StatusFail)
	observer := newBareObserver(t, request)
	workflow := workflowForReceipt(receipt)
	workflow.ArtifactCount = 0
	assertReceiptOracleError(t, observer.CaptureUnverifiedWorkflow(workflow), OracleNW003)
}

func TestProvenanceReceiptRejectsDuplicateWorkflowJob(t *testing.T) {
	receipt, request, _ := newCancelledReceipt(t)
	observer := newBareObserver(t, request)
	workflow := workflowForReceipt(receipt)
	workflow.Jobs[1].Name = workflow.Jobs[0].Name
	assertReceiptOracleError(t, observer.CaptureUnverifiedWorkflow(workflow), OracleNW003)
}

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
