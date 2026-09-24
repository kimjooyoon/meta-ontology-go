package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestObserveExecutionEvidenceReceiptClosureCompletesOnlyWithFullEvidence(t *testing.T) {
	prefix := ObserveExecutionEvidencePrefix([]string{
		"declaration-digest",
		"source-digest",
		"ir-digest",
		"generated-digest",
		"reverse-digest",
	})
	receipt := valueexecution.ExecutionOriginReceipt{
		Status:          valueexecution.ExecutionOriginStatusBound,
		Phase:           valueexecution.ExecutionPhaseCompleted,
		ReceiptDigest:   "receipt-digest",
		ExecutionDigest: "execution-digest",
	}
	observation := ObserveExecutionEvidenceReceiptClosure(prefix, receipt)
	if observation.Status != ExecutionEvidenceReceiptClosureComplete {
		t.Fatalf("status=%q, want complete", observation.Status)
	}
	if observation.MissingStageIndex != -1 || !observation.NonAuthorizing {
		t.Fatalf("completion boundary=%+v", observation)
	}
	if !ValidateExecutionEvidenceReceiptClosureObservation(observation, prefix, receipt) {
		t.Fatal("complete closure observation did not validate")
	}
}

func TestObserveExecutionEvidenceReceiptClosurePreservesUnknownBoundary(t *testing.T) {
	prefix := ObserveExecutionEvidencePrefix([]string{
		"declaration-digest",
		"",
		"ir-digest",
		"generated-digest",
	})
	receipt := valueexecution.ExecutionOriginReceipt{
		Status:          valueexecution.ExecutionOriginStatusBound,
		Phase:           valueexecution.ExecutionPhaseCompleted,
		ReceiptDigest:   "receipt-digest",
		ExecutionDigest: "execution-digest",
	}
	observation := ObserveExecutionEvidenceReceiptClosure(prefix, receipt)
	if observation.Status != ExecutionEvidenceReceiptClosureUnknown ||
		observation.MissingStageIndex != 1 ||
		observation.Reason != "EXECUTION_EVIDENCE_PREFIX_INCOMPLETE" {
		t.Fatalf("unknown boundary=%+v", observation)
	}
}

func TestValidateExecutionEvidenceReceiptClosureObservationRejectsTamper(t *testing.T) {
	prefix := ObserveExecutionEvidencePrefix([]string{
		"declaration-digest",
		"source-digest",
		"ir-digest",
	})
	receipt := valueexecution.ExecutionOriginReceipt{
		Status:          valueexecution.ExecutionOriginStatusBound,
		Phase:           valueexecution.ExecutionPhaseCompleted,
		ReceiptDigest:   "receipt-digest",
		ExecutionDigest: "execution-digest",
	}
	observation := ObserveExecutionEvidenceReceiptClosure(prefix, receipt)
	observation.EvidencePrefixDigest = "tampered-prefix"
	if ValidateExecutionEvidenceReceiptClosureObservation(observation, prefix, receipt) {
		t.Fatal("tampered prefix digest unexpectedly validated")
	}
}
