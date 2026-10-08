package bodyexecution

import (
	"context"
	"strings"
	"testing"
)

func TestJointDependencyChangesRecheckCallerLocalCases(t *testing.T) {
	source := []byte(`package nested
namespace nested
entity Integer id "nested://integer"
entity Box id "nested://box" fields { field value id "nested://value" type integer required one }
activity Seed(Integer) -> Box computes "return Box{value: input}" assembling {
 choice "value" field_value at "0" alternative "input * 2" intent "Double when needed by the caller."
 value_case "[0]" -> "{\"value\":0}"
 attempts "2"
}
activity Wrap(Integer) -> Box computes "let next = Seed(input); return Box{value: next.value * 2}" assembling {
 choice "value" field_value at "0" alternative "next.value" intent "Keep the locally required result as Seed changes."
 value_case "[1]" -> "{\"value\":2}"
 attempts "2"
}
activity Main(Integer) -> Integer computes "let a = Wrap(input); let b = Seed(input); return a.value + b.value"
`)
	cases := jointCases(t, `[{"inputs":{"Main":3},"expected":{"Main":12}}]`)
	r, err := ConstructJointComposition(context.Background(), "nested.gooo", source, cases,
		JointOptions{EntryActivity: "Main", ProgramBudget: 4, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if r.Decision != "COMPLETE_FINITE" || len(r.Attempts) != 4 || r.SelectedAttempt != 3 {
		t.Fatal("nested construction", r)
	}
	for i, want := range []int{1, 0, 0, 1} {
		if r.Attempts[i].Candidates[1].Activity != "Wrap" || r.Attempts[i].Candidates[1].Attempt.Passed != want {
			t.Fatal("dependent local score was reused under a changed Seed", i, r.Attempts[i])
		}
	}
	if r.Attempts[0].Candidates[1].InputSourceSHA256 == r.Attempts[2].Candidates[1].InputSourceSHA256 {
		t.Fatal("changed dependency checkpoint was not retained")
	}
	if _, err := ReplayJointComposition(context.Background(), "nested.gooo", source, r, cases, nativeTool()); err != nil {
		t.Fatal(err)
	}
}

func TestJointConstructionPreservesExplicitBinds(t *testing.T) {
	source := []byte(strings.Replace(jointAmbiguousSource,
		`activity Main(Integer) -> Integer computes "let result = Choose(input); return result.value"`,
		"activity Main(Box) -> Integer computes \"return input.value\"\nbind Choose.result -> Main.input", 1))
	cases := jointCases(t, `[{"inputs":{"Choose":3},"expected":{"Main":6}}]`)
	r, err := ConstructJointComposition(context.Background(), "bound.gooo", source, cases,
		JointOptions{ProgramBudget: 2, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if r.Decision != "COMPLETE_FINITE" || len(r.Selected.Plan.Edges) != 1 || len(r.Selected.Steps) != 2 {
		t.Fatal("explicit data-flow edge was not retained", r)
	}
}
