package query

import (
	"reflect"
	"testing"
)

func TestDerivedInverseLimitPreservesCanonicalPrefix(t *testing.T) {
	root := id("urn:derived:inverse:limit:root")
	graph := New()
	for _, activity := range []string{"z", "a", "m"} {
		assertAdd(t, graph, NewFact(id("urn:derived:inverse:limit:"+activity), Used, root))
	}
	limited, err := graph.Execute(derivedEnvelope(root, RuleUsedBy, LayerDeterministic, 1, 1))
	if err != nil {
		t.Fatal(err)
	}
	full, err := graph.Execute(derivedEnvelope(root, RuleUsedBy, LayerDeterministic, 1, 10))
	if err != nil {
		t.Fatal(err)
	}
	if len(full.Result.DerivedDeterministic) != 3 ||
		!reflect.DeepEqual(limited.Result.DerivedDeterministic, full.Result.DerivedDeterministic[:1]) {
		t.Fatalf("inverse limit changed canonical prefix: %#v vs %#v", limited.Result, full.Result)
	}
}
