package pressureindependence

import (
	"reflect"
	"testing"
)

func TestPermutationAndExpectedMutationInvariants(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	input := mustCorpusInput(t, "two-independent-groups-pass")
	permuted := input
	permuted.PressureRecords = append([]PressureRecord(nil), input.PressureRecords...)
	permuted.RequiredPressureIDs = append([]string(nil), input.RequiredPressureIDs...)
	permuted.GuardIDs = append([]string(nil), input.GuardIDs...)
	permuted.FinitePathIDs = append([]string(nil), input.FinitePathIDs...)
	reverseRecords(permuted.PressureRecords)
	reverseStrings(permuted.RequiredPressureIDs)
	reverseStrings(permuted.GuardIDs)
	reverseStrings(permuted.FinitePathIDs)
	if CanonicalInputDigest(input) != CanonicalInputDigest(permuted) ||
		!reflect.DeepEqual(Evaluate(input), Evaluate(permuted)) {
		t.Fatal("permutation changed input digest or output")
	}
	left := Evaluate(input)
	right := left
	right.Decision, right.Reason = DecisionFailClosed, ReasonPredicateFalse
	if CanonicalInputDigest(input) != CanonicalInputDigest(input) || left.InputDigest != right.InputDigest {
		t.Fatal("expected-only mutation changed input digest")
	}
	if CanonicalOutputDigest(left) == CanonicalOutputDigest(right) {
		t.Fatal("output mutation did not change output digest")
	}
	mutatedExpected := corpus.Cases[0]
	mutatedExpected.Expected.Decision = DecisionFailClosed
	if !reflect.DeepEqual(Evaluate(mutatedExpected.Input), Evaluate(corpus.Cases[0].Input)) {
		t.Fatal("expected-only mutation changed oracle result")
	}
}
func TestGroupMutationChangesResult(t *testing.T) {
	input := mustCorpusInput(t, "two-independent-groups-pass")
	mutated := input
	mutated.PressureRecords = append([]PressureRecord(nil), input.PressureRecords...)
	mutated.PressureRecords[1].IndependenceGroupID = mutated.PressureRecords[0].IndependenceGroupID
	left, stale := Evaluate(input), Evaluate(mutated)
	if left.Decision != DecisionPass || stale.Decision != DecisionUnknown || stale.Reason != ReasonStaleDigest {
		t.Fatalf("group mutation left=%#v stale=%#v", left, stale)
	}
	contracts, ok := readArtifactContracts()
	if !ok {
		t.Fatal("artifact contracts invalid")
	}
	mutated.RegistryDigest = registryBindingDigest(mutated.PressureRecords, contracts.registry)
	right := Evaluate(mutated)
	if right.Decision != DecisionUnknown || right.Reason != ReasonIndependentGroupShortfall ||
		left.InputDigest == right.InputDigest {
		t.Fatalf("rebound group mutation left=%#v right=%#v", left, right)
	}
}
