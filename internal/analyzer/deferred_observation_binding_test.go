package analyzer

import "testing"

func TestGeneratedBillingDeferredObservationTamperFailsReconcileWithoutWrite(t *testing.T) {
	source := generatedBillingSource(t)
	policy := generatedBillingPolicy(t)
	result := adaptGeneratedBillingSource(t, source, generatedBillingRegistry(t), policy)
	result.NormalizedDelta.DeferredImplementation[0].Span.Filename = "tampered.go"
	result.NormalizedDelta.Digest = result.NormalizedDelta.StableHash()
	result.BindingDigest = semanticAdapterBindingDigest(result)

	reconcile := ReconcileSemantic(result, result.IR, result.SourceDigest, result.PolicyDigest,
		result.ToolchainDigest, result.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.DeltaValid || reconcile.WriteEffect != ReconcileNoWrite ||
		reconcile.FailureCode != "invalid-delta-binding" {
		t.Fatalf("tampered deferred observation reconcile = %#v, want invalid no-write", reconcile)
	}
}
