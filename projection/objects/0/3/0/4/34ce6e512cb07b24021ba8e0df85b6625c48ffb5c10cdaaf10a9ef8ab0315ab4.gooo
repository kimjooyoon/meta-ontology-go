package adapter

import (
	"encoding/json"
	"testing"
)

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
