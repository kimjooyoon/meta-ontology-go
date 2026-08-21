package adapter

import "testing"

func TestProvenanceReceiptReplayGuardAcceptsDistinctEvents(t *testing.T) {
	first, _, _ := newCancelledReceipt(t)
	second := first
	second.EventRef = "event-003"
	guard := NewReceiptReplayGuard()
	if err := guard.Accept(first); err != nil {
		t.Fatal(err)
	}
	if err := guard.Accept(second); err != nil {
		t.Fatal(err)
	}
}

func TestProvenanceReceiptReplayGuardRejectsExactReplay(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	guard := NewReceiptReplayGuard()
	if err := guard.Accept(receipt); err != nil {
		t.Fatal(err)
	}
	assertReceiptOracleError(t, guard.Accept(receipt), OracleNW003)
}

func TestProvenanceReceiptReplayGuardRejectsConflictingEvent(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	conflict := receipt
	conflict.PreconditionDigest = digestBytes([]byte("conflicting precondition"))
	guard := NewReceiptReplayGuard()
	if err := guard.Accept(receipt); err != nil {
		t.Fatal(err)
	}
	assertReceiptOracleError(t, guard.Accept(conflict), OracleNW003)
}

func TestProvenanceReceiptValidateAppendOnlyRejectsPriorReplay(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	assertReceiptOracleError(t, receipt.ValidateAppendOnly([]ProvenanceReceipt{receipt}), OracleNW003)
}

func TestProvenanceReceiptReplayGuardRejectsDuplicatePriorHistory(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	other := receipt
	other.EventRef = "event-003"
	assertReceiptOracleError(t, receipt.ValidateAppendOnly(
		[]ProvenanceReceipt{other, other},
	), OracleNW003)
}

func TestProvenanceReceiptReplayGuardObservedRequiresMutationProof(t *testing.T) {
	receipt, request, _ := newCancelledReceipt(t)
	observer := newBareObserver(t, request)
	attachVerifiedWorkflow(t, observer, request)
	observation, err := observer.CaptureRejected(RejectionCancelled)
	if err != nil {
		t.Fatal(err)
	}
	guard := NewReceiptReplayGuard()
	assertReceiptOracleError(t, guard.AcceptObserved(request, receipt, &observation), OracleNW001)
}

func TestProvenanceReceiptReplayGuardObservedRejectsDuplicate(t *testing.T) {
	receipt, request, observation := newCancelledReceipt(t)
	guard := NewReceiptReplayGuard()
	if err := guard.AcceptObserved(request, receipt, &observation); err != nil {
		t.Fatal(err)
	}
	assertReceiptOracleError(t, guard.AcceptObserved(request, receipt, &observation), OracleNW003)
}
