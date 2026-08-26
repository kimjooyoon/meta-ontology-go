package query

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

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
