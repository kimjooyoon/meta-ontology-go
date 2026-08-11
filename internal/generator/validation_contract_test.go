package generator

import (
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

func TestGenerateRejectsInputOutputNameCollision(t *testing.T) {
	ir := billingIR()
	ir.Activities[0].Outputs[0].GoName = ir.Activities[0].Inputs[0].GoName
	assertGenerationError(t, ir, "conflicts with input name")
}

func assertGenerationError(t *testing.T, ir SemanticIR, message string) {
	t.Helper()
	if _, err := Generate(ir, nil); err == nil || !strings.Contains(err.Error(), message) {
		t.Fatalf("expected generation error containing %q, got %v", message, err)
	}
}
