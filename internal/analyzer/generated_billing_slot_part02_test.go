package analyzer

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestGeneratedBillingProtectedSlotMutationChangesOnlyDeferredObservation(t *testing.T) {
	source := generatedBillingSource(t)
	originalBytes := append([]byte(nil), source.Source...)
	policy := generatedBillingPolicy(t)
	first := adaptGeneratedBillingSource(t, source, generatedBillingRegistry(t), policy)
	mutated := mutateGeneratedBillingSlotBody(t, source)
	second := adaptGeneratedBillingSource(t, mutated, generatedBillingRegistry(t), policy)
	comparison := semantic.CompareIR(first.IR, second.IR)
	if !comparison.SemanticEqual {
		t.Fatalf("slot body mutation changed authoritative contract facts: %#v", comparison)
	}
	if first.SlotObservations[0].BodyDigest == second.SlotObservations[0].BodyDigest ||
		first.SlotObservations[0].Fingerprint() == second.SlotObservations[0].Fingerprint() ||
		first.SlotObservationDigest == second.SlotObservationDigest ||
		first.ImplementationObservationDigest == second.ImplementationObservationDigest ||
		first.BindingDigest == second.BindingDigest ||
		first.NormalizedDelta.Digest == second.NormalizedDelta.Digest {
		t.Fatal("slot body mutation did not change deferred observation evidence")
	}
	before := irSnapshot(first.IR)
	reconcile := ReconcileSemantic(second, first.IR, first.SourceDigest, first.PolicyDigest,
		first.ToolchainDigest, first.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.WriteEffect != ReconcileNoWrite ||
		reconcile.FailureCode != "source-or-observation-mismatch" || reconcile.SourceMatch ||
		reconcile.ObservationMatch {
		t.Fatalf("slot body mutation reconcile = %#v, want fail-closed no-write", reconcile)
	}
	if got := irSnapshot(first.IR); got != before {
		t.Fatalf("rejected slot body mutation changed IR: before=%q after=%q", before, got)
	}
	if !bytes.Equal(source.Source, originalBytes) {
		t.Fatal("slot analysis mutated the original generated source bytes")
	}
}
