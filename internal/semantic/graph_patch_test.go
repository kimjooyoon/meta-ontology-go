package semantic

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestValidatePatchPreconditionsAcceptsTypedNodeFieldEdit(t *testing.T) {
	graph, node := patchFixture(t)
	sourceDigest := StableHashString("source")
	irDigest := StableHashString("ir")
	fieldDigest, err := NodeFieldHash(node, "name")
	if err != nil {
		t.Fatal(err)
	}
	err = graph.ValidatePatchPreconditions(GraphPatchBase{SourceDigest: sourceDigest, IRDigest: irDigest}, GraphPatchRequest{
		SchemaVersion: GraphPatchSchemaVersion, Operation: GraphPatchSetNodeField,
		ExpectedGraphHash: graph.StableHash(), NodeID: node.ID, ExpectedNodeHash: node.StableHash(),
		Field: "name", ExpectedFieldHash: fieldDigest, ExpectedSourceDigest: sourceDigest,
		ExpectedIRDigest: irDigest, AllowedIntent: "rename node", Locality: "node:" + node.ID.String(),
	})
	if err != nil {
		t.Fatalf("valid patch rejected: %v", err)
	}
}

func TestValidatePatchPreconditionsRejectsStaleAndMismatchedFields(t *testing.T) {
	graph, node := patchFixture(t)
	sourceDigest := StableHashString("source")
	irDigest := StableHashString("ir")
	fieldDigest, err := NodeFieldHash(node, "name")
	if err != nil {
		t.Fatal(err)
	}
	base := GraphPatchBase{SourceDigest: sourceDigest, IRDigest: irDigest}
	request := GraphPatchRequest{
		SchemaVersion: GraphPatchSchemaVersion, Operation: GraphPatchSetNodeField,
		ExpectedGraphHash: graph.StableHash(), NodeID: node.ID, ExpectedNodeHash: node.StableHash(),
		Field: "name", ExpectedFieldHash: fieldDigest, ExpectedSourceDigest: sourceDigest,
		ExpectedIRDigest: irDigest, AllowedIntent: "rename node", Locality: "node:" + node.ID.String(),
	}
	for name, mutate := range map[string]func(*GraphPatchRequest){
		"stale graph": func(r *GraphPatchRequest) { r.ExpectedGraphHash = StableHashString("stale") },
		"node hash":   func(r *GraphPatchRequest) { r.ExpectedNodeHash = StableHashString("stale") },
		"field hash":  func(r *GraphPatchRequest) { r.ExpectedFieldHash = StableHashString("stale") },
		"base tuple":  func(r *GraphPatchRequest) { r.ExpectedIRDigest = StableHashString("other") },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := request
			mutate(&candidate)
			err := graph.ValidatePatchPreconditions(base, candidate)
			if err == nil {
				t.Fatal("invalid patch was accepted")
			}
			var conflict GraphPatchConflict
			if !errors.As(err, &conflict) {
				t.Fatalf("error is not GraphPatchConflict: %v", err)
			}
			if errors.Is(err, ErrGraphPatch) == false {
				t.Fatalf("error does not unwrap to ErrGraphPatch: %v", err)
			}
		})
	}
}

func TestValidatePatchPreconditionsRejectsTypedFactErrorsWithoutMutation(t *testing.T) {
	graph, _ := patchFixture(t)
	sourceDigest := StableHashString("source")
	irDigest := StableHashString("ir")
	base := GraphPatchBase{SourceDigest: sourceDigest, IRDigest: irDigest}
	beforeCanonical, beforeHash := graph.Canonical(), graph.StableHash()
	request := GraphPatchRequest{
		SchemaVersion: GraphPatchSchemaVersion, Operation: GraphPatchAddFact,
		ExpectedGraphHash: beforeHash, Subject: MustIdentity("urn:gooo:missing"),
		Predicate: WasDerivedFrom, Object: MustIdentity("urn:gooo:entity"),
		ExpectedSourceDigest: sourceDigest, ExpectedIRDigest: irDigest,
		AllowedIntent: "add relation", Locality: "fact:urn:gooo:missing",
	}
	err := graph.ValidatePatchPreconditions(base, request)
	if err == nil {
		t.Fatal("unknown endpoint patch was accepted")
	}
	var conflict GraphPatchConflict
	if !errors.As(err, &conflict) || conflict.Code != PatchUnknownEndpoint {
		t.Fatalf("unknown endpoint error = %v, want %s", err, PatchUnknownEndpoint)
	}
	if graph.Canonical() != beforeCanonical || graph.StableHash() != beforeHash {
		t.Fatal("rejected patch mutated graph")
	}

	request.Subject = MustIdentity("urn:gooo:entity")
	request.Object = MustIdentity("urn:gooo:activity")
	request.Predicate = Used
	err = graph.ValidatePatchPreconditions(base, request)
	if err == nil {
		t.Fatal("reversed relation patch was accepted")
	}
	if !errors.As(err, &conflict) || conflict.Code != PatchEndpointKindMismatch {
		t.Fatalf("kind mismatch error = %v, want %s", err, PatchEndpointKindMismatch)
	}
}

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
