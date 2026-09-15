package query

import (
	"testing"
)

func TestEnvelopeTraversalInvalidRelationFailsClosedWithoutMutation(t *testing.T) {
	graph := New()
	root := id("urn:query:invalid-traversal:root")
	target := id("urn:query:invalid-traversal:target")
	assertAdd(t, graph, NewFact(root, WasDerivedFrom, target))
	request := traversalEnvelope(root, LayerDeterministic, 2, 2)
	request.Relation = Relation("gooo:unknown")
	beforeHash := graph.StableHash()
	response, err := graph.Execute(request)
	if err == nil || response.Error == nil || response.Error.Code != "unsupported_relation" {
		t.Fatalf("invalid traversal relation was not rejected: %#v, err=%v", response, err)
	}
	if response.Metadata.GraphHash != beforeHash || graph.StableHash() != beforeHash {
		t.Fatalf("invalid traversal relation changed graph hash: response=%q graph=%q want=%q",
			response.Metadata.GraphHash, graph.StableHash(), beforeHash)
	}
	if response.Hash == "" || response.Hash != response.CanonicalDigestValue() {
		t.Fatalf("invalid traversal rejection was not sealed: %#v", response)
	}
}
