package bodycodegen

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestRecordFlowNestedReturnPreservesUnionOfLivePaths(t *testing.T) {
	body := `let copy = input0
copy.title = "draft"
if input1 { if copy.title == "stop" { return copy }; copy.state = "wait" } else { copy.state = "keep" }
copy.reason = copy.state
return copy`
	source := []byte(strings.Replace(string(flowSource(t, body)), `choice "reason" field_update at "2"`, `choice "reason" field_update at "3"`, 1))
	r := exportedFlow(t, source).ValueFlow
	guard := r.Nodes[r.Choices[2].Guard-1]
	if guard.Kind != "guard_join" || guard.Operator != "or" || guard.Parents[0] == 0 || guard.Parents[1] == 0 {
		t.Fatal("nested return lost union of surviving paths", guard)
	}
	if !flowContains(r, r.Choices[2].First, "join", "") {
		t.Fatal("field definitions were not joined")
	}
}

func TestRecordFlowConstantConditionAndElseIfRemainSourceDerived(t *testing.T) {
	body := `let copy = input0
copy.title = "draft"
if false { return copy }
if input1 { copy.state = "wait" } else if copy == input0 { copy.reason = copy.state }
return copy`
	r := exportedFlow(t, flowSource(t, body)).ValueFlow
	known := false
	for _, node := range r.Nodes {
		known = known || node.KnownTruth == "false"
	}
	if !known || r.Choices[2].Guard == 0 || !flowContains(r, r.Choices[2].Guard, "expression", "==") {
		t.Fatal("constant or fieldwise record comparison was lost")
	}
	guard := r.Nodes[r.Choices[2].Guard-1]
	if guard.Operator != "true" || !flowContains(r, guard.Parents[1], "guard", "false") {
		t.Fatal("else-if parent path was lost")
	}
}

func TestRecordFlowWholeRecordReassignmentDoesNotMutateSnapshot(t *testing.T) {
	body := `let copy = input0
let saved = copy
copy.title = "draft"
copy.state = "wait"
copy.reason = copy.state
copy = saved
return copy`
	r := exportedFlow(t, flowSource(t, body)).ValueFlow
	returns := 0
	for _, node := range r.Nodes {
		if node.Kind == "return" {
			returns++
			if flowContains(r, node.ID, "choice", "") || !flowContains(r, node.ID, "write", "") {
				t.Fatal("whole reassignment retained overwritten field roots")
			}
		}
	}
	if returns != 3 {
		t.Fatal("record return must retain all three roots")
	}
}

func TestRecordFlowBoundsDeclineWithoutPartialGraph(t *testing.T) {
	base := "let copy = input0\ncopy.title = \"draft\"\ncopy.state = \"wait\"\ncopy.reason = copy.state\n"
	var locals, reads strings.Builder
	for i := range 64 {
		fmt.Fprintf(&locals, "let local%d = \"x\"\n", i)
		fmt.Fprintf(&reads, " + local%d", i)
	}
	for _, test := range []struct{ body, reason string }{
		{base + locals.String() + "return Candidate{title: copy.title, state: copy.state, reason: copy.reason" + reads.String() + "}", "FLOW_LOCAL_BOUND"},
		{base + strings.Repeat("copy.reason = \"x\"\n", 260) + "return copy", "FLOW_NODE_BOUND"},
		{"let copy = input0\n" + strings.Repeat("if input1 {", 17) + strings.TrimPrefix(base, "let copy = input0\n") + strings.Repeat("}", 17) + "\nreturn copy", "FLOW_BRANCH_BOUND"},
		{"let copy = input0\nif true { return copy }\n" + strings.TrimPrefix(base, "let copy = input0\n") + "return copy", "FLOW_CHOICE_UNREACHED"},
	} {
		t.Run(test.reason, func(t *testing.T) {
			r, err := ExportRecordAssemblyContextWithFlow(context.Background(), "bound.gooo", flowSource(t, test.body), "Select", false, "")
			if err != nil || r.ValueFlow == nil || r.ValueFlow.Status != "UNRESOLVED" || r.ValueFlow.Reason != test.reason ||
				len(r.ValueFlow.Nodes) != 0 || len(r.ValueFlow.Choices) != 0 || r.ModelPredictions != 0 || r.CandidateTests != 0 {
				t.Fatal("bound observation was incomplete or misleading", r.ValueFlow, err)
			}
		})
	}
}

func TestRecordFlowSpansAndConcurrentSourceObservations(t *testing.T) {
	source := recordUpdatesFixture(t)
	r := exportedFlow(t, source).ValueFlow
	plan, err := prepareRecordAssembly(context.Background(), "r.gooo", source, "Select")
	if err != nil {
		t.Fatal(err)
	}
	for i, choice := range r.Choices {
		span := choice.Span
		if span.View != "computes" || plan.body.body[span.Start:span.End] != plan.choices[i].First {
			t.Fatal("choice byte span differs", span)
		}
		for _, node := range r.Nodes {
			if node.Span.View == "alternative:"+choice.ID {
				if node.Span.Start < 0 || node.Span.End > len(plan.choices[i].Second) || node.Span.Start >= node.Span.End {
					t.Fatal("alternative byte span exceeds its own expression", node)
				}
			}
		}
	}
	var wait sync.WaitGroup
	for range 8 {
		wait.Go(func() {
			next, e := ExportRecordAssemblyContextWithFlow(context.Background(), "r.gooo", source, "Select", false, "")
			if e != nil || !reflect.DeepEqual(r, next.ValueFlow) {
				t.Error("concurrent graph differs", e)
			}
		})
	}
	wait.Wait()
}

func TestRecordFlowCheckpointUsesBaselineByteCoordinates(t *testing.T) {
	source := recordUpdatesFixture(t)
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(context.Background(), "r.gooo", source, "Select")
	if err != nil {
		t.Fatal(err)
	}
	r := exportedFlow(t, []byte(result.GoooSource)).ValueFlow
	if r.BodyView != "baseline" {
		t.Fatal("saved selection graph used selected computes coordinates")
	}
	for _, choice := range r.Choices {
		if choice.Span.View != "baseline" {
			t.Fatal("saved choice span is not baseline-relative")
		}
	}
}
