package bodyexecution

import (
	"context"
	"os"
	"testing"
)

func TestCompositionHoleContextCandidatesRunAndReplay(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/hole-context-composition.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/hole-context-composition-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	prior, err := GenerateComposition(ctx, "holes.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	for i, expression := range []string{"1", "3", "input * 4 + 3"} {
		search := prior.Steps[i].Generation.Report.BodySearch
		if search.SelectedExpression != expression || search.CandidateGeneration.HoleContext.EvaluationCalls != 4 {
			t.Fatal("source-owned context was not used", i)
		}
	}
	runtime, err := ExecuteComposition(ctx, "holes.gooo", source, prior, suite, nativeTool())
	if err != nil || runtime.FinitePassed != 12 || runtime.FiniteTotal != 12 || !runtime.RuntimeReplayed || runtime.ModelCalls != 0 {
		t.Fatal("context-generated graph execution differs", err, runtime.FinitePassed)
	}
	if runtime.InputSeparation.DisjointInputs != 1 || runtime.InputSeparation.DisjointCasesPassed != 1 {
		t.Fatal("large integer input observation was lost", runtime.InputSeparation)
	}
	prior.Steps[1].Generation.Report.BodySearch.CandidateGeneration.HoleContext.Probes[0].Expected++
	if _, err := ExecuteComposition(ctx, "holes.gooo", source, prior, suite, nativeTool()); err == nil {
		t.Fatal("altered contextual observation reached native execution")
	}
}
