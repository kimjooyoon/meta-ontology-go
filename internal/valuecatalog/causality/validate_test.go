package causality

import (
	"encoding/json"
	"testing"
)

func TestValidateRejectsGraphMutation(t *testing.T) {
	receipt, err := Evaluate(fixtureReport(t, ModeSuccess), ModeSuccess)
	if err != nil {
		t.Fatal(err)
	}
	receipt.Graph.Edges[0].Kind = "MUTATED"
	receipt.Graph.Digest, err = graphDigest(receipt.Graph)
	if err != nil {
		t.Fatal(err)
	}
	receipt.Subject.GraphDigest = receipt.Graph.Digest
	receipt.ReceiptDigest, err = receiptDigest(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if err := Validate(receipt); err == nil {
		t.Fatal("mutated graph accepted")
	}
}

func TestEvaluateRejectsMixedResolution(t *testing.T) {
	data := fixtureReport(t, ModeUnknown)
	var report inputReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	report.OperationClaimTransitions[ClaimTotal].Event = "EVIDENCE_ACCEPTED"
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Evaluate(data, ""); err == nil {
		t.Fatal("mixed resolution accepted")
	}
}
