package query

import (
	"reflect"
	"testing"
)

func TestDerivedCycleUsesBoundedNodeStates(t *testing.T) {
	graph := New()
	root := id("urn:derived:state:root")
	a := id("urn:derived:state:a")
	b := id("urn:derived:state:b")
	c := id("urn:derived:state:c")
	d := id("urn:derived:state:d")
	assertAdd(t, graph, NewFact(root, WasDerivedFrom, a))
	assertAdd(t, graph, NewFact(a, WasDerivedFrom, b))
	assertAdd(t, graph, NewFact(b, WasDerivedFrom, a))
	assertAdd(t, graph, NewFact(b, WasDerivedFrom, c))
	assertAdd(t, graph, NewFact(c, WasDerivedFrom, b))
	assertAdd(t, graph, NewFact(c, WasDerivedFrom, d))
	beforeNodes := graph.Nodes()
	result, err := graph.Derive(root, DerivedOptions{
		Rule: RuleDependsOn, MaxDepth: 4, Limit: MaxEnvelopeLimit, Selection: SelectDeterministic,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Deterministic) != 4 {
		t.Fatalf("cycle state visit emitted wrong closure: %#v", result.Deterministic)
	}
	for _, row := range result.Deterministic {
		if row.Object == root || row.Depth > 4 {
			t.Fatalf("cycle or depth escaped closure: %#v", row)
		}
	}
	if !reflect.DeepEqual(graph.Nodes(), beforeNodes) {
		t.Fatal("bounded closure mutated graph nodes")
	}
}
func TestDerivedEnvelopeRejectsUnknownOrReversedRulesWithoutMutation(t *testing.T) {
	graph := New()
	root, target := id("urn:derived:invalid:root"), id("urn:derived:invalid:target")
	assertAdd(t, graph, NewFact(root, WasDerivedFrom, target))
	base := derivedEnvelope(root, RuleDependsOn, LayerAll, 1, 10)
	cases := []struct {
		name string
		edit func(*Request)
		code string
	}{
		{"unknown-rule", func(request *Request) { request.Rule = "used" }, "unsupported_rule"},
		{"reversed-rule", func(request *Request) {
			request.Rule = DerivedRuleID(DerivedRuleSchemaVersion + "/inverse/used")
		}, "unsupported_rule"},
		{"relation-filter", func(request *Request) { request.Relation = Used }, "derived_relation_rejected"},
		{"reversed-direction", func(request *Request) { request.Direction = "incoming" }, "reversed_direction"},
		{"inverse-depth", func(request *Request) { request.Rule = RuleUsedBy; request.MaxDepth = 2 }, "invalid_rule_options"},
		{"unknown-root", func(request *Request) { request.Root = id("urn:derived:invalid:missing") }, "unknown_endpoint"},
	}
	for _, testCase := range cases {
		request := base
		testCase.edit(&request)
		beforeCanonical, beforeHash := graph.Canonical(), graph.StableHash()
		response, err := graph.Execute(request)
		if err == nil || response.Status != ResponseError || response.Error == nil {
			t.Fatalf("%s was not rejected: %#v %v", testCase.name, response, err)
		}
		if response.Error.Code != testCase.code {
			t.Fatalf("%s code = %q, want %q", testCase.name, response.Error.Code, testCase.code)
		}
		if graph.Canonical() != beforeCanonical || graph.StableHash() != beforeHash {
			t.Fatalf("%s mutated graph state", testCase.name)
		}
	}
}
