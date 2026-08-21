package generator

import (
	"bytes"
	"testing"
)

func TestProjectionMetadataV1RejectsCollectionTamperWithoutMutation(t *testing.T) {
	result, err := GenerateProjectionV1(acceptanceFixture(), nil)
	if err != nil {
		t.Fatal(err)
	}
	before, err := result.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	tampered := cloneProjectionV1(result)
	tampered.SemanticIR.Entities = append(tampered.SemanticIR.Entities, Entity{
		ID: "entity:tampered", Name: "Tampered", GoName: "Tampered",
	})
	if _, err := tampered.CanonicalJSON(); err == nil {
		t.Fatal("tampered semantic collection was accepted")
	}
	after, err := result.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("rejecting collection tamper mutated the original projection")
	}
}
