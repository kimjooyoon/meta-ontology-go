package bidir

import "testing"

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
