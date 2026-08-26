package generator

import (
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
			for attempt := range 20 {
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
