package query

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestEnvelopeReplayAndPermutationAreCanonical(t *testing.T) {
	facts := []Fact{
		NewFact(id("urn:gooo:activity:pay"), Used, id("urn:gooo:entity:order")),
		NewCandidateFact(id("urn:gooo:entity:order"), WasDerivedFrom, id("urn:gooo:entity:archive"), "ambiguous"),
		NewFact(id("urn:gooo:entity:payment"), WasGeneratedBy, id("urn:gooo:activity:pay")),
	}
	first, second := New(), New()
	for _, fact := range facts {
		assertAdd(t, first, fact)
	}
	for index := len(facts) - 1; index >= 0; index-- {
		assertAdd(t, second, facts[index])
	}
	request := traversalEnvelope(id("urn:gooo:activity:pay"), LayerAll, 2, 10)
	request.Relation = PROVUsed
	beforeCanonical, beforeNodes := first.Canonical(), first.Nodes()
	want, err := first.Execute(request)
	if err != nil {
		t.Fatal(err)
	}
	for run := 0; run < 3; run++ {
		got, err := second.Execute(request)
		if err != nil {
			t.Fatal(err)
		}
		assertSameEnvelope(t, got, want)
	}
	canonical, err := want.CanonicalJSON()
	if err != nil || !json.Valid(canonical) {
		t.Fatalf("invalid canonical response JSON: %s %v", canonical, err)
	}
	if want.Hash == "" || want.Hash != want.CanonicalDigestValue() {
		t.Fatalf("response hash is not a receipt: %q", want.Hash)
	}
	canonicalRequest, err := request.CanonicalDigest()
	if err != nil || canonicalRequest != want.RequestHash {
		t.Fatalf("request hash is not canonical: %q/%q", canonicalRequest, want.RequestHash)
	}
	if first.Canonical() != beforeCanonical || !reflect.DeepEqual(first.Nodes(), beforeNodes) {
		t.Fatal("successful envelope query mutated the graph")
	}
	if want.Metadata.ProvenanceStatus != StatusDeferred {
		t.Fatalf("missing provenance was not deferred: %#v", want.Metadata)
	}
	if label := authorityLabel(want.Metadata, "provenance"); label.Status != StatusDeferred {
		t.Fatalf("provenance authority label = %#v", label)
	}
	if label := authorityLabel(want.Metadata, "query_graph"); label.Authority != "derived" {
		t.Fatalf("query graph authority label = %#v", label)
	}
	var wire map[string]any
	if err := json.Unmarshal(mustMarshal(t, want), &wire); err != nil {
		t.Fatal(err)
	}
	if wire["schema"] != QueryEnvelopeSchema || wire["canonical_hash"] != want.Hash {
		t.Fatalf("wire envelope identity = %#v", wire)
	}
}

func TestEnvelopeFiltersLayersAndBoundsResults(t *testing.T) {
	graph := New()
	root, order := id("urn:query:activity:root"), id("urn:query:entity:order")
	invoice := id("urn:query:entity:invoice")
	assertAdd(t, graph, NewFact(root, Used, order))
	assertAdd(t, graph, NewCandidateFact(root, Used, invoice, "unresolved"))

	all := traversalEnvelope(root, LayerAll, 1, 1)
	response, err := graph.Execute(all)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DeterministicPaths) != 1 || len(response.Result.CandidatePaths) != 0 {
		t.Fatalf("limit did not prefer deterministic paths: %#v", response.Result)
	}

	candidate := traversalEnvelope(root, LayerCandidate, 1, 10)
	response, err = graph.Execute(candidate)
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DeterministicPaths) != 0 || len(response.Result.CandidatePaths) != 1 {
		t.Fatalf("candidate layer leaked another layer: %#v", response.Result)
	}

	exact := exactEnvelope(root, invoice, LayerCandidate)
	response, err = graph.Execute(exact)
	if err != nil || len(response.Result.CandidateMatches) != 1 || len(response.Result.DeterministicMatches) != 0 {
		t.Fatalf("candidate exact result = %#v, err=%v", response.Result, err)
	}
}

func TestEnvelopeRejectsInvalidRequestsWithoutMutation(t *testing.T) {
	graph := New()
	root, target := id("urn:query:activity:root"), id("urn:query:entity:target")
	assertAdd(t, graph, NewFact(root, Used, target))
	base := exactEnvelope(root, target, LayerDeterministic)
	cases := []struct {
		name string
		edit func(*Request)
		code string
	}{
		{"schema", func(request *Request) { request.Schema = "gooo-query/v0" }, "invalid_schema"},
		{"operation", func(request *Request) { request.Operation = "scan" }, "unsupported_operation"},
		{"layer", func(request *Request) { request.Layer = "probable" }, "unsupported_layer"},
		{"relation", func(request *Request) { request.Relation = "prov:unknown" }, "unsupported_relation"},
		{"limit-zero", func(request *Request) { request.Limit = 0 }, "invalid_limit"},
		{"limit-overflow", func(request *Request) { request.Limit = MaxEnvelopeLimit + 1 }, "invalid_limit"},
		{"depth-overflow", func(request *Request) { request.MaxDepth = MaxEnvelopeDepth + 1 }, "invalid_max_depth"},
		{"unknown-root", func(request *Request) { request.Root = id("urn:query:activity:missing") }, "unknown_endpoint"},
		{"unknown-target", func(request *Request) { request.Target = id("urn:query:entity:missing") }, "unknown_endpoint"},
		{"exact-direction", func(request *Request) { request.Direction = "incoming" }, "ambiguous_direction"},
	}
	traversalCases := []struct {
		name    string
		request Request
		code    string
	}{
		{"ambiguous-traversal", traversalWithoutDirection(root), "ambiguous_direction"},
		{"unsupported-direction", withDirection(traversalWithoutDirection(root), "diagonal"), "unsupported_direction"},
		{"unexpected-target", withTarget(traversalWithoutDirection(root), target), "unexpected_target"},
	}
	for _, testCase := range cases {
		request := base
		testCase.edit(&request)
		assertRejectedEnvelope(t, graph, request, testCase.name, testCase.code)
	}
	for _, testCase := range traversalCases {
		assertRejectedEnvelope(t, graph, testCase.request, testCase.name, testCase.code)
	}
}

func TestEnvelopeProjectionMarksMissingProvenanceDeferred(t *testing.T) {
	ir := semantic.NewIR("billing", "billing")
	activity, err := semantic.NewActivity("billing://activity/pay", "billing", "PayOrder")
	if err != nil {
		t.Fatal(err)
	}
	if err := ir.AddNode(activity); err != nil {
		t.Fatal(err)
	}
	graph, err := FromSemanticIR(ir)
	if err != nil {
		t.Fatal(err)
	}
	response, err := graph.Execute(traversalEnvelope(ID(activity.ID.String()), LayerAll, 1, 1))
	if err != nil {
		t.Fatal(err)
	}
	if response.Metadata.IRStatus != "available" || response.Metadata.SemanticDigest != ir.StableHash() {
		t.Fatalf("IR projection metadata = %#v", response.Metadata)
	}
	if response.Metadata.ProvenanceStatus != StatusDeferred {
		t.Fatalf("missing external provenance was fabricated: %#v", response.Metadata)
	}
	if response.Metadata.GraphHash == response.Metadata.SemanticDigest {
		t.Fatal("graph hash was treated as semantic source of truth")
	}
}

func TestEnvelopeCycleRemainsSimpleAndBounded(t *testing.T) {
	graph := New()
	a, b, c := id("urn:query:a"), id("urn:query:b"), id("urn:query:c")
	assertAdd(t, graph, NewFact(a, WasDerivedFrom, b))
	assertAdd(t, graph, NewFact(b, WasDerivedFrom, c))
	assertAdd(t, graph, NewFact(c, WasDerivedFrom, a))
	response, err := graph.Execute(traversalEnvelope(a, LayerDeterministic, 8, 100))
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Result.DeterministicPaths) != 2 {
		t.Fatalf("cycle was not bounded as simple paths: %#v", response.Result)
	}
	for _, path := range response.Result.DeterministicPaths {
		if path.Depth() > 8 || hasRepeatedID(path.IDs) {
			t.Fatalf("cycle path is invalid: %#v", path)
		}
	}
}

func assertSameEnvelope(t *testing.T, got, want Response) {
	t.Helper()
	gotJSON, gotErr := got.CanonicalJSON()
	wantJSON, wantErr := want.CanonicalJSON()
	if gotErr != nil || wantErr != nil || !reflect.DeepEqual(gotJSON, wantJSON) || got.Hash != want.Hash {
		t.Fatalf("envelope replay changed: got=%s/%q want=%s/%q", gotJSON, got.Hash, wantJSON, want.Hash)
	}
}

func assertRejectedEnvelope(t *testing.T, graph *Graph, request Request, name, wantCode string) {
	t.Helper()
	beforeCanonical, beforeNodes := graph.Canonical(), graph.Nodes()
	response, err := graph.Execute(request)
	if err == nil || response.Status != ResponseError || response.Error == nil {
		t.Fatalf("%s was not rejected: response=%#v err=%v", name, response, err)
	}
	if response.Error.Code != wantCode {
		t.Fatalf("%s error code = %q, want %q (%v)", name, response.Error.Code, wantCode, err)
	}
	if response.Hash == "" || response.Hash != response.CanonicalDigestValue() {
		t.Fatalf("%s rejection was not sealed: %#v", name, response)
	}
	if graph.Canonical() != beforeCanonical || !reflect.DeepEqual(graph.Nodes(), beforeNodes) {
		t.Fatalf("%s mutated the query graph", name)
	}
}

func traversalEnvelope(root ID, layer Layer, depth, limit int) Request {
	return Request{
		Schema: QueryEnvelopeSchema, Operation: OperationTraversal, Root: root,
		Layer: layer, Direction: "outgoing", MaxDepth: depth, Limit: limit,
	}
}

func traversalWithoutDirection(root ID) Request {
	request := traversalEnvelope(root, LayerAll, 1, 10)
	request.Direction = ""
	return request
}

func exactEnvelope(root, target ID, layer Layer) Request {
	return Request{
		Schema: QueryEnvelopeSchema, Operation: OperationExact, Root: root,
		Target: target, Relation: Used, Layer: layer, MaxDepth: 1, Limit: 10,
		Direction: "outgoing",
	}
}

func withDirection(request Request, direction string) Request {
	request.Direction = direction
	return request
}

func withTarget(request Request, target ID) Request {
	request.Target = target
	return request
}

func authorityLabel(metadata EnvelopeMetadata, view string) AuthorityLabel {
	for _, label := range metadata.AuthorityLabels {
		if label.View == view {
			return label
		}
	}
	return AuthorityLabel{}
}

func mustMarshal(t *testing.T, value any) []byte {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func (response Response) CanonicalDigestValue() string {
	digest, err := response.CanonicalDigest()
	if err != nil {
		return ""
	}
	return digest
}
