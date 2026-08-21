package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
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
