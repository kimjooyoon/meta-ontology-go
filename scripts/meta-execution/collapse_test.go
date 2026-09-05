package main

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestCollapseDispatchKeepsUnsupportedOperationFailClosed(t *testing.T) {
	_, failure := executeAction("", "", "", generation.Plan{}, generation.Action{Operation: sourcepolicy.OperationObserve}, metaExecutionTrace{})
	if failure == nil || failure.reason != "UNSUPPORTED_SELECTED_OPERATION" {
		t.Fatalf("unsupported operation failure = %#v, want fail-closed unsupported selection", failure)
	}
}

func TestCollapseDispatchRejectsContaminatedBinding(t *testing.T) {
	binding, ok := generation.BindingForOperation(generation.DefaultRegistry(), sourcepolicy.OperationCollapseAssign)
	if !ok {
		t.Fatal("canonical collapse binding unavailable")
	}
	action := generation.Action{
		IndicatorID:                 "indicator",
		Subject:                     "fixture.go:3:value",
		SubjectKind:                 binding.InputSubjectKind,
		InputSubjectKind:            binding.InputSubjectKind,
		InputContractSourceDigest:   binding.InputContractSourceDigest,
		InputContractSemanticDigest: binding.InputContractSemanticDigest,
		Operation:                   binding.Operation,
		Activity:                    binding.Activity,
		Output:                      binding.Output,
		IndependenceGroupID:         binding.IndependenceGroupID,
		ProofChoice:                 binding.ProofChoice,
		MetricProofChoice:           sourcepolicy.ProofRegression,
		Executor:                    "tampered-executor",
		Evaluator:                   binding.Evaluator,
		RequiredIndicatorIDs:        binding.RequiredIndicatorIDs,
		ReceiptRequired:             binding.ReceiptRequired,
		Priority:                    binding.Priority,
		SourceIndicator: sourcepolicy.Indicator{
			MetricID: sourcepolicy.DimensionRefactorAssign, Subject: "fixture.go:3:value",
			SubjectKind: binding.InputSubjectKind, Applicability: sourcepolicy.ApplicabilityApplicable,
			Operation: binding.Operation, Satisfied: false,
		},
	}
	_, failure := executeAction("", "", "", generation.Plan{}, action, metaExecutionTrace{})
	if failure == nil || failure.reason != "ACTION_BINDING_INVALID" {
		t.Fatalf("contaminated binding failure = %#v, want fail-closed binding rejection", failure)
	}
}

func TestInspectCollapseSourceBindsExactSubjectAndReceiver(t *testing.T) {
	source := []byte("package fixture\n\ntype receiver struct{}\n\nfunc (receiver) value() int {\n\tresult := 1\n\treturn result\n}\n")
	subject := sourcepolicy.SourceSubject{Path: "fixture.go", Line: 5, Name: "method value"}
	before, err := inspectCollapseSource(source, subject)
	if err != nil {
		t.Fatal(err)
	}
	if !before.AssignmentReturn || !before.CommentsPreserved || before.Receiver != "(receiver)" {
		t.Fatalf("before inspection = %#v", before)
	}
	afterSource := []byte("package fixture\n\ntype receiver struct{}\n\nfunc (receiver) value() int {\n\treturn 1\n}\n")
	after, err := inspectCollapseSource(afterSource, subject)
	if err != nil {
		t.Fatal(err)
	}
	if !after.SingleReturn || after.ReturnExpression != before.ReturnExpression || after.Receiver != before.Receiver {
		t.Fatalf("after inspection = %#v", after)
	}
	if failure := validateCollapseOutput(before, after, afterSource); failure != nil {
		t.Fatalf("valid collapse rejected: %#v", failure)
	}
}
