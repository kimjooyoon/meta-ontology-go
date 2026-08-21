package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestProofBundleValidatesAndPreservesReceiptSchema(t *testing.T) {
	bundle := validProof()
	if err := validateProof(bundle); err != nil {
		t.Fatal(err)
	}
	receipt := makeReceipt(bundle, contextInput{})
	if receipt.Schema != receiptSchema || receipt.Relation != "conformance" || receipt.Repository != bundle.Repository {
		t.Fatal("receipt schema or relation changed")
	}
}
func TestOldProofAndReceiptSchemasFailClosed(t *testing.T) {
	bundle := validProof()
	bundle.Schema = "gooo/ci-proof/v2"
	if err := validateProof(bundle); err == nil {
		t.Fatal("old proof schema was accepted after GuardianEvidence contract migration")
	}
	bundle = validProof()
	receipt := makeReceipt(bundle, contextInput{})
	receipt.Schema = "gooo/provenance-receipt/v2"
	filename := writeReceiptFixture(t, receipt)
	if err := verifyReceipt(filename, bundle); err == nil {
		t.Fatal("old receipt schema was accepted after GuardianEvidence contract migration")
	}
}
func TestProofUnknownFieldsFailClosed(t *testing.T) {
	bundle := validProof()
	data, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	document["legacy_guardian_evidence"] = map[string]any{"decision": "PASS"}
	data, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	filename := t.TempDir() + "/proof.json"
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readStrictJSON[proofBundle](filename); err == nil {
		t.Fatal("proof unknown field was accepted")
	}
}
