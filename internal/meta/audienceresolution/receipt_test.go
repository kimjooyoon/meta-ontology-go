package audienceresolution

import "testing"

func TestIndependentCheckerRejectsForgedDecision(t *testing.T) {
	receipt := Evaluate(fixtureInput(t))
	receipt.Views[0].LocalDecision = "PASS"
	if err := ValidateReceipt(receipt); err == nil {
		t.Fatal("checker accepted a contradictory audience decision")
	}
}

func TestIndependentCheckerRejectsForgedDigest(t *testing.T) {
	receipt := Evaluate(fixtureInput(t))
	receipt.Digest = "sha256:" + "0" + receipt.Digest[8:]
	if err := ValidateReceipt(receipt); err == nil {
		t.Fatal("checker accepted a forged receipt digest")
	}
}
