package query

import (
	"reflect"
	"testing"
)

func TestDerivedRulePermutationReplayAndTransitiveCycleBounds(t *testing.T) {
	root, middle, leaf := id("urn:derived:root"), id("urn:derived:middle"), id("urn:derived:leaf")
	facts := []Fact{
		NewFact(root, WasDerivedFrom, middle),
		NewFact(middle, WasDerivedFrom, leaf),
		NewFact(leaf, WasDerivedFrom, root),
	}
	first, second := New(), New()
	for _, fact := range facts {
		assertAdd(t, first, fact)
	}
	for index := len(facts) - 1; index >= 0; index-- {
		assertAdd(t, second, facts[index])
	}
	request := derivedEnvelope(root, RuleDependsOn, LayerDeterministic, 8, 10)
	want, err := first.Execute(request)
	if err != nil {
		t.Fatal(err)
	}
	got, err := second.Execute(request)
	if err != nil {
		t.Fatal(err)
	}
	wantJSON, err := want.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	gotJSON, err := got.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(wantJSON, gotJSON) || want.Hash != got.Hash ||
		want.RequestHash != got.RequestHash {
		t.Fatalf("permuted derived response changed: %s/%s vs %s/%s", wantJSON, want.Hash, gotJSON, got.Hash)
	}
	ruleDigest, err := RuleDependsOn.CanonicalDigest()
	if err != nil || want.Metadata.DerivedRuleDigest != ruleDigest {
		t.Fatalf("rule digest was not canonical: %q/%q", want.Metadata.DerivedRuleDigest, ruleDigest)
	}
	requestDigest, err := request.CanonicalDigest()
	if err != nil || want.RequestHash != requestDigest {
		t.Fatalf("request digest was not canonical: %q/%q", want.RequestHash, requestDigest)
	}
	if len(want.Result.DerivedDeterministic) != 2 {
		t.Fatalf("cycle was not reduced to simple bounded closure: %#v", want.Result)
	}
	for _, row := range want.Result.DerivedDeterministic {
		if row.Subject != root || row.Depth > 8 || row.Status != DerivedFactStatus || row.Predicate != DerivedDependsOn {
			t.Fatalf("invalid closure row: %#v", row)
		}
	}
	for run := 0; run < 2; run++ {
		replay, err := first.Execute(request)
		if err != nil || replay.Hash != want.Hash {
			t.Fatalf("derived replay changed on run %d: %#v %v", run, replay, err)
		}
	}
	short, err := first.Execute(derivedEnvelope(root, RuleDependsOn, LayerDeterministic, 1, 10))
	if err != nil || len(short.Result.DerivedDeterministic) != 1 || short.Result.DerivedDeterministic[0].Depth != 1 {
		t.Fatalf("max depth was not enforced: %#v %v", short.Result, err)
	}
}
