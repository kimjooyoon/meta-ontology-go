package valueexecution

import "testing"

func TestObserveSelfImprovementPromotionRequiresDecision(t *testing.T) {
	observation := ObserveSelfImprovementPromotion(SelfImprovementEvidenceLedger{})
	if observation.Status != SelfImprovementPromotionStatusUnknown {
		t.Fatalf("status=%q, want unknown", observation.Status)
	}
	if observation.Reason != "LEDGER_UNKNOWN" {
		t.Fatalf("reason=%q, want unknown ledger", observation.Reason)
	}
	if !observation.NonAuthorizing {
		t.Fatal("promotion observation must remain non-authorizing")
	}
	if observation.LedgerDigest == "" || observation.Digest == "" {
		t.Fatal("promotion observation must retain both digests")
	}
}
