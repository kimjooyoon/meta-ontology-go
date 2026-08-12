package generator

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type reflectiveNodeFixture struct {
	ID   string
	Kind string
	Name string
}

type reflectiveFactFixture struct {
	Subject   string
	Predicate string
	Object    string
}

type reflectiveGraphFixture struct {
	Package string
	Nodes   any
	Facts   any
}

type missingNameNodeFixture struct {
	ID   string
	Kind string
}

type missingFactsGraphFixture struct {
	Package string
	Nodes   any
}

type semanticIRProviderFixture struct {
	ir SemanticIR
}

func (provider semanticIRProviderFixture) SemanticIR() SemanticIR {
	return provider.ir
}

func TestReflectiveFallbackRejectsMalformedInputsDeterministically(t *testing.T) {
	base := reflectiveGraph()
	cases := []struct {
		name  string
		input any
		want  string
	}{
		{name: "unknown node kind", input: withNode(base, reflectiveNodeFixture{ID: "agent:a", Kind: "agent", Name: "Agent"}), want: "unsupported semantic node kind"},
		{name: "malformed node", input: withNodes(base, []any{"not a node"}), want: "semantic node 0 must be a struct"},
		{name: "missing node field", input: withNodes(base, []any{missingNameNodeFixture{ID: "entity:missing", Kind: "entity"}}), want: "missing Name"},
		{name: "duplicate node ID", input: withNodes(base, map[string]reflectiveNodeFixture{
			"first":  {ID: "entity:duplicate", Kind: "entity", Name: "First"},
			"second": {ID: "entity:duplicate", Kind: "entity", Name: "Second"},
		}), want: "duplicate semantic node ID"},
		{name: "malformed fact", input: withFacts(base, []any{"not a fact"}), want: "semantic fact 0 must be a struct"},
		{name: "missing endpoint", input: withFacts(base, []reflectiveFactFixture{{Subject: "activity:run", Predicate: "used", Object: "entity:missing"}}), want: "missing endpoint"},
		{name: "unsupported relation", input: withFacts(base, []reflectiveFactFixture{{Subject: "activity:run", Predicate: "wasDerivedFrom", Object: "entity:z"}}), want: "unsupported reflected relation"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var first string
			for attempt := 0; attempt < 20; attempt++ {
				_, _, err := GenerateFrom(testCase.input, Options{})
				if err == nil || !strings.Contains(err.Error(), testCase.want) {
					t.Fatalf("expected error containing %q, got %v", testCase.want, err)
				}
				if attempt == 0 {
					first = err.Error()
					continue
				}
				if err.Error() != first {
					t.Fatalf("nondeterministic error: first=%q current=%q", first, err)
				}
			}
		})
	}
}

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
			{ID: "entity:z", Name: "Zeta", GoName: "Zeta", Fields: []Field{{Name: "ID", GoName: "ID", GoType: "string"}}},
			{ID: "entity:a", Name: "Alpha", GoName: "Alpha", Fields: []Field{{Name: "ID", GoName: "ID", GoType: "string"}}},
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

func reflectiveGraph() reflectiveGraphFixture {
	return reflectiveGraphFixture{Package: "reflectgen", Nodes: baseNodes(), Facts: []reflectiveFactFixture{}}
}

func baseNodes() map[string]reflectiveNodeFixture {
	return map[string]reflectiveNodeFixture{
		"activity": {ID: "activity:run", Kind: "activity", Name: "Run"},
		"alpha":    {ID: "entity:a", Kind: "entity", Name: "Alpha"},
		"zeta":     {ID: "entity:z", Kind: "entity", Name: "Zeta"},
	}
}

func withNode(input reflectiveGraphFixture, node reflectiveNodeFixture) reflectiveGraphFixture {
	input.Nodes = map[string]reflectiveNodeFixture{"base": node}
	return input
}

func withNodes(input reflectiveGraphFixture, nodes any) reflectiveGraphFixture {
	input.Nodes = nodes
	return input
}

func withFacts(input reflectiveGraphFixture, facts any) reflectiveGraphFixture {
	input.Facts = facts
	return input
}
