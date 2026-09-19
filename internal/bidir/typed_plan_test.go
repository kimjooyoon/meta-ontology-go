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

func TestDocumentFromSyntaxUsesTypedPlanValidation(t *testing.T) {
	source := strings.Replace(feedbackSource, "Finish.input", "Finish.missing", 1)
	file, diagnostics := syntax.ParseFile("typed-plan.gooo", source)
	if diagnostics.HasErrors() {
		t.Fatal(diagnostics)
	}

	if _, err := DocumentFromSyntax(file); err == nil || !strings.Contains(err.Error(), "typed plan") {
		t.Fatalf("lowering did not report the typed-plan defect: %v", err)
	}
}
