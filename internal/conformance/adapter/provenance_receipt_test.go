package adapter

import (
	"bytes"
	"strings"
	"testing"
)

func TestProvenanceReceiptCanonicalRoundTrip(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	payload, err := receipt.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := ParseProvenanceReceipt(payload)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Schema != ProvenanceReceiptSchema || decoded.EventRef != receipt.EventRef {
		t.Fatalf("canonical round trip changed receipt identity: %+v", decoded)
	}
	second, err := decoded.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(payload, second) {
		t.Fatal("receipt serialization is not deterministic")
	}
	firstDigest, err := receipt.Digest()
	if err != nil {
		t.Fatal(err)
	}
	secondDigest, err := decoded.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if firstDigest != secondDigest {
		t.Fatalf("receipt digest changed after round trip: %q != %q", firstDigest, secondDigest)
	}
}

func TestProvenanceReceiptObservedNoWrite(t *testing.T) {
	receipt, request, observation := newCancelledReceipt(t)
	if err := receipt.ValidateObservedNoWrite(request, &observation); err != nil {
		t.Fatal(err)
	}
	if receipt.WriteEffect != ReceiptWriteEffectNone {
		t.Fatalf("cancelled receipt reports a write effect: %q", receipt.WriteEffect)
	}
}

func TestProvenanceReceiptDeferredSixJobTuple(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	receipt.ProvenanceStatus = ReceiptProvenanceDeferred
	receipt.Jobs[len(receipt.Jobs)-1].Status = "queued"
	receipt.Jobs[len(receipt.Jobs)-1].Conclusion = ""
	if err := receipt.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestProvenanceReceiptAppendIsImmutable(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	digest := digestBytes([]byte("previous evidence"))
	updated, err := receipt.AppendPredecessor(ReceiptPredecessor{EventRef: "event-001", Digest: digest})
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.Predecessors) != 0 || len(updated.Predecessors) != 1 {
		t.Fatalf("append mutated source or omitted predecessor: %d %d", len(receipt.Predecessors), len(updated.Predecessors))
	}
	if _, err := updated.AppendPredecessor(ReceiptPredecessor{EventRef: "event-000", Digest: digest}); err == nil {
		t.Fatal("out-of-order predecessor was accepted")
	}
}

func newCancelledReceipt(t *testing.T) (ProvenanceReceipt, Request, NoWriteObservation) {
	t.Helper()
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	observer := newStableObserver(t, request)
	observation, err := observer.CaptureRejected(RejectionCancelled)
	if err != nil {
		t.Fatal(err)
	}
	digests, err := observation.StateDigests(request)
	if err != nil {
		t.Fatal(err)
	}
	head := strings.Repeat("b", 40)
	return ProvenanceReceipt{
		Schema: ProvenanceReceiptSchema, Repository: "caller/repository",
		BaseSHA: strings.Repeat("a", 40), HeadSHA: head,
		EventRef: "event-002", CheckoutRef: "refs/pull/104/merge",
		Run: "run-002", Attempt: 1, Jobs: successfulReceiptJobs(head),
		Binding:            requestObservationBinding(request),
		PreconditionDigest: digestBytes([]byte("precondition")),
		BeforeStateDigest:  digests.Before, AfterStateDigest: digests.After,
		Outcome: ReceiptOutcomeCancelled, WriteEffect: ReceiptWriteEffectNone,
		Predecessors: []ReceiptPredecessor{}, ProvenanceStatus: ReceiptProvenanceVerified,
	}, request, observation
}

func successfulReceiptJobs(head string) []ReceiptJob {
	jobs := make([]ReceiptJob, 0, len(receiptJobNames))
	for _, name := range receiptJobNames {
		jobs = append(jobs, ReceiptJob{Name: name, HeadSHA: head, Status: "completed", Conclusion: "success"})
	}
	return jobs
}
