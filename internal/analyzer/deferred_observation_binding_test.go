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

func TestGeneratedBillingDeferredObservationShapeTamperFailsReconcileWithoutWrite(t *testing.T) {
	base := adaptGeneratedBillingSource(t, generatedBillingSource(t),
		generatedBillingRegistry(t), generatedBillingPolicy(t))
	for _, testCase := range []struct {
		name   string
		mutate func(*ImplementationObservation)
	}{
		{name: "source file", mutate: func(observation *ImplementationObservation) {
			observation.SourceFile = "tampered.go"
		}},
		{name: "negative span", mutate: func(observation *ImplementationObservation) {
			observation.Span.Start.Offset = -1
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			observed := base
			testCase.mutate(&observed.ImplementationObservations[0])
			testCase.mutate(&observed.NormalizedDelta.DeferredImplementation[0])
			observed.ImplementationObservationDigest = implementationObservationDigest(
				observed.ImplementationObservations, observed.SlotObservations,
			)
			observed.NormalizedDelta.Digest = observed.NormalizedDelta.StableHash()
			observed.BindingDigest = semanticAdapterBindingDigest(observed)

			reconcile := ReconcileSemantic(observed, observed.IR, observed.SourceDigest,
				observed.PolicyDigest, observed.ToolchainDigest, observed.ImplementationObservationDigest)
			if reconcile.Accepted || reconcile.DeltaValid || reconcile.WriteEffect != ReconcileNoWrite ||
				reconcile.FailureCode != "invalid-delta-binding" {
				t.Fatalf("shape-tampered observation reconcile = %#v, want invalid no-write", reconcile)
			}
		})
	}
}
