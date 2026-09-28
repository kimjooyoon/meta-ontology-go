package valueexecution

import "testing"

func TestObserveSelfImprovementPromotionBindingFailsClosed(t *testing.T) {
	observation := ObserveSelfImprovementPromotionBinding(
		SelfImprovementPromotionObservation{},
		ExecutionOriginReceipt{},
	)
	if observation.Status != SelfImprovementPromotionBindingStatusUnknown {
		t.Fatalf("status=%q, want unknown", observation.Status)
	}
	if observation.Reason != "PROMOTION_STATUS_UNKNOWN" {
		t.Fatalf("reason=%q, want unknown promotion", observation.Reason)
	}
	if !observation.NonAuthorizing {
		t.Fatal("promotion binding must remain non-authorizing")
	}
	if observation.Digest == "" {
		t.Fatal("promotion binding must retain its digest")
	}
}
