package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestJointSearchCallsConstructedRecordHelper(t *testing.T) {
	read := func(name string) []byte {
		t.Helper()
		raw, err := os.ReadFile("../../examples/caller-search-calls/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	cases := func(name string) CompositionCases {
		t.Helper()
		suite, err := DecodeCompositionCases(read(name))
		if err != nil {
			t.Fatal(err)
		}
		return suite
	}
	source := read("main.gooo.fixture")
	construction, evaluation := cases("construction-cases.json"), cases("evaluation-cases.json")
	ctx := context.Background()
	prior, got, err := ConstructAndEvaluateJoint(ctx, "calls.gooo", source, construction, evaluation,
		JointOptions{EntryActivity: "Main", ProgramBudget: 5, GoBinary: nativeTool()})
	if err != nil {
		t.Fatal(err)
	}
	if prior.Decision != "COMPLETE_FINITE" || len(prior.Initial.Preparations) != 2 || len(prior.Attempts) != 1 ||
		got.Runtime.FinitePassed != 3 || got.NewModelCalls != 0 {
		t.Fatal("constructed helper did not reach the actual caller", prior.Decision, got)
	}
	search := prior.Initial.Preparations[1].Generation.Report.BodySearch
	if search == nil || search.SelectedExpression != "1" || search.ProviderOperations != 0 {
		t.Fatal("local search did not retain its selected denominator", search)
	}
	failed := 0
	for _, attempt := range search.Attempts {
		if attempt.TypecheckPassed && !attempt.ScoringCompleted && strings.Contains(attempt.Error, "zero divisor") {
			failed++
		}
	}
	if failed == 0 {
		t.Fatal("local arithmetic failure was omitted from search history", search.Attempts)
	}
	if search.LocalCalls == nil || search.LocalCalls.Status != "RECORDED" || len(search.LocalCalls.Evaluations) < 1 {
		t.Fatal("original search evaluation inputs were not retained", search.LocalCalls)
	}
	first := search.LocalCalls.Evaluations[0]
	if first.EvaluationCompleted || first.CasesCompleted != 1 || len(first.Calls) != 2 || string(first.Calls[1].Inputs) != "[0]" {
		t.Fatal("failed local helper input was lost", first)
	}
	m := got.Runtime.InputSeparation
	if m.Status != "UNKNOWN" || m.Reason != "CONSTRUCTION_CALL_INPUTS_NOT_OBSERVED" ||
		m.UnknownInputs != 1 || m.OverlappingInputs != 2 || m.DisjointInputs != 0 {
		t.Fatal("failed local input exposure or remaining unknown coverage differs", m)
	}
	raw, err := json.MarshalIndent(prior, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	saved, err := DecodeJointConstruction(raw)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := ReplayJointComposition(ctx, "calls.gooo", source, saved, evaluation, nativeTool())
	if err != nil || !replayed.ConstructionReplayed || replayed.NewModelCalls != 0 || replayed.Runtime.FinitePassed != 3 {
		t.Fatal("helper search did not replay without inference", replayed, err)
	}
	if !reflect.DeepEqual(replayed.Runtime.InputSeparation, m) {
		t.Fatal("saved local exposure changed during replay", replayed.Runtime.InputSeparation, m)
	}
}
