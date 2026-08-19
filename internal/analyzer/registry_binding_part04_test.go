package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestDeferredSlotOnlyDeltaRejectsTamperedRegistryDigest(t *testing.T) {
	source := generatedBillingSource(t)
	policy := generatedBillingPolicy(t)
	registry := generatedBillingRegistry(t)
	base := adaptGeneratedBillingSource(t, source, registry, policy)
	observed := base
	observed.NormalizedDelta.SignatureFacts = nil
	observed.NormalizedDelta.CandidateFacts = nil
	observed.NormalizedDelta.DeferredImplementation = nil
	observed.NormalizedDelta.DeferredDetails = nil
	observed.NormalizedDelta.DeferredSlots = append([]ProtectedSlotObservation(nil),
		base.NormalizedDelta.DeferredSlots...)
	observed.NormalizedDelta.Digest = observed.NormalizedDelta.StableHash()
	observed.RegistryDigest = semantic.StableHashString("tampered-slot-registry")
	assertLegacyRegistryTamperNoWrite(t, observed)
}
func assertLegacyRegistryTamperNoWrite(t *testing.T, observed SemanticAdapterResult) {
	t.Helper()
	before := irSnapshot(observed.IR)
	reconcile := ReconcileSemantic(observed, observed.IR, observed.SourceDigest, observed.PolicyDigest,
		observed.ToolchainDigest, observed.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.DeltaValid || reconcile.WriteEffect != ReconcileNoWrite ||
		reconcile.FailureCode != "invalid-delta-binding" {
		t.Fatalf("observation-only registry tamper = %#v, want invalid no-write", reconcile)
	}
	if got := irSnapshot(observed.IR); got != before {
		t.Fatalf("registry tamper changed IR: before=%q after=%q", before, got)
	}
}
