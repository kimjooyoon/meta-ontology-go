package main

import (
	"strings"
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

func TestCollapseDispatchRejectsSingleFieldBindingTampering(t *testing.T) {
	action := collapsePlannerAction(t, "fixture.go:3:value", strings.Repeat("1", 40))
	if failure := validateCollapseAction(action); failure != nil {
		t.Fatalf("real planner action rejected: %#v", failure)
	}
	mutations := []struct {
		name   string
		mutate func(*generation.Action)
	}{
		{name: "executor", mutate: func(candidate *generation.Action) { candidate.Executor = "tampered-executor" }},
		{name: "evaluator", mutate: func(candidate *generation.Action) { candidate.Evaluator = "tampered-evaluator" }},
		{name: "source contract digest", mutate: func(candidate *generation.Action) { candidate.InputContractSourceDigest = strings.Repeat("0", 64) }},
		{name: "semantic contract digest", mutate: func(candidate *generation.Action) { candidate.InputContractSemanticDigest = strings.Repeat("0", 64) }},
		{name: "required indicator", mutate: func(candidate *generation.Action) {
			candidate.RequiredIndicatorIDs = append([]string{}, candidate.RequiredIndicatorIDs...)
			candidate.RequiredIndicatorIDs[0] = "tampered-indicator"
		}},
		{name: "metric producer", mutate: func(candidate *generation.Action) { candidate.MetricProducer = "tampered-producer" }},
		{name: "source proof", mutate: func(candidate *generation.Action) { candidate.SourceIndicator.Proof = sourcepolicy.ProofCoherence }},
		{name: "indicator outcome", mutate: func(candidate *generation.Action) { candidate.IndicatorOutcome.Decision = sourcepolicy.IndicatorDecisionPass }},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			candidate := action
			mutation.mutate(&candidate)
			if failure := validateCollapseAction(candidate); failure == nil || failure.reason != "ACTION_BINDING_INVALID" {
				t.Fatalf("tampered %s validation = %#v, want fail-closed binding rejection", mutation.name, failure)
			}
		})
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
	packageChangedSource := []byte("package alternate\n\ntype receiver struct{}\n\nfunc (receiver) value() int {\n\treturn 1\n}\n")
	packageChanged, err := inspectCollapseSource(packageChangedSource, subject)
	if err != nil {
		t.Fatal(err)
	}
	if failure := validateCollapseOutput(before, packageChanged, packageChangedSource); failure == nil {
		t.Fatal("package-name mutation was accepted")
	}
}

func TestInspectCollapseSourceRejectsBodylessDeclaration(t *testing.T) {
	source := []byte("package fixture\n\nfunc value() int\n")
	if _, err := inspectCollapseSource(source, sourcepolicy.SourceSubject{Path: "fixture.go", Line: 3, Name: "value"}); err == nil {
		t.Fatal("bodyless declaration was accepted")
	}
}

func collapsePlannerAction(t *testing.T, subject, head string) generation.Action {
	t.Helper()
	_, action, _ := collapsePlannerFixture(t, subject, head)
	return action
}

func collapsePlannerFixture(t *testing.T, subject, head string) (generation.Plan, generation.Action, sourcepolicy.Report) {
	t.Helper()
	report, err := sourcepolicy.Evaluate(sourcepolicy.Default(), []sourcepolicy.Observation{
		{Subject: subject, Dimension: sourcepolicy.DimensionRefactorAssign, Value: 2, Detail: "assignment then return result", Producer: "linecaps.Analyze"},
		{Subject: "operations.go:1:executeAction", Dimension: sourcepolicy.DimensionFunctionLines, Value: 76, Producer: "linecaps.Analyze"},
	})
	if err != nil {
		t.Fatal(err)
	}
	plan := generation.Build(strings.Repeat("0", 40), head, report)
	for _, action := range plan.Selected {
		if action.Operation == sourcepolicy.OperationCollapseAssign {
			return plan, action, report
		}
	}
	t.Fatalf("planner did not select collapse action: %#v", plan)
	return generation.Plan{}, generation.Action{}, sourcepolicy.Report{}
}
