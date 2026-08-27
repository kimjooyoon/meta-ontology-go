package proofchoicejudge

import (
	"encoding/json"
	"testing"
)

func TestJudgeFailsClosedForMalformedAndTamperedReceipts(t *testing.T) {
	for _, data := range []string{"{}", `{"schema":"gooo/proof-choice-algebra-receipt/v1","decision":"PASS"}`} {
		result := Judge([]byte(data))
		if result.Decision != failClosed {
			t.Fatalf("decision = %s for %s", result.Decision, data)
		}
	}
	valid := receipt{
		Schema: schema, Decision: pass, Reason: "PROOF_CHOICES_COMPOSED", Resolution: "EXACT",
		SourcePath: "fixture.gooo", SourceDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		FixedDenom: fixedDenominator,
		Items:      []item{{Kind: "METRIC", ID: "metric", Statement: "three", Choice: "REGRESSION", Producer: "p", Consumer: "c", MetaOperation: "m", Stage: "s", Step: "t", Reason: "r", Numerator: 3, Denominator: 3}},
		Effects:    effects{},
	}
	data, err := json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}
	valid.Digest, err = digest(valid)
	if err != nil {
		t.Fatal(err)
	}
	data, err = json.Marshal(valid)
	if err != nil {
		t.Fatal(err)
	}
	result := Judge(data)
	if result.Decision != pass {
		t.Fatalf("valid receipt decision = %s (%s)", result.Decision, result.Reason)
	}
	valid.Decision = failClosed
	data, _ = json.Marshal(valid)
	if result := Judge(data); result.Decision != failClosed || result.Reason != "RECEIPT_DIGEST_MISMATCH" {
		t.Fatalf("tampered receipt result = %+v", result)
	}
}
