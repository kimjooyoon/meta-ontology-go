package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestRegistryDigestIsStableForNilEmptyAndPermutation(t *testing.T) {
	var nilRegistry *Registry
	empty := NewRegistry()
	if nilRegistry.Canonical() != empty.Canonical() || nilRegistry.Digest() != empty.Digest() {
		t.Fatal("nil and empty registries have different canonical identities")
	}
	first := generatedBillingRegistry(t)
	permuted := reversedGeneratedBillingRegistry(t)
	if first.Canonical() != permuted.Canonical() || first.Digest() != permuted.Digest() {
		t.Fatal("registry digest changed with registration order")
	}
	if !validDigest(first.Digest()) || first.Digest() != semantic.StableHashString(first.Canonical()) {
		t.Fatalf("registry digest is not schema-bound: %q", first.Digest())
	}
}
func TestGeneratedBillingRegistryMutationIsBoundAndNoWrite(t *testing.T) {
	source := generatedBillingSource(t)
	policy := generatedBillingPolicy(t)
	registry := generatedBillingRegistry(t)
	first := adaptGeneratedBillingSource(t, source, registry, policy)
	permuted := adaptGeneratedBillingSource(t, source, reversedGeneratedBillingRegistry(t), policy)
	if first.RegistryDigest != registry.Digest() || first.RegistryDigest != permuted.RegistryDigest ||
		first.NormalizedDelta.Digest != permuted.NormalizedDelta.Digest ||
		first.BindingDigest != permuted.BindingDigest {
		t.Fatal("registry replay or permutation changed a semantically equivalent handoff")
	}

	mutatedRegistry := generatedBillingRegistryWithOrder(t, "billing://entity/wrong-order")
	mutated := adaptGeneratedBillingSource(t, source, mutatedRegistry, policy)
	if mutated.RegistryDigest == first.RegistryDigest || mutated.NormalizedDelta.Digest == first.NormalizedDelta.Digest ||
		mutated.BindingDigest == first.BindingDigest {
		t.Fatal("registry identity mutation was not reflected in bound digests")
	}
	if semantic.CompareIR(mutated.IR, first.IR).SemanticEqual {
		t.Fatal("registry identity mutation retained authoritative semantic equality")
	}

	before := irSnapshot(first.IR)
	reconcile := ReconcileSemanticWithRegistry(mutated, first.IR, first.SourceDigest, first.PolicyDigest,
		first.ToolchainDigest, first.RegistryDigest, first.ImplementationObservationDigest)
	if reconcile.Accepted || reconcile.RegistryMatch || reconcile.WriteEffect != ReconcileNoWrite ||
		reconcile.FailureCode != "registry-mismatch" {
		t.Fatalf("registry mutation reconcile = %#v, want registry no-write rejection", reconcile)
	}
	if got := irSnapshot(first.IR); got != before {
		t.Fatalf("rejected registry mutation changed expected IR: before=%q after=%q", before, got)
	}
}
