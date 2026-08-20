package semantic

import (
	"errors"
	"testing"
)

func TestValidatePatchPreconditionsRejectsInvalidFieldAndBase(t *testing.T) {
	graph, node := patchFixture(t)
	sourceDigest := StableHashString("source")
	irDigest := StableHashString("ir")
	fieldDigest, err := NodeFieldHash(node, "name")
	if err != nil {
		t.Fatal(err)
	}
	request := GraphPatchRequest{
		SchemaVersion: GraphPatchSchemaVersion, Operation: GraphPatchSetNodeField,
		ExpectedGraphHash: graph.StableHash(), NodeID: node.ID, ExpectedNodeHash: node.StableHash(),
		Field: "kind", ExpectedFieldHash: fieldDigest, ExpectedSourceDigest: sourceDigest,
		ExpectedIRDigest: irDigest, AllowedIntent: "rename node", Locality: "node:" + node.ID.String(),
	}
	err = graph.ValidatePatchPreconditions(GraphPatchBase{SourceDigest: sourceDigest, IRDigest: irDigest}, request)
	if err == nil {
		t.Fatal("immutable field patch was accepted")
	}
	var conflict GraphPatchConflict
	if !errors.As(err, &conflict) || conflict.Code != PatchImmutableField {
		t.Fatalf("immutable field error = %v, want %s", err, PatchImmutableField)
	}

	request.Field = "name"
	request.ExpectedFieldHash, _ = NodeFieldHash(node, "name")
	err = graph.ValidatePatchPreconditions(GraphPatchBase{SourceDigest: sourceDigest, IRDigest: StableHashString("other")}, request)
	if err == nil {
		t.Fatal("mismatched base tuple was accepted")
	}
	if !errors.As(err, &conflict) || conflict.Code != PatchBaseTupleMismatch {
		t.Fatalf("base tuple error = %v, want %s", err, PatchBaseTupleMismatch)
	}
}
func TestApplyGraphPatchReturnsCopyAndPreservesOriginal(t *testing.T) {
	graph, node := patchFixture(t)
	sourceDigest := StableHashString("source")
	irDigest := StableHashString("ir")
	fieldDigest, err := NodeFieldHash(node, "name")
	if err != nil {
		t.Fatal(err)
	}
	request := GraphPatchRequest{
		SchemaVersion: GraphPatchSchemaVersion, Operation: GraphPatchSetNodeField,
		ExpectedGraphHash: graph.StableHash(), NodeID: node.ID, ExpectedNodeHash: node.StableHash(),
		Field: "name", ExpectedFieldHash: fieldDigest, ExpectedSourceDigest: sourceDigest,
		ExpectedIRDigest: irDigest, AllowedIntent: "rename node", Locality: "node:" + node.ID.String(),
	}
	beforeCanonical, beforeHash := graph.Canonical(), graph.StableHash()
	updated, err := graph.ApplyGraphPatch(GraphPatchBase{SourceDigest: sourceDigest, IRDigest: irDigest}, request, GraphPatchMutation{Name: "Purchase"})
	if err != nil {
		t.Fatalf("rename patch rejected: %v", err)
	}
	if graph.Canonical() != beforeCanonical || graph.StableHash() != beforeHash {
		t.Fatal("applying patch mutated original graph")
	}
	updatedNode, ok := updated.Node(node.ID)
	if !ok || updatedNode.Name != "Purchase" {
		t.Fatalf("updated node = %#v, want renamed node", updatedNode)
	}
	if updated.StableHash() != beforeHash || updated.Canonical() == beforeCanonical {
		t.Fatal("presentation-only rename changed semantic hash or failed to change canonical view")
	}
}
