package bodyexecution

import (
	"context"
	"os"
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
	m := got.Runtime.InputSeparation
	if m.Status != "UNKNOWN" || m.Reason != "CONSTRUCTION_CALL_INPUTS_NOT_OBSERVED" ||
		m.UnknownInputs != 2 || m.OverlappingInputs != 1 {
		t.Fatal("unrecorded local helper inputs gained an invented separation score", m)
	}
	replayed, err := ReplayJointComposition(ctx, "calls.gooo", source, prior, evaluation, nativeTool())
	if err != nil || !replayed.ConstructionReplayed || replayed.NewModelCalls != 0 || replayed.Runtime.FinitePassed != 3 {
		t.Fatal("helper search did not replay without inference", replayed, err)
	}
}
