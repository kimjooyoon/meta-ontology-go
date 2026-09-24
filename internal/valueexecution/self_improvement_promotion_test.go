package valueexecution

import "testing"

func TestObserveSelfImprovementPromotionRequiresDecision(t *testing.T) {
	observation := ObserveSelfImprovementPromotion(SelfImprovementEvidenceLedger{})
	if observation.Status != SelfImprovementPromotionStatusInsufficientEvidence {
		t.Fatalf("status=%q, want insufficient evidence", observation.Status)
	}
	if observation.Reason != "LEDGER_DECISION_MISSING" {
		t.Fatalf("reason=%q, want missing decision", observation.Reason)
	}
	if !observation.NonAuthorizing {
		t.Fatal("promotion observation must remain non-authorizing")
	}
	if observation.LedgerDigest == "" || observation.Digest == "" {
		t.Fatal("promotion observation must retain both digests")
	}
}
