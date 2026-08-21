package semantic

import (
	"testing"
)

func TestApplyGraphPatchRejectsMutationWithoutChangingOriginal(t *testing.T) {
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
	_, err = graph.ApplyGraphPatch(GraphPatchBase{SourceDigest: sourceDigest, IRDigest: irDigest}, request, GraphPatchMutation{Name: "Process"})
	if err == nil {
		t.Fatal("name-collision mutation was accepted")
	}
	if graph.Canonical() != beforeCanonical || graph.StableHash() != beforeHash {
		t.Fatal("rejected mutation changed original graph")
	}
}
func TestApplyGraphPatchAddsOnlyValidatedDeterministicFact(t *testing.T) {
	graph, entity := patchFixture(t)
	other := mustEntity(t, MustIdentity("urn:gooo:source"), Namespace("billing"), "Source")
	if err := graph.AddNode(other); err != nil {
		t.Fatal(err)
	}
	sourceDigest := StableHashString("source")
	irDigest := StableHashString("ir")
	request := GraphPatchRequest{
		SchemaVersion: GraphPatchSchemaVersion, Operation: GraphPatchAddFact,
		ExpectedGraphHash: graph.StableHash(), Subject: entity.ID, Predicate: WasDerivedFrom,
		Object: other.ID, ExpectedSourceDigest: sourceDigest, ExpectedIRDigest: irDigest,
		AllowedIntent: "add derivation", Locality: "fact:" + entity.ID.String(),
	}
	fact := NewWasDerivedFromFact(entity.ID, other.ID)
	before := graph.StableHash()
	updated, err := graph.ApplyGraphPatch(GraphPatchBase{SourceDigest: sourceDigest, IRDigest: irDigest}, request, GraphPatchMutation{Fact: &fact})
	if err != nil {
		t.Fatalf("fact patch rejected: %v", err)
	}
	if !updated.HasFact(fact.Key()) || graph.HasFact(fact.Key()) || graph.StableHash() != before {
		t.Fatal("fact patch did not preserve copy-on-write semantics")
	}
}
