package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

const jointPathsSource = `package callerpaths
namespace callerpaths
entity Integer id "callerpaths://integer"
activity Choose(Integer) -> Integer computes "if input < 0 { return input } else { return 0 - input }" assembling {
    choice "sign" branch_layout at "0" intent "호출자에 맞는 분기를 고른다. Choose the branch needed by the caller."
    case "0" -> "0"
    attempts "2"
}
activity Main(Integer) -> Integer computes "return Choose(input) + input"
`

func TestJointTypedPathsCallerReopensLocalWinner(t *testing.T) {
	source := []byte(jointPathsSource)
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":6}}]`)
	prior, err := ConstructJointComposition(context.Background(), "paths.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 8, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if prior.Schema != "gooo/joint-construction/v7" || prior.Decision != "COMPLETE_FINITE" ||
		len(prior.Attempts) != 2 || prior.CandidateSpace != "2" ||
		prior.Attempts[0].Runtime.FinitePassed != 0 || prior.Attempts[1].Runtime.FinitePassed != 1 {
		t.Fatal("caller did not reopen the locally complete branch", prior)
	}
	for _, attempt := range prior.Attempts {
		if attempt.LocalPassed != 1 || attempt.LocalTotal != 1 {
			t.Fatal("local source case was lost", attempt)
		}
	}
	evaluation := jointCases(t, `[{"inputs":{"Main":7},"expected":{"Main":14}},{"inputs":{"Main":-5},"expected":{"Main":0}},{"inputs":{"Main":9007199254740993},"expected":{"Main":18014398509481986}}]`)
	raw, _ := json.Marshal(prior)
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := ReplayJointComposition(context.Background(), "paths.gooo", source, saved, evaluation, nativeTool())
	if err != nil || replay.Runtime.FinitePassed != 3 || replay.NewModelCalls != 0 ||
		!replay.ConstructionReplayed || replay.InputSeparation.OtherInputs != 3 {
		t.Fatal(replay, err)
	}
}

func TestJointTypedPathsRespectsSourceAndProgramBudgets(t *testing.T) {
	for _, tc := range []struct {
		name, source, space, stop string
		budget                    int
	}{
		{"program", jointPathsSource, "2", "PROGRAM_BUDGET_EXHAUSTED", 1},
		{"source", strings.Replace(jointPathsSource, `attempts "2"`, `attempts "1"`, 1), "1", "DECLARED_SPACE_EXHAUSTED", 8},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := []byte(tc.source)
			cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":6}}]`)
			prior, err := ConstructJointComposition(context.Background(), "bounded-paths.gooo", source, cases,
				JointOptions{EntryActivity: "Main", ProgramBudget: tc.budget, GoBinary: nativeTool()})
			if err != nil || len(prior.Attempts) != 1 || prior.CandidateSpace != tc.space ||
				prior.Decision != "PARTIAL_FINITE" || prior.StopReason != tc.stop {
				t.Fatal(prior, err)
			}
			if _, err := ReplayJointComposition(context.Background(), "bounded-paths.gooo", source, prior, cases, nativeTool()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestJointTypedPathsRetainsRejectedCombinedAssignments(t *testing.T) {
	source := []byte(`package pathreject
namespace pathreject
entity Integer id "pathreject://integer"
activity Choose(Integer) -> Integer computes "let x = input; let y = input + 1; let z = x; return z + (y * 0) + (x * 0)" assembling {
    choice "reference" local_reference at "0" alternative "y" intent "Choose the local value."
    choice "order" root_order at "1" intent "Choose the declaration order."
    case "0" -> "999"
    attempts "4"
}
activity Main(Integer) -> Integer computes "return Choose(input)"
`)
	cases := jointCases(t, `[{"inputs":{"Main":1},"expected":{"Main":2}}]`)
	prior, err := ConstructJointComposition(context.Background(), "rejected-paths.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 4, GoBinary: nativeTool()})
	if err != nil || len(prior.Attempts) != 4 || prior.Decision != "PARTIAL_FINITE" {
		t.Fatal(prior, err)
	}
	last := prior.Attempts[3]
	if last.Rejection == nil || last.Rejection.Stage != "LOCAL_TYPED_PATH" ||
		last.PathCandidates[0].Stage != "TYPE_CHECK" || last.LocalTotal != 0 || last.Runtime.Stage != "" ||
		last.PathCandidates[0].CandidateSet.Declared != 4 {
		t.Fatal("rejected interacting edits received a caller score", last)
	}
	if _, err := ReplayJointComposition(context.Background(), "rejected-paths.gooo", source, prior, cases, nativeTool()); err != nil {
		t.Fatal(err)
	}
}

func TestJointTypedPathsReplayRechecksFullObservation(t *testing.T) {
	source := []byte(jointPathsSource)
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":6}}]`)
	prior, err := ConstructJointComposition(context.Background(), "paths.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(prior)
	for name, change := range map[string]func(*JointConstruction){
		"missing": func(r *JointConstruction) { r.Attempts[0].PathCandidates = nil },
		"mask":    func(r *JointConstruction) { r.Attempts[0].PathCandidates[0].Mask = 1 },
		"actual":  func(r *JointConstruction) { r.Attempts[0].PathCandidates[0].Cases[0].Actual = 99 },
		"source":  func(r *JointConstruction) { r.Attempts[0].PathCandidates[0].InputSourceSHA256 = "different" },
		"choice":  func(r *JointConstruction) { r.Attempts[0].PathCandidates[0].Choices["sign"] = "undeclared" },
		"bound":   func(r *JointConstruction) { r.Attempts[0].PathCandidates[0].CandidateSet.Masks = []uint16{0} },
		"kind":    func(r *JointConstruction) { r.CandidateKinds[0] = "record_mask" },
		"schema":  func(r *JointConstruction) { r.Schema = jointFaultSchema },
	} {
		t.Run(name, func(t *testing.T) {
			saved, err := DecodeJointConstruction(raw)
			if err != nil {
				t.Fatal(err)
			}
			change(&saved)
			if _, err := ReplayJointComposition(context.Background(), "paths.gooo", source, saved, cases, nativeTool()); err == nil {
				t.Fatal("changed typed path observation replayed")
			}
		})
	}
}

func TestJointTypedPathsAndRecordChoicesCompose(t *testing.T) {
	source := strings.Replace(jointPathsSource, `activity Main(Integer) -> Integer computes "return Choose(input) + input"`, `
entity Box id "callerpaths://box" fields { field value id "callerpaths://value" type integer required one }
activity Pick(Integer) -> Box computes "return Box{value:input}" assembling {
    choice "double" field_value at "0" alternative "input * 2" intent "Double the value."
    value_case "[0]" -> "{\"value\":0}"
    attempts "2"
}
activity Main(Integer) -> Integer computes "let x = Choose(input); let p = Pick(input); return x + p.value"`, 1)
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":9}}]`)
	prior, err := ConstructJointComposition(context.Background(), "mixed-paths.gooo", []byte(source), cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 4, GoBinary: nativeTool()})
	if err != nil || prior.Decision != "COMPLETE_FINITE" || prior.CandidateSpace != "4" || len(prior.Attempts) != 4 {
		t.Fatal(prior, err)
	}
	for _, attempt := range prior.Attempts {
		if len(attempt.Candidates) != 1 || len(attempt.PathCandidates) != 1 || attempt.LocalPassed != 2 || attempt.LocalTotal != 2 {
			t.Fatal("mixed local cases were combined incorrectly", attempt)
		}
	}
	if _, err := ReplayJointComposition(context.Background(), "mixed-paths.gooo", []byte(source), prior, cases, nativeTool()); err != nil {
		t.Fatal(err)
	}
}

func TestJointTypedPathsRetainsNativeCallerFault(t *testing.T) {
	source := []byte(strings.Replace(jointPathsSource,
		"return Choose(input) + input", "return input / (Choose(input) - input)", 1))
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":1}}]`)
	prior, err := ConstructJointComposition(context.Background(), "path-fault.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil || prior.Schema != jointPathSchema || len(prior.Attempts) != 2 ||
		prior.Decision != "PARTIAL_FINITE" || !hasCompositionFault(prior.Attempts[1].Runtime) {
		t.Fatal(prior, err)
	}
	if _, err := ReplayJointComposition(context.Background(), "path-fault.gooo", source, prior, cases, nativeTool()); err != nil {
		t.Fatal(err)
	}
}
