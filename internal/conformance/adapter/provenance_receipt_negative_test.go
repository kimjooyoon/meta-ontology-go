package adapter

import (
	"bytes"
	"encoding/json"
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

func TestProvenanceReceiptNoWriteRejectsForgedState(t *testing.T) {
	receipt, request, observation := newCancelledReceipt(t)
	receipt.AfterStateDigest = digestBytes([]byte("forged state"))
	err := receipt.ValidateObservedNoWrite(request, &observation)
	assertReceiptOracleError(t, err, OracleNW003)
}

func TestProvenanceReceiptRejectsMissingObserverAndProducerClaim(t *testing.T) {
	receipt, request, _ := newCancelledReceipt(t)
	assertReceiptOracleError(t, receipt.ValidateObservedNoWrite(request, nil), OracleNW001)
	claim := true
	response := sampleResponse(StatusFail, false)
	response.ProducerClaims.NoWrite = &claim
	evaluation := Evaluate(request, response)
	assertOracleFailure(t, evaluation, OracleNW001)
	if evaluation.PromotionEligible {
		t.Fatal("producer claim made a receipt path promotable")
	}
}

func TestProvenanceReceiptRejectsWireOnlyObservation(t *testing.T) {
	receipt, request, observation := newCancelledReceipt(t)
	payload, err := json.Marshal(observation)
	if err != nil {
		t.Fatal(err)
	}
	var decoded NoWriteObservation
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	assertReceiptOracleError(t, receipt.ValidateObservedNoWrite(request, &decoded), OracleNW003)
}

func assertReceiptOracleError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected %s error", want)
	}
	evidence, ok := err.(NoWriteEvidenceError)
	if !ok || evidence.Code != want {
		t.Fatalf("expected %s, got %T %v", want, err, err)
	}
}
