package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func TestCompositionArithmeticFaultKeepsIndependentDeliveries(t *testing.T) {
	source := []byte(`package faults
namespace faults
entity Integer id "faults://integer"
activity First(Integer) -> Integer computes "return input + 1"
activity Divide(Integer, Integer) -> Integer computes "return input0 / input1"
activity Dependent(Integer) -> Integer computes "return input + 1"
activity Independent(Integer) -> Integer computes "return input * 2"
bind First.result -> Divide.input0
bind Divide.result -> Dependent.input
`)
	suite := calledCompositionCases(t, `{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"First":9007199254740992,"Divide.input1":0,"Independent":7},"expected":{"First":9007199254740993,"Divide":0,"Dependent":1,"Independent":14}},
{"inputs":{"First":7,"Divide.input1":2,"Independent":4},"expected":{"First":8,"Divide":4,"Dependent":5,"Independent":8}}
]}`)
	ctx := context.Background()
	prior, err := GenerateComposition(ctx, "faults.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(ctx, "faults.gooo", source, prior, suite, nativeTool())
	if err != nil || run.Stage != "COMPLETE" || !run.RuntimeReplayed || len(run.Runs) != 2 || run.FinitePassed != 6 || run.FiniteTotal != 8 {
		t.Fatalf("a language fault must retain both rows and independent results: %v %+v", err, run)
	}
	byName := map[string]CompositionDelivery{}
	for i, node := range prior.Plan.Activities {
		byName[node.Name] = run.Traces[0].Deliveries[i]
	}
	if string(byName["First"].Actual) != "9007199254740993" || string(byName["Independent"].Actual) != "14" {
		t.Fatal("earlier or independent results lost", byName)
	}
	if len(byName["Divide"].Actual) != 0 || len(byName["Dependent"].Actual) != 0 {
		t.Fatal("faulted or blocked activities invented a value", byName)
	}
	assertFaultJSON(t, run, 1, 1, 6)
}

// Use the public JSON contract so this regression runs before the new Go types exist.
func assertFaultJSON(t *testing.T, run CompositionRuntime, faulted, blocked, matched int) {
	t.Helper()
	raw, err := json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	var observed struct {
		Outcomes *struct{ Matched, Mismatched, Faulted, Blocked, Unobserved int } `json:"outcomes"`
	}
	if err := json.Unmarshal(raw, &observed); err != nil || observed.Outcomes == nil ||
		observed.Outcomes.Faulted != faulted || observed.Outcomes.Blocked != blocked ||
		observed.Outcomes.Matched != matched || observed.Outcomes.Mismatched != 0 || observed.Outcomes.Unobserved != 0 {
		t.Fatal("expected outcomes not distinguished", err, string(raw))
	}
}

func TestJointCallerArithmeticFaultContinuesOriginalSpace(t *testing.T) {
	source, err := os.ReadFile("../../examples/caller-native-failure/source.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	load := func(name string) CompositionCases {
		raw, err := os.ReadFile("../../examples/caller-native-failure/" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		return calledCompositionCases(t, string(raw))
	}
	cases, evaluation := load("construction-cases"), load("evaluation-cases")
	r, err := ConstructJointComposition(context.Background(), "caller.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 6, GoBinary: nativeTool()})
	if err != nil || r.Stage != "COMPLETE" || r.Schema != "gooo/joint-construction/v6" ||
		r.CandidateSpace != "6" || len(r.Attempts) != 6 || r.SelectedAttempt != 5 || r.Decision != "COMPLETE_FINITE" {
		t.Fatalf("the sixth source candidate must remain reachable: %v %+v", err, r)
	}
	if r.Attempts[4].LocalPassed != 1 || r.Attempts[4].LocalTotal != 1 || !r.Attempts[4].Runtime.RuntimeReplayed {
		t.Fatal("caller failure replaced actual local success", r.Attempts[4])
	}
	replay, err := ReplayJointComposition(context.Background(), "caller.gooo", source, r, evaluation, nativeTool())
	if err != nil || !replay.ConstructionReplayed || replay.NewModelCalls != 0 || replay.Runtime.FinitePassed != 4 {
		t.Fatal("fault history must replay without changing original expectations or invoking a model", err, replay)
	}
	for _, name := range []string{"schema", "fault", "count"} {
		t.Run(name, func(t *testing.T) {
			raw, _ := json.Marshal(r)
			saved, err := DecodeJointConstruction(raw)
			if err != nil {
				t.Fatal(err)
			}
			switch name {
			case "schema":
				saved.Schema = jointFillRejectionSchema
			case "fault":
				saved.Attempts[4].Runtime.Traces[0].Deliveries[0].Fault.Left++
			case "count":
				saved.Attempts[4].Runtime.Outcomes.Faulted = 0
			}
			if _, err := ReplayJointComposition(context.Background(), "caller.gooo", source, saved, evaluation, nativeTool()); err == nil {
				t.Fatal("modified native fault history replayed")
			}
		})
	}
	partial, err := ConstructJointComposition(context.Background(), "caller.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 5, GoBinary: nativeTool()})
	if err != nil || partial.Decision != "PARTIAL_FINITE" || len(partial.Attempts) != 5 ||
		partial.StopReason != "PROGRAM_BUDGET_EXHAUSTED" || !hasCompositionFault(partial.Attempts[4].Runtime) {
		t.Fatal("native failure did not consume the original bounded attempt", err, partial)
	}
}

func TestJointCompletionPrefersNoFaultAtEqualExpectedScore(t *testing.T) {
	clean := JointAttempt{LocalPassed: 1, LocalTotal: 1, Runtime: CompositionRuntime{FinitePassed: 1, FiniteTotal: 1}}
	faulted := clean
	faulted.Runtime.Traces = []CompositionTrace{{Deliveries: []CompositionDelivery{{Fault: &CompositionFault{Kind: "ZERO_DIVISOR"}}}}}
	if jointComplete(faulted) || !jointComplete(clean) || !betterJoint(clean, faulted) || betterJoint(faulted, clean) {
		t.Fatal("a failure outside the named expected outputs outranked a complete program")
	}
}
