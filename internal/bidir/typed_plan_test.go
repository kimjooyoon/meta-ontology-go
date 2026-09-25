package bidir

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestCompileTypedPlanExcludesCrossInvocationFeedback(t *testing.T) {
	document := feedbackDocument(t, feedbackSource)

	plan, err := CompileTypedPlan(document)
	if err != nil {
		t.Fatalf("feedback edge was treated as an in-invocation cycle: %v", err)
	}
	if len(plan.Edges) != 1 {
		t.Fatalf("typed plan edges = %d, want only the in-invocation bind", len(plan.Edges))
	}
	if got := plan.Edges[0]; got.SourceActivity == got.TargetActivity || got.SourcePort != "result" || got.TargetPort != "input" {
		t.Fatalf("typed plan edge = %#v", got)
	}
	if len(plan.Activities) != 2 {
		t.Fatalf("typed plan activities = %d, want both activities", len(plan.Activities))
	}
}

func TestLowerDocumentUsesTypedPlanValidationForTypedCarrier(t *testing.T) {
	file, diagnostics := syntax.ParseFile("typed-plan.gooo", feedbackSource)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}

	document, err := DocumentFromSyntax(file)
	if err != nil {
		t.Fatal(err)
	}
	document.BindingEdges[0].TargetPort = "missing"
	document.RuntimeBindings = nil
	if _, err := LowerDocument(document); err == nil || !strings.Contains(err.Error(), "typed plan") {
		t.Fatalf("lowering did not report the typed-plan defect: %v", err)
	}
}

func TestCompileTypedPlanRejectsAmbiguousPorts(t *testing.T) {
	document := Document{
		Namespace: "ambiguous",
		Declarations: []Declaration{
			{Kind: ActivityKind, ID: "gooo://activity/produce", Name: "Produce", Outputs: []Reference{
				{ID: "gooo://entity/integer", Name: "result"},
				{ID: "gooo://entity/integer", Name: "result"},
			}},
			{Kind: ActivityKind, ID: "gooo://activity/consume", Name: "Consume", Inputs: []Reference{
				{ID: "gooo://entity/integer", Name: "input"},
			}},
		},
		BindingEdges: []BindingEdge{{
			SourceActivity: "gooo://activity/produce", SourcePort: "result",
			TargetActivity: "gooo://activity/consume", TargetPort: "input",
		}},
	}
	if _, err := CompileTypedPlan(document); err == nil || !strings.Contains(err.Error(), "ambiguous output port") {
		t.Fatalf("ambiguous port result = %v, want fail-closed diagnostic", err)
	}
}

func TestTypedPlanDigestIgnoresSourceSpans(t *testing.T) {
	document := feedbackDocument(t, feedbackSource)
	first, err := CompileTypedPlan(document)
	if err != nil {
		t.Fatal(err)
	}
	spanOnly := document
	spanOnly.BindingEdges = append([]BindingEdge(nil), document.BindingEdges...)
	spanOnly.BindingEdges[0].Span = SourceSpan{File: "different.gooo"}
	second, err := CompileTypedPlan(spanOnly)
	if err != nil {
		t.Fatal(err)
	}
	if first.Digest() != second.Digest() || first.Canonical() != second.Canonical() {
		t.Fatalf("source span changed typed plan identity: %s != %s", first.Digest(), second.Digest())
	}
}
