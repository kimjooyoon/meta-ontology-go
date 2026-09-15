package query

import (
	"reflect"
	"testing"
)

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
