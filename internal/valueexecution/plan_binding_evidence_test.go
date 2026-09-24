package valueexecution

import (
	"slices"
	"testing"
)

func TestPlanExecutionReceiptPreservesBindingEdgeOrder(t *testing.T) {
	source := []byte(`package runtimebinding
namespace runtimebinding

entity Integer id "gooo://runtime-binding/entity/integer"

activity ProposeCandidate(Integer) -> Integer computes "int.add:1"
activity RecordIndependentReview(Integer) -> Integer computes "int.add:1"
activity CommitCandidate(Integer) -> Integer computes "int.add:1"

bind ProposeCandidate.result -> RecordIndependentReview.input
bind RecordIndependentReview.result -> CommitCandidate.input
`)
	plan, err := CompilePlan("typed-plan.gooo", source)
	if err != nil {
		t.Fatalf("CompilePlan() error = %v", err)
	}
	execution, err := plan.Execute(map[string]int64{"ProposeCandidate": 7})
	if err != nil {
		t.Fatalf("Plan.Execute() error = %v", err)
	}
	wantEdges := []string{
		"ProposeCandidate:result->RecordIndependentReview:input",
		"RecordIndependentReview:result->CommitCandidate:input",
	}
	if !slices.Equal(execution.BindingEdgeOrder, wantEdges) {
		t.Fatalf("binding edge order = %#v, want %#v", execution.BindingEdgeOrder, wantEdges)
	}
	wantActivities := []string{"ProposeCandidate", "RecordIndependentReview", "CommitCandidate"}
	if !slices.Equal(execution.Activities, wantActivities) {
		t.Fatalf("activities = %#v, want %#v", execution.Activities, wantActivities)
	}
	if execution.ApplyCalls != 3 || execution.Deliveries != 2 {
		t.Fatalf("execution counts = apply %d deliveries %d, want 3 and 2", execution.ApplyCalls, execution.Deliveries)
	}
}
