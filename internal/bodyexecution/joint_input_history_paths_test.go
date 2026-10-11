package bodyexecution

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestJointInputsIncludeConditionsAndExactIntegers(t *testing.T) {
	construction := jointCases(t, `[{"inputs":{"Choose":11},"expected":{"Choose":12}}]`)
	evaluation := jointCases(t, `[
 {"inputs":{"Choose":7},"expected":{"Choose":8}},
 {"inputs":{"Choose":11},"expected":{"Choose":12}},
 {"inputs":{"Choose":9},"expected":{"Choose":10}},
 {"inputs":{"Choose":9007199254740993},"expected":{"Choose":9007199254740994}},
 {"inputs":{"Choose":-9007199254740993},"expected":{"Choose":-9007199254740993}}]`)
	prior, got, err := ConstructAndEvaluateJoint(context.Background(), "joint-condition.gooo", []byte(conditionInputSource),
		construction, evaluation, JointOptions{EntryActivity: "Choose", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	m := got.Runtime.InputSeparation
	if got.Runtime.FinitePassed != 5 || m.Status != "PASS" || m.UniqueInputs != 5 || m.OverlappingInputs != 4 ||
		m.DisjointInputs != 1 || m.DisjointCasesPassed != 1 || got.Runtime.ConstructionInputHistorySHA256 != compositionDigest(prior) {
		t.Fatal("declared or consumed input lost", m)
	}
	evaluation.Cases = append(evaluation.Cases, evaluation.Cases[2])
	evaluation.Cases[5].Expected = jointCases(t, `[{"inputs":{"Choose":9},"expected":{"Choose":0}}]`).Cases[0].Expected
	replayed, err := ReplayJointComposition(context.Background(), "joint-condition.gooo", []byte(conditionInputSource), prior, evaluation, nativeTool())
	if err != nil {
		t.Fatal(err)
	}
	m = replayed.Runtime.InputSeparation
	if m.Status != "PROGRESS" || m.DuplicateRows != 1 || m.DisjointInputs != 1 || m.DisjointCasesPassed != 0 {
		t.Fatal("conflicting duplicate expectation disappeared", m)
	}
}

func TestJointInputsIncludeFailedCandidateRoutes(t *testing.T) {
	source := `package pastpaths
namespace pastpaths
entity Integer id "pastpaths://integer"
entity Box id "pastpaths://box" fields { field value id "pastpaths://value" type integer required one }
activity Left(Integer) -> Box computes "return Box{value: input}" assembling {
 choice "left" field_value at "0" alternative "input * 2" intent "double"
 value_case "[0]" -> "{\"value\":0}"
 attempts "2"
}
activity Right(Integer) -> Box computes "return Box{value: input}" assembling {
 choice "right" field_value at "0" alternative "input * 3" intent "triple"
 value_case "[1]" -> "{\"value\":1}"
 attempts "2"
}
activity Main(Integer) -> Integer computes "let a = Left(input); let b = Right(a.value); return b.value"
`
	construction := jointCases(t, `[{"inputs":{"Main":4},"expected":{"Main":8}}]`)
	evaluation := jointCases(t, `[{"inputs":{"Main":2},"expected":{"Main":4}},{"inputs":{"Main":3},"expected":{"Main":6}}]`)
	prior, got, err := ConstructAndEvaluateJoint(context.Background(), "pastpaths.gooo", []byte(source), construction, evaluation,
		JointOptions{EntryActivity: "Main", ProgramBudget: 4, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	m := got.Runtime.InputSeparation
	if len(prior.Attempts) != 3 || prior.SelectedAttempt != 2 || got.Runtime.FinitePassed != 2 ||
		m.Status != "PASS" || m.OverlappingInputs != 1 || m.DisjointInputs != 1 || m.DisjointCasesPassed != 1 {
		t.Fatal("failed candidates' helper inputs were forgotten", m, len(prior.Attempts))
	}
}

func TestJointInputsObserveRepeatedNestedCalls(t *testing.T) {
	source := strings.Replace(jointAmbiguousSource,
		`activity Main(Integer) -> Integer computes "let result = Choose(input); return result.value"`,
		`activity Wrap(Integer) -> Integer computes "let result = Choose(input * input); return result.value"
activity Main(Integer) -> Integer computes "return Wrap(input) + Wrap(input + 1)"`, 1)
	construction := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":50}}]`)
	evaluation := jointCases(t, `[{"inputs":{"Main":-4},"expected":{"Main":50}},{"inputs":{"Main":5},"expected":{"Main":122}}]`)
	prior, got, err := ConstructAndEvaluateJoint(context.Background(), "nested-inputs.gooo", []byte(source), construction, evaluation,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	trace := prior.Attempts[0].Runtime.Traces[0]
	if len(trace.Calls) != 2 || fmt.Sprint(trace.Calls[0].Inputs) == fmt.Sprint(trace.Calls[1].Inputs) {
		t.Fatal("repeated nested helper inputs were collapsed", trace)
	}
	m := got.Runtime.InputSeparation
	if got.Runtime.FinitePassed != 2 || m.OverlappingInputs != 1 || m.DisjointInputs != 1 || m.DisjointCasesPassed != 1 || m.Status != "PASS" {
		t.Fatal("nested calls were replaced by caller input identity", m)
	}
}

func TestJointUnknownHistoryStillCountsObservedRoots(t *testing.T) {
	graph := compositionGraph{count: 1}
	graph.nodes[0] = CompositionActivity{Name: "Main", ID: "main", InputType: "Integer", OutputType: "Integer", InputFrom: -1}
	suite := jointCases(t, `[{"inputs":{"Main":1},"expected":{"Main":1}},
 {"inputs":{"Main":1},"expected":{"Main":1}},{"inputs":{"Main":2},"expected":{"Main":2}}]`)
	m := graph.unclassifiedJointInputs(suite, unknownCompositionInputSeparation("SELECTION_INPUTS_UNAVAILABLE"))
	if m.Status != "UNKNOWN" || m.UniqueInputs != 2 || m.UnknownInputs != 2 || m.DuplicateRows != 1 ||
		m.DisjointInputs != 0 || m.Reason != "SELECTION_INPUTS_UNAVAILABLE" {
		t.Fatal("missing history was presented as zero observed inputs", m)
	}
}
