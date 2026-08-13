package analyzer

import (
	"errors"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestGeneratedBillingLocalityEnvelopeUsesCanonicalClosure(t *testing.T) {
	base := localityBase(t, false)
	result := adaptGeneratedBillingWithBase(t, base, generatedBillingRegistry(t))
	if err := ValidateLocalityEnvelope(base, result); err != nil {
		t.Fatalf("validate locality envelope: %v", err)
	}
	wantTouched := []semantic.ID{
		semantic.MustIdentity("billing://activity/pay-order"),
		semantic.MustIdentity("billing://entity/payment"),
		semantic.MustIdentity("billing://entity/payment-method"),
	}
	wantAffected := append(append([]semantic.ID(nil), wantTouched...),
		semantic.MustIdentity("billing://entity/order"))
	if !reflect.DeepEqual(result.Locality.Touched, wantTouched) {
		t.Fatalf("touched = %v, want %v", result.Locality.Touched, wantTouched)
	}
	if !reflect.DeepEqual(result.Locality.Affected, sortedLocalityIDs(wantAffected)) {
		t.Fatalf("affected = %v, want %v", result.Locality.Affected, sortedLocalityIDs(wantAffected))
	}
	if len(result.Locality.PreservedFacts) != 1 {
		t.Fatalf("preserved facts = %d, want one base fact", len(result.Locality.PreservedFacts))
	}
}

func TestGeneratedBillingLocalityExcludesUnrelatedAndCandidateIDs(t *testing.T) {
	base := localityBase(t, true)
	result := adaptGeneratedBillingWithBase(t, base, ambiguousGeneratedBillingRegistry(t))
	localityIDs := append(append([]semantic.ID(nil), result.Locality.Touched...), result.Locality.Affected...)
	unrelated := semantic.MustIdentity("billing://activity/unrelated")
	alternate := semantic.MustIdentity("billing://entity/alternate-order")
	for _, id := range []semantic.ID{unrelated, alternate} {
		if containsLocalityID(localityIDs, id) {
			t.Fatalf("non-authoritative/unrelated ID %s entered locality: %v", id, localityIDs)
		}
	}
	if len(result.NormalizedDelta.CandidateFacts) != 1 || len(result.DeferredCandidates) != 1 {
		t.Fatalf("candidate separation = %d normalized, %d deferred", len(result.NormalizedDelta.CandidateFacts), len(result.DeferredCandidates))
	}
}

func TestGeneratedBillingPartialObservationPreservesBaseFacts(t *testing.T) {
	base := localityBase(t, true)
	result := adaptGeneratedBillingWithBase(t, base, generatedBillingRegistry(t))
	for _, fact := range base.Graph.DeterministicFacts() {
		if !result.IR.Graph.HasFact(fact.Key()) {
			t.Fatalf("partial observation deleted base fact: %v", fact.Key())
		}
	}
	if got := len(result.Locality.PreservedFacts); got != 2 {
		t.Fatalf("preserved base facts = %d, want 2", got)
	}
	if err := ValidateLocalityEnvelope(base, result); err != nil {
		t.Fatalf("partial locality validation: %v", err)
	}
}

func TestGeneratedBillingLocalityReplayHasStableDigest(t *testing.T) {
	base := localityBase(t, true)
	first := adaptGeneratedBillingWithBase(t, base, generatedBillingRegistry(t))
	replay := adaptGeneratedBillingWithBase(t, base, generatedBillingRegistry(t))
	if first.Locality.Digest != replay.Locality.Digest || first.Locality.Canonical() != replay.Locality.Canonical() {
		t.Fatalf("locality changed on replay: %q vs %q", first.Locality.Digest, replay.Locality.Digest)
	}
	if first.BindingDigest != replay.BindingDigest {
		t.Fatalf("binding changed on replay: %q vs %q", first.BindingDigest, replay.BindingDigest)
	}
}

func TestGeneratedBillingLocalityRejectsMissingTamperedAndRelabeledNoWrite(t *testing.T) {
	base := localityBase(t, true)
	result := adaptGeneratedBillingWithBase(t, base, generatedBillingRegistry(t))
	tests := map[string]func(*LocalityEnvelope){
		"missing": func(envelope *LocalityEnvelope) { *envelope = LocalityEnvelope{} },
		"tampered": func(envelope *LocalityEnvelope) {
			envelope.Affected = append([]semantic.ID(nil), envelope.Affected...)
			envelope.Affected[0] = semantic.MustIdentity("billing://entity/tampered")
			envelope.Digest = envelope.StableHash()
		},
		"relabeled": func(envelope *LocalityEnvelope) {
			touched := append([]semantic.ID(nil), envelope.Touched...)
			envelope.Touched = append([]semantic.ID(nil), envelope.Affected...)
			envelope.Affected = touched
			envelope.Digest = envelope.StableHash()
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			beforeBase, beforeResult := irSnapshot(base), irSnapshot(result.IR)
			mutated := result
			mutated.Locality.Touched = append([]semantic.ID(nil), result.Locality.Touched...)
			mutated.Locality.Affected = append([]semantic.ID(nil), result.Locality.Affected...)
			mutate(&mutated.Locality)
			err := ValidateLocalityEnvelope(base, mutated)
			var adapterErr AdapterError
			if !errors.As(err, &adapterErr) || adapterErr.Code != AdapterLocalityConfig ||
				adapterErr.WriteEffect != ReconcileNoWrite {
				t.Fatalf("error = %v, want locality-config/no-write", err)
			}
			if got := irSnapshot(base); got != beforeBase {
				t.Fatalf("base changed after rejection: before=%q after=%q", beforeBase, got)
			}
			if got := irSnapshot(result.IR); got != beforeResult {
				t.Fatalf("result changed after rejection: before=%q after=%q", beforeResult, got)
			}
		})
	}
}

func localityBase(t *testing.T, includeUnrelated bool) semantic.IR {
	t.Helper()
	base := semantic.NewIR("billing", semantic.Namespace("billing"))
	activity := mustBillingNode(t, semantic.Activity, "billing://activity/pay-order", "PayOrder")
	order := mustBillingNode(t, semantic.Entity, "billing://entity/order", "Order")
	for _, node := range []semantic.Node{activity, order} {
		if err := base.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := base.AddFact(semantic.NewUsedFact(activity.ID, order.ID)); err != nil {
		t.Fatal(err)
	}
	if !includeUnrelated {
		return base
	}
	unrelatedActivity := mustBillingNode(t, semantic.Activity, "billing://activity/unrelated", "Unrelated")
	unrelatedEntity := mustBillingNode(t, semantic.Entity, "billing://entity/unrelated", "UnrelatedEntity")
	for _, node := range []semantic.Node{unrelatedActivity, unrelatedEntity} {
		if err := base.AddNode(node); err != nil {
			t.Fatal(err)
		}
	}
	if err := base.AddFact(semantic.NewUsedFact(unrelatedActivity.ID, unrelatedEntity.ID)); err != nil {
		t.Fatal(err)
	}
	return base
}

func adaptGeneratedBillingWithBase(t *testing.T, base semantic.IR, registry *Registry) SemanticAdapterResult {
	t.Helper()
	result, err := AnalyzeAndAdaptSemantic(SourceSemanticAdapterInput{
		Base: base, Sources: []SourceFile{generatedBillingSource(t)}, Registry: registry,
		Policy: generatedBillingPolicy(t), Producer: semantic.GoHostedCompilerID,
		EvidenceKind: semantic.CompilerRunEvidence, ToolchainIdentity: billingToolchain(),
	})
	if err != nil {
		t.Fatalf("adapt generated billing locality fixture: %v", err)
	}
	return result
}

func containsLocalityID(ids []semantic.ID, want semantic.ID) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}
