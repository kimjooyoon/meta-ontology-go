package adapter

import (
	"bytes"
	"strings"
	"testing"
)

func TestProvenanceReceiptTamperRejectsCanonicalPayload(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	payload, err := receipt.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	tampered := bytes.Replace(payload, []byte(receipt.HeadSHA), []byte(strings.Repeat("c", 40)), 1)
	if _, err := ParseProvenanceReceipt(tampered); err == nil {
		t.Fatal("tampered head was accepted")
	}
}
func TestProvenanceReceiptStaleJobHeadRejects(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	receipt.Jobs[0].HeadSHA = strings.Repeat("c", 40)
	if err := receipt.Validate(); err == nil {
		t.Fatal("stale job head was accepted")
	}
}
func TestProvenanceReceiptRejectsDeferredPrematureConclusion(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	receipt.ProvenanceStatus = ReceiptProvenanceDeferred
	receipt.Jobs[0].Status = "queued"
	receipt.Jobs[0].Conclusion = "success"
	if err := receipt.Validate(); err == nil {
		t.Fatal("deferred job with a premature conclusion was accepted")
	}
}
func TestProvenanceReceiptRejectsDuplicateTerminalSuccess(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	receipt.Jobs[1].Name = receipt.Jobs[0].Name
	if err := receipt.Validate(); err == nil {
		t.Fatal("duplicate terminal-success job was accepted")
	}
}
func TestProvenanceReceiptRejectsNonCanonicalJobOrder(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	receipt.Jobs[0], receipt.Jobs[1] = receipt.Jobs[1], receipt.Jobs[0]
	if err := receipt.Validate(); err == nil {
		t.Fatal("non-canonical job order was accepted")
	}
}
func TestProvenanceReceiptReplayAndOrderReject(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	digest := digestBytes([]byte("previous evidence"))
	receipt.Predecessors = []ReceiptPredecessor{
		{EventRef: "event-003", Digest: digest},
		{EventRef: "event-001", Digest: digest},
	}
	if err := receipt.Validate(); err == nil {
		t.Fatal("out-of-order predecessors were accepted")
	}
	receipt.Predecessors = []ReceiptPredecessor{{EventRef: receipt.EventRef, Digest: digest}}
	if err := receipt.Validate(); err == nil {
		t.Fatal("current event replay was accepted")
	}
}
func TestProvenanceReceiptNoWriteRejectsObservedWrite(t *testing.T) {
	receipt, _, _ := newCancelledReceipt(t)
	receipt.WriteEffect = ReceiptWriteEffectObserved
	if err := receipt.Validate(); err == nil {
		t.Fatal("cancelled receipt accepted a write effect")
	}
}
