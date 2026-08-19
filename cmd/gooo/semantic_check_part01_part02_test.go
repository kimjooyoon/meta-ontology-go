package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestSemanticDiagnosticClassifiesInvalidRelation(t *testing.T) {
	if code := semanticDiagnosticCode(semantic.ErrUnknownRelation); code != "semantic.invalid-relation" {
		t.Fatalf("invalid relation code = %q", code)
	}
}
