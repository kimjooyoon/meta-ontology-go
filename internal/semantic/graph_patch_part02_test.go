package semantic

import (
	"errors"
	"testing"
)

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
