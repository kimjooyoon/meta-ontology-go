package semantic

import (
	"errors"
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
			if _, ok := errors.AsType[GraphPatchConflict](err); !ok {
				t.Fatalf("error is not GraphPatchConflict: %v", err)
			}
			if errors.Is(err, ErrGraphPatch) == false {
				t.Fatalf("error does not unwrap to ErrGraphPatch: %v", err)
			}
		})
	}
}
