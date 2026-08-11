package generator

import (
	"reflect"
	"strings"
	"testing"
)

func TestGenerateRejectsDuplicateImportPath(t *testing.T) {
	ir := billingIR()
	ir.Imports = []Import{{Name: "first", Path: "example/shared"}, {Name: "second", Path: "example/shared"}}
	assertGenerationError(t, ir, "duplicate import path")
}

func TestGenerateRejectsTopLevelNameCollision(t *testing.T) {
	ir := billingIR()
	ir.Entities[0].GoName = ir.Activities[0].GoName
	assertGenerationError(t, ir, "Go name")
}

func TestGenerateRejectsCrossKindStableIDCollision(t *testing.T) {
	cases := []struct {
		name string
		edit func(*SemanticIR)
	}{
		{name: "entity-activity", edit: func(ir *SemanticIR) { ir.Activities[0].ID = ir.Entities[0].ID }},
		{name: "entity-slot", edit: func(ir *SemanticIR) { ir.Activities[0].Slots[0].ID = ir.Entities[0].ID }},
		{name: "activity-slot", edit: func(ir *SemanticIR) { ir.Activities[0].Slots[0].ID = ir.Activities[0].ID }},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			ir := acceptanceFixture()
			testCase.edit(&ir)
			assertGenerationError(t, ir, "stable ID")
		})
	}
}

func TestGenerateRejectsUnknownPortType(t *testing.T) {
	// GEN-COMPILE-001: go/types rejects a parseable but uncompilable port type.
	ir := billingIR()
	ir.Activities[0].Inputs[0].GoType = "MissingOrder"
	assertGenerationError(t, ir, "invalid Go type")
}

func TestGenerateFailureDoesNotMutateIR(t *testing.T) {
	ir := acceptanceFixture()
	ir.Activities[0].Slots[0].ID = ir.Entities[0].ID
	before := copyIR(ir)
	if _, err := Generate(ir, nil); err == nil {
		t.Fatal("stable-ID collision was accepted")
	}
	if !reflect.DeepEqual(ir, before) {
		t.Fatal("rejected generation mutated caller-owned IR")
	}
}

func assertGenerationError(t *testing.T, ir SemanticIR, message string) {
	t.Helper()
	if _, err := Generate(ir, nil); err == nil || !strings.Contains(err.Error(), message) {
		t.Fatalf("expected generation error containing %q, got %v", message, err)
	}
}
