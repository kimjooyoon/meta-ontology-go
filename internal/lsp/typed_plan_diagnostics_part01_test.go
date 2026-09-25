package lsp

import (
	"errors"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestValidateTypedPlanForLSPAcceptsExplicitTypedChain(t *testing.T) {
	source := []byte(`package runtimebinding
namespace runtimebinding
entity Integer id "gooo://runtime-binding/entity/integer"
activity ProposeCandidate(Integer) -> Integer computes "int.add:1"
activity RecordIndependentReview(Integer) -> Integer computes "int.add:1"
activity CommitCandidate(Integer) -> Integer computes "int.add:1"
bind ProposeCandidate.result -> RecordIndependentReview.input
bind RecordIndependentReview.result -> CommitCandidate.input
`)
	file, diagnostics := syntax.ParseFile("typed-chain.gooo", string(source))
	if diagnostics.HasErrors() || file == nil {
		t.Fatalf("parse typed chain: diagnostics=%v file=%#v", diagnostics, file)
	}
	if err := validateTypedPlanForLSP(file, syntax.EntityFieldsV1Support()); err != nil {
		t.Fatalf("typed plan validation: %v", err)
	}
	result, err := adaptSyntaxResult("typed-chain.gooo", string(source), file, diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	if !result.semanticValid || len(result.Diagnostics) != 0 {
		t.Fatalf("typed chain LSP result = %#v", result)
	}
}

func TestTypedPlanDiagnosticIsNonAuthorizingAndSourceBound(t *testing.T) {
	source := "package runtimebinding\nnamespace runtimebinding\n"
	diagnostic, err := typedPlanDiagnostic("typed-chain.gooo", source, errors.New("edge mismatch"))
	if err != nil {
		t.Fatal(err)
	}
	if diagnostic.Code != typedPlanDiagnosticCode || diagnostic.Source != "gooo" || diagnostic.Severity != DiagnosticError ||
		diagnostic.Range.Start.Line != 0 || diagnostic.Range.Start.Character != 0 ||
		diagnostic.Range.End.Character != len(source) || diagnostic.Message != "typed runtime plan is invalid: edge mismatch" {
		t.Fatalf("typed plan diagnostic = %#v", diagnostic)
	}
}
