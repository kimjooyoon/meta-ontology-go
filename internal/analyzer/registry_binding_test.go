package analyzer

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
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

func TestDuplicateRegistryRegistrationIsNoMutation(t *testing.T) {
	registry := NewRegistry()
	entry := Registration{
		Ref:  SymbolRef{PackagePath: "billing", PackageName: "billing", Name: "Order"},
		Kind: KindEntity, Identity: NewIdentity("billing", "billing://entity/order"),
	}
	if err := registry.Register(entry); err != nil {
		t.Fatal(err)
	}
	canonical := registry.Canonical()
	digest := registry.Digest()
	duplicate := entry
	duplicate.Span = Span{Filename: "different.go", Start: Position{Offset: 9}, End: Position{Offset: 12}}
	if err := registry.Register(duplicate); err != nil {
		t.Fatal(err)
	}
	if registry.Canonical() != canonical || registry.Digest() != digest || len(registry.all()) != 1 {
		t.Fatal("duplicate registration changed registry identity")
	}
}

func TestRegistryRejectsMalformedIdentityAndNamespaceWithoutMutation(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		identity    Identity
		wantMessage string
	}{
		{name: "invalid identity", identity: NewIdentity("billing", "Order"), wantMessage: "semantic identity"},
		{name: "empty namespace", identity: NewIdentity("", "billing://entity/order"), wantMessage: "semantic namespace"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			registry := NewRegistry()
			beforeCanonical, beforeDigest := registry.Canonical(), registry.Digest()
			err := registry.Register(Registration{
				Ref:  SymbolRef{PackagePath: "billing", PackageName: "billing", Name: "Order"},
				Kind: KindEntity, Identity: testCase.identity,
			})
			if err == nil || !strings.Contains(err.Error(), testCase.wantMessage) {
				t.Fatalf("malformed registration error = %v, want %q", err, testCase.wantMessage)
			}
			if registry.Canonical() != beforeCanonical || registry.Digest() != beforeDigest || len(registry.all()) != 0 {
				t.Fatal("malformed registration changed registry state")
			}
		})
	}
}

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
