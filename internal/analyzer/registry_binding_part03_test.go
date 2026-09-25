package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func TestRegistryBindingTamperRejectsWithoutIRMutation(t *testing.T) {
	source := generatedBillingSource(t)
	policy := generatedBillingPolicy(t)
	registry := generatedBillingRegistry(t)
	base := adaptGeneratedBillingSource(t, source, registry, policy)
	cases := []struct {
		name             string
		mutate           func(*SemanticAdapterResult)
		wantBindingMatch bool
	}{
		{name: "result digest", wantBindingMatch: false, mutate: func(result *SemanticAdapterResult) {
			result.RegistryDigest = semantic.StableHashString("tampered-registry")
		}},
		{name: "delta binding", wantBindingMatch: true, mutate: func(result *SemanticAdapterResult) {
			result.NormalizedDelta.SignatureFacts[0].Binding.RegistryDigest =
				semantic.StableHashString("tampered-binding")
		}},
		{name: "binding digest", wantBindingMatch: false, mutate: func(result *SemanticAdapterResult) {
			result.BindingDigest = semantic.StableHashString("tampered-binding-digest")
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			observed := base
			testCase.mutate(&observed)
			before := irSnapshot(observed.IR)
			reconcile := ReconcileSemanticWithRegistry(observed, observed.IR, observed.SourceDigest,
				observed.PolicyDigest, observed.ToolchainDigest, registry.Digest(),
				observed.ImplementationObservationDigest)
			if reconcile.Accepted || reconcile.DeltaValid || reconcile.BindingMatch != testCase.wantBindingMatch ||
				reconcile.WriteEffect != ReconcileNoWrite ||
				reconcile.FailureCode != "invalid-delta-binding" {
				t.Fatalf("tampered registry reconcile = %#v, want invalid no-write", reconcile)
			}
			if got := irSnapshot(observed.IR); got != before {
				t.Fatalf("tamper rejection changed IR: before=%q after=%q", before, got)
			}
			if !reflect.DeepEqual(observed.IR.Graph.Nodes(), base.IR.Graph.Nodes()) {
				t.Fatal("tamper rejection changed graph nodes")
			}
		})
	}
}
func TestDeferredObservationOnlyDeltaRejectsTamperedRegistryDigest(t *testing.T) {
	source := generatedBillingSource(t)
	policy := generatedBillingPolicy(t)
	registry := generatedBillingRegistry(t)
	base := adaptGeneratedBillingSource(t, source, registry, policy)
	observed := base
	observed.NormalizedDelta.SignatureFacts = nil
	observed.NormalizedDelta.CandidateFacts = nil
	observed.NormalizedDelta.DeferredDetails = nil
	observed.NormalizedDelta.DeferredSlots = nil
	observed.NormalizedDelta.DeferredImplementation = append([]ImplementationObservation(nil),
		base.NormalizedDelta.DeferredImplementation...)
	observed.NormalizedDelta.Digest = observed.NormalizedDelta.StableHash()
	observed.RegistryDigest = semantic.StableHashString("tampered-observation-registry")
	assertLegacyRegistryTamperNoWrite(t, observed)
}
