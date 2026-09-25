package generator

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestReflectiveFallbackRejectsUnorderedRelationMap(t *testing.T) {
	input := reflectiveGraph()
	input.Facts = map[string]reflectiveFactFixture{
		"used": {Subject: "activity:run", Predicate: "used", Object: "entity:z"},
	}
	_, _, err := GenerateFrom(input, Options{})
	if !errors.Is(err, ErrDeferredRelationOrder) {
		t.Fatalf("expected deferred relation order, got %v", err)
	}
}
func TestReflectiveFallbackRequiresFactsCollection(t *testing.T) {
	input := missingFactsGraphFixture{Package: "reflectgen", Nodes: baseNodes()}
	_, _, err := GenerateFrom(input, Options{})
	if err == nil || !strings.Contains(err.Error(), "does not expose Facts") {
		t.Fatalf("expected missing Facts rejection, got %v", err)
	}
}
func TestReflectiveFallbackDoesNotMutateInput(t *testing.T) {
	input := reflectiveGraph()
	input.Facts = []reflectiveFactFixture{{Subject: "activity:run", Predicate: "used", Object: "entity:missing"}}
	before := input
	_, _, err := GenerateFrom(&input, Options{})
	if err == nil {
		t.Fatal("missing endpoint was accepted")
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatal("rejected reflective input was mutated")
	}
}
func TestTypedAdapterPreservesAuthoritativePortOrder(t *testing.T) {
	provider := semanticIRProviderFixture{ir: SemanticIR{
		Package: "orderedgen",
		Entities: []Entity{
			{ID: "entity:z", Name: "Zeta", GoName: "Zeta"},
			{ID: "entity:a", Name: "Alpha", GoName: "Alpha"},
		},
		Activities: []Activity{{
			ID: "activity:run", Name: "Run", GoName: "Run",
			Inputs: []Port{
				{EntityID: "entity:z", Name: "zeta", GoName: "zeta"},
				{EntityID: "entity:a", Name: "alpha", GoName: "alpha"},
			},
		}},
	}}
	source, _, err := GenerateFrom(provider, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), "func Run(zeta Zeta, alpha Alpha)") {
		t.Fatalf("typed adapter port order was not preserved:\n%s", source)
	}
}
