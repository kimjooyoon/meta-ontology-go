package query

import (
	"errors"
	"reflect"
	"testing"
)

func TestDerivedInverseRulesSeparateLayersAndStayOutOfGraphHash(t *testing.T) {
	graph := New()
	activity := id("urn:derived:activity:compile")
	entity := id("urn:derived:entity:source")
	other := id("urn:derived:activity:other")
	generated := id("urn:derived:entity:generated")
	nonIncident := id("urn:derived:activity:non-incident")
	derived := id("urn:derived:entity:derived")
	assertAdd(t, graph, NewFact(activity, Used, entity))
	assertAdd(t, graph, NewCandidateFact(other, Used, entity, "unresolved"))
	assertAdd(t, graph, NewFact(generated, WasGeneratedBy, activity))
	assertAdd(t, graph, NewFact(nonIncident, WasGeneratedBy, id("urn:derived:entity:elsewhere")))
	assertAdd(t, graph, NewFact(derived, WasDerivedFrom, entity))
	assertAdd(t, graph, NewFact(id("urn:derived:entity:unrelated"), WasDerivedFrom, id("urn:derived:entity:elsewhere")))
	beforeCanonical, beforeHash := graph.Canonical(), graph.StableHash()
	response, err := graph.Execute(derivedEnvelope(entity, RuleUsedBy, LayerAll, 1, 10))
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DerivedDeterministic) != 1 || len(response.Result.DerivedCandidates) != 1 {
		t.Fatalf("inverse layers were not separated: %#v", response.Result)
	}
	deterministic := response.Result.DerivedDeterministic[0]
	if deterministic.Subject != entity || deterministic.Predicate != DerivedUsedBy || deterministic.Object != activity {
		t.Fatalf("unexpected usedBy result: %#v", deterministic)
	}
	if deterministic.Status != DerivedFactStatus || deterministic.SourceLayer != FactDeterministic.String() {
		t.Fatalf("derived authority status was not explicit: %#v", deterministic)
	}
	if response.Metadata.DerivedStatus != DerivedStatusNonAuthoritative ||
		response.Metadata.DerivedRuleSchema != DerivedRuleSchemaVersion {
		t.Fatalf("derived metadata = %#v", response.Metadata)
	}
	if response.Metadata.GraphHash != beforeHash || graph.Canonical() != beforeCanonical ||
		graph.StableHash() != beforeHash {
		t.Fatal("derived evaluation mutated graph authority or hash")
	}
	label := envelopeAuthorityLabel(response.Metadata, "derived_query")
	if label.Authority != "derived" || label.Status != DerivedStatusNonAuthoritative {
		t.Fatalf("derived authority label = %#v", label)
	}

	generatedResponse, err := graph.Execute(derivedEnvelope(activity, RuleGeneratedBy, LayerDeterministic, 1, 10))
	if err != nil {
		t.Fatal(err)
	}
	if len(generatedResponse.Result.DerivedDeterministic) != 1 ||
		generatedResponse.Result.DerivedDeterministic[0].Object != generated {
		t.Fatalf("generatedBy did not filter incident edges: %#v", generatedResponse.Result)
	}
	derivedResponse, err := graph.Execute(derivedEnvelope(entity, RuleDerivedTo, LayerDeterministic, 1, 10))
	if err != nil {
		t.Fatal(err)
	}
	if len(derivedResponse.Result.DerivedDeterministic) != 1 ||
		derivedResponse.Result.DerivedDeterministic[0].Object != derived {
		t.Fatalf("derivedTo did not filter incident edges: %#v", derivedResponse.Result)
	}
}

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

func TestDerivedCandidateClosureAndLimit(t *testing.T) {
	root := id("urn:derived:candidate:root")
	first := id("urn:derived:candidate:first")
	second := id("urn:derived:candidate:second")
	third := id("urn:derived:candidate:third")
	graph := New()
	assertAdd(t, graph, NewFact(root, WasDerivedFrom, first))
	assertAdd(t, graph, NewCandidateFact(root, WasDerivedFrom, second, "ambiguous"))
	assertAdd(t, graph, NewFact(root, WasDerivedFrom, third))
	response, err := graph.Execute(derivedEnvelope(root, RuleDependsOn, LayerAll, 1, 1))
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DerivedDeterministic) != 1 || len(response.Result.DerivedCandidates) != 0 {
		t.Fatalf("limit did not prefer deterministic derived rows: %#v", response.Result)
	}
	candidates, err := graph.Execute(derivedEnvelope(root, RuleDependsOn, LayerCandidate, 1, 10))
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates.Result.DerivedDeterministic) != 0 || len(candidates.Result.DerivedCandidates) != 1 {
		t.Fatalf("candidate layer leaked: %#v", candidates.Result)
	}
	if candidates.Result.DerivedCandidates[0].SourceLayer != FactCandidate.String() {
		t.Fatalf("candidate source layer was promoted: %#v", candidates.Result.DerivedCandidates[0])
	}
	direct, err := graph.Derive(root, DerivedOptions{
		Rule: RuleDependsOn, MaxDepth: 1, Limit: 1, Selection: SelectAll,
	})
	if err != nil || len(direct.Deterministic) != 1 || len(direct.Candidates) != 0 {
		t.Fatalf("direct derived API did not enforce row limit: %#v %v", direct, err)
	}
}

func TestDerivedBoundsRejectInvalidDirectOptions(t *testing.T) {
	graph := New()
	root, target := id("urn:derived:bounds:root"), id("urn:derived:bounds:target")
	assertAdd(t, graph, NewFact(root, WasDerivedFrom, target))
	cases := []DerivedOptions{
		{Rule: RuleDependsOn, MaxDepth: 0, Limit: 1, Selection: SelectAll},
		{Rule: RuleDependsOn, MaxDepth: MaxEnvelopeDepth + 1, Limit: 1, Selection: SelectAll},
		{Rule: RuleDependsOn, MaxDepth: 1, Limit: 0, Selection: SelectAll},
		{Rule: RuleDependsOn, MaxDepth: 1, Limit: MaxEnvelopeLimit + 1, Selection: SelectAll},
	}
	for _, options := range cases {
		if _, err := graph.Derive(root, options); !errors.Is(err, ErrInvalidDerivedQuery) {
			t.Fatalf("invalid direct options returned %v", err)
		}
	}
}

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

func derivedEnvelope(root ID, rule DerivedRuleID, layer Layer, depth, limit int) Request {
	return Request{
		Schema: QueryEnvelopeSchema, Operation: OperationDerived, Root: root,
		Rule: rule, Layer: layer, Direction: "outgoing", MaxDepth: depth, Limit: limit,
	}
}

func envelopeAuthorityLabel(metadata EnvelopeMetadata, view string) AuthorityLabel {
	for _, label := range metadata.AuthorityLabels {
		if label.View == view {
			return label
		}
	}
	return AuthorityLabel{}
}
