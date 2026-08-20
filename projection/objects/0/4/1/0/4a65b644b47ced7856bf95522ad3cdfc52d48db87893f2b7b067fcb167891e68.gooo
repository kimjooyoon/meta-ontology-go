package generator

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestProjectionMetadataV1CanonicalizesNilAndEmptyIRCollections(t *testing.T) {
	nilIR := SemanticIR{
		Package:  "canonicalgen",
		Entities: []Entity{{ID: "entity:item", Name: "Item", GoName: "Item"}},
		Activities: []Activity{{
			ID: "activity:run", Name: "Run", GoName: "Run",
			Slots: []Slot{{ID: "slot:run", Default: ""}},
		}},
	}
	emptyIR := SemanticIR{
		Package:  "canonicalgen",
		Imports:  []Import{},
		Entities: []Entity{{ID: "entity:item", Name: "Item", GoName: "Item", Fields: []Field{}}},
		Activities: []Activity{{
			ID: "activity:run", Name: "Run", GoName: "Run",
			Inputs: []Port{}, Outputs: []Port{}, Slots: []Slot{{ID: "slot:run", Default: ""}},
		}},
	}
	nilBefore, err := json.Marshal(nilIR)
	if err != nil {
		t.Fatal(err)
	}
	emptyBefore, err := json.Marshal(emptyIR)
	if err != nil {
		t.Fatal(err)
	}

	nilResult, err := GenerateProjectionV1(nilIR, nil)
	if err != nil {
		t.Fatal(err)
	}
	emptyResult, err := GenerateProjectionV1(emptyIR, nil)
	if err != nil {
		t.Fatal(err)
	}
	nilJSON, err := nilResult.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	emptyJSON, err := emptyResult.CanonicalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(nilResult.Source, emptyResult.Source) ||
		!reflect.DeepEqual(nilResult.Metadata, emptyResult.Metadata) ||
		!bytes.Equal(nilJSON, emptyJSON) {
		t.Fatal("nil and empty IR collections produced different projection metadata")
	}
	nilAfter, err := json.Marshal(nilIR)
	if err != nil {
		t.Fatal(err)
	}
	emptyAfter, err := json.Marshal(emptyIR)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(nilBefore, nilAfter) || !bytes.Equal(emptyBefore, emptyAfter) {
		t.Fatal("collection canonicalization mutated caller-owned IR")
	}
}
