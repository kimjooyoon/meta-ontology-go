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
