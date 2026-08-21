package analyzer

import (
	"testing"
)

const generatedBillingSlotID = "billing://activity/pay-order/implementation"

func TestGeneratedBillingProtectedSlotIsStableAcrossReplayAndOrder(t *testing.T) {
	source := generatedBillingSource(t)
	support := SourceFile{Filename: "support.go", PackagePath: "billing", Source: []byte("package billing\n")}
	sources := []SourceFile{source, support}
	policy := generatedBillingPolicy(t)
	first := adaptGeneratedBillingSources(t, sources, generatedBillingRegistry(t), policy)
	second := adaptGeneratedBillingSources(t, sources, generatedBillingRegistry(t), policy)
	permuted := adaptGeneratedBillingSources(t, []SourceFile{support, source},
		reversedGeneratedBillingRegistry(t), policy)

	for _, result := range []SemanticAdapterResult{first, second, permuted} {
		if len(result.SlotObservations) != 1 || len(result.NormalizedDelta.DeferredSlots) != 1 {
			t.Fatalf("protected slots = %d/%d, want one", len(result.SlotObservations),
				len(result.NormalizedDelta.DeferredSlots))
		}
		slot := result.SlotObservations[0]
		if slot.SlotID != generatedBillingSlotID || slot.Status != ProtectedSlotDeferred ||
			slot.SourceFile != source.Filename || slot.SourceDigest != result.SourceDigest ||
			slot.BaseDigest != result.NormalizedDelta.SignatureFacts[0].Binding.BaseDigest ||
			slot.PolicyDigest != result.PolicyDigest || slot.ToolchainDigest != result.ToolchainDigest ||
			slot.RegistryDigest != result.RegistryDigest ||
			slot.BodySpan.End.Offset <= slot.BodySpan.Start.Offset {
			t.Fatalf("protected slot = %#v", slot)
		}
		if !validDigest(slot.BodyDigest) || !validDigest(slot.Fingerprint()) ||
			!validDigest(result.SlotObservationDigest) {
			t.Fatalf("protected slot digest binding = %#v, digest=%q", slot, result.SlotObservationDigest)
		}
	}
	if first.SourceDigest != second.SourceDigest || first.SourceDigest != permuted.SourceDigest ||
		first.SlotObservations[0].Span != second.SlotObservations[0].Span ||
		first.SlotObservations[0].Span != permuted.SlotObservations[0].Span ||
		first.SlotObservations[0].BodySpan != second.SlotObservations[0].BodySpan ||
		first.SlotObservations[0].BodySpan != permuted.SlotObservations[0].BodySpan ||
		first.SlotObservationDigest != second.SlotObservationDigest ||
		first.SlotObservationDigest != permuted.SlotObservationDigest ||
		first.NormalizedDelta.Digest != second.NormalizedDelta.Digest ||
		first.NormalizedDelta.Digest != permuted.NormalizedDelta.Digest {
		t.Fatal("protected slot handoff changed across replay or source/registry order")
	}
}
