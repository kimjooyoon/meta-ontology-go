package bodyexecution

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

const jointAmbiguousSource = `package ambiguous
namespace ambiguous
entity Integer id "ambiguous://integer"
entity Box id "ambiguous://box" fields { field value id "ambiguous://value" type integer required one }
activity Choose(Integer) -> Box computes "return Box{value: input}" assembling {
 choice "value" field_value at "0" alternative "input * 2" intent "Use the doubled input when the caller requires it."
 value_case "[0]" -> "{\"value\":0}"
 attempts "2"
}
activity Main(Integer) -> Integer computes "let result = Choose(input); return result.value"
`

func jointCases(t *testing.T, text string) CompositionCases {
	t.Helper()
	r, err := DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":` + text + `}`))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestJointCallerFailureReopensLocallyCompleteHelper(t *testing.T) {
	ctx := context.Background()
	source := []byte(jointAmbiguousSource)
	training := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":6}}]`)
	r, err := ConstructJointComposition(ctx, "joint.gooo", source, training, JointOptions{
		EntryActivity: "Main", ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if r.Decision != "COMPLETE_FINITE" || r.SelectedAttempt != 1 || len(r.Attempts) != 2 || r.CandidateSpace != "2" {
		t.Fatalf("joint result: %+v", r)
	}
	for _, a := range r.Attempts {
		if a.LocalPassed != 1 || a.LocalTotal != 1 {
			t.Fatal("local obligations changed", a)
		}
	}
	if r.Attempts[0].Runtime.FinitePassed != 0 || r.Attempts[1].Runtime.FinitePassed != 1 ||
		r.Initial.Preparations[0].Generation.Report.RecordAssembly.SelectedMask != 0 || strings.Contains(r.SelectedSource, "assembling") {
		t.Fatal("caller did not choose a different source-owned body")
	}
	// Round-trip the receipt, including indented RawMessage case values.
	raw, _ := json.MarshalIndent(r, "", "  ")
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	evaluation := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":6}},{"inputs":{"Main":4},"expected":{"Main":8}},{"inputs":{"Main":4},"expected":{"Main":8}}]`)
	got, err := ReplayJointComposition(ctx, "joint.gooo", source, saved, evaluation, nativeTool())
	if err != nil {
		t.Fatal(err)
	}
	if !got.ConstructionReplayed || got.NewModelCalls != 0 || got.Runtime.FinitePassed != 3 ||
		got.InputSeparation.UniqueInputs != 2 || got.InputSeparation.DuplicateRows != 1 ||
		got.InputSeparation.ConstructionInputs != 1 || got.InputSeparation.OtherInputs != 1 {
		t.Fatal("feedback was not separated from evaluation", got)
	}
}

func TestJointConstructionCrossesTwoHelperPlateau(t *testing.T) {
	source := []byte(`package paired
namespace paired
entity Integer id "paired://integer"
entity Box id "paired://box" fields { field value id "paired://value" type integer required one }
activity Left(Integer) -> Box computes "return Box{value: -input}" assembling {
 choice "value" field_value at "0" alternative "input" intent "positive value"
 value_case "[0]" -> "{\"value\":0}"
 attempts "2"
}
activity Right(Integer) -> Box computes "return Box{value: -input}" assembling {
 choice "value" field_value at "0" alternative "input" intent "positive value"
 value_case "[0]" -> "{\"value\":0}"
 attempts "2"
}
activity Main(Integer) -> Integer computes "let a = Left(input); let b = Right(input); if a.value > 0 && b.value > 0 { return 1 }; return 0"
`)
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":1}}]`)
	for _, budget := range []int{1, 2, 4} {
		r, err := ConstructJointComposition(context.Background(), "paired.gooo", source, cases,
			JointOptions{EntryActivity: "Main", ProgramBudget: budget, GoBinary: nativeTool()})
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Attempts) != budget || r.CandidateSpace != "4" {
			t.Fatal("joint budget", r)
		}
		if budget < 4 && (r.Decision != "PARTIAL_FINITE" || r.StopReason != "PROGRAM_BUDGET_EXHAUSTED") {
			t.Fatal("plateau should remain partial", r)
		}
		if budget == 4 && (r.Decision != "COMPLETE_FINITE" || r.SelectedAttempt != 3) {
			t.Fatal("both helpers must change together", r)
		}
	}
}

func TestJointCallerSuccessDoesNotEraseConflictingHelperObligation(t *testing.T) {
	source := []byte(strings.Replace(jointAmbiguousSource,
		`value_case "[0]" -> "{\"value\":0}"`, `value_case "[3]" -> "{\"value\":3}"`, 1))
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":6}}]`)
	r, err := ConstructJointComposition(context.Background(), "conflict.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 8, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if r.Decision != "PARTIAL_FINITE" || r.StopReason != "DECLARED_SPACE_EXHAUSTED" || len(r.Attempts) != 2 ||
		r.SelectedAttempt != 0 || r.Attempts[1].LocalPassed != 0 || r.Attempts[1].Runtime.FinitePassed != 1 {
		t.Fatal("contradictory caller must not override local obligation", r)
	}
}

func TestJointPreflightAndCancellationBeforeModelLoad(t *testing.T) {
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":6}}]`)
	for _, budget := range []int{0, 65} {
		_, err := ConstructJointComposition(context.Background(), "joint.gooo", []byte(jointAmbiguousSource), cases,
			JointOptions{ProgramBudget: budget, ModelPath: "missing-model"})
		if err == nil || strings.Contains(err.Error(), "load") {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ConstructJointComposition(ctx, "joint.gooo", []byte(jointAmbiguousSource), cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 2, ModelPath: "missing-model"})
	if err == nil || strings.Contains(err.Error(), "load") {
		t.Fatal(err)
	}
}
