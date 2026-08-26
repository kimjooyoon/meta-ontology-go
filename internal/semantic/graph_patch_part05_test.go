package semantic

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGraphPatchBindingCanonicalIsStableAndComplete(t *testing.T) {
	base := GraphPatchBase{SourceDigest: StableHashString("source"), IRDigest: StableHashString("ir")}
	request := GraphPatchRequest{
		SchemaVersion: GraphPatchSchemaVersion, Operation: GraphPatchSetNodeField,
		ExpectedGraphHash: StableHashString("graph"), NodeID: MustIdentity("urn:gooo:entity"),
		ExpectedNodeHash: StableHashString("node"), Field: "name", ExpectedFieldHash: StableHashString("field"),
		ExpectedSourceDigest: base.SourceDigest, ExpectedIRDigest: base.IRDigest,
		AllowedIntent: "rename node", Locality: "node:urn:gooo:entity",
	}
	canonical := CanonicalGraphPatchBinding(base, request)
	if !json.Valid([]byte(canonical)) || canonical != CanonicalGraphPatchBinding(base, request) {
		t.Fatalf("binding is not stable JSON: %s", canonical)
	}
	for _, field := range []string{`"base"`, `"request"`, `"schema_version"`, `"expected_graph_hash"`, `"allowed_intent"`, `"locality"`} {
		if !strings.Contains(canonical, field) {
			t.Fatalf("canonical binding omits %s: %s", field, canonical)
		}
	}
	if GraphPatchBindingHash(base, request) != StableHashString(canonical) || request.StableHash() != StableHashString(request.Canonical()) {
		t.Fatal("canonical binding hash is not a SHA-256 of canonical bytes")
	}
	request.Locality = "node:urn:gooo:other"
	if GraphPatchBindingHash(base, request) == StableHashString(canonical) {
		t.Fatal("locality mutation did not change binding hash")
	}
}
func patchFixture(t *testing.T) (Graph, Node) {
	t.Helper()
	graph := NewGraph()
	node := mustEntity(t, MustIdentity("urn:gooo:entity"), Namespace("billing"), "Order")
	activity := mustActivity(t, MustIdentity("urn:gooo:activity"), Namespace("billing"), "Process")
	if err := graph.AddNode(node); err != nil {
		t.Fatal(err)
	}
	if err := graph.AddNode(activity); err != nil {
		t.Fatal(err)
	}
	if err := graph.AddFact(NewUsedFact(activity.ID, node.ID)); err != nil {
		t.Fatal(err)
	}
	return graph, node
}
