package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func sourceFillCompositionFixture(t *testing.T) ([]byte, CompositionCases) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/source-fill-composition.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/source-fill-composition-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	return source, suite
}

func TestCompositionSourceFillNativeGraphAndReuse(t *testing.T) {
	source, suite := sourceFillCompositionFixture(t)
	prior, err := GenerateComposition(context.Background(), "fill.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	if prior.FillModel != nil || prior.Model != nil || len(prior.Steps) != 3 {
		t.Fatal("unexpected model or steps")
	}
	for _, step := range prior.Steps[:2] {
		fill := step.Generation.Report.BodyFill
		if fill == nil || fill.TestCasesPassed != 3 || fill.HoldoutCasesPassed != 1 || fill.LocalModelPredictions != 0 {
			t.Fatal("body-fill evidence missing", fill)
		}
	}
	raw, _ := json.MarshalIndent(prior, "", "  ")
	decoded, err := DecodeComposition(raw)
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(context.Background(), "fill.gooo", source, decoded, suite, nativeTool())
	if err != nil || run.FinitePassed != 15 || run.ModelCalls != 0 || !run.RuntimeReplayed {
		t.Fatal("native fill graph", err, run)
	}
	next, err := GenerateComposition(context.Background(), "fill.gooo", []byte(prior.GoooSource), suite, "")
	if err != nil || next.Source != prior.Source || next.GoooSource != prior.GoooSource {
		t.Fatal("ordinary checkpoint differs", err)
	}
}

func TestCompositionSourceFillPreflightBeforeModel(t *testing.T) {
	source, suite := sourceFillCompositionFixture(t)
	bad := []byte(strings.Replace(string(source), "return input > 0", "return missing > 0", 1))
	options := CompositionOptions{FillModelPath: "missing.json"}
	failed, err := GenerateCompositionWithOptions(context.Background(), "fill.gooo", bad, suite, options)
	if err == nil || failed.Stage != "BODY_PREFLIGHT" || failed.FillModel != nil || strings.Contains(err.Error(), "load retained") {
		t.Fatal("model loaded before full preflight", err)
	}
	_, err = GenerateCompositionWithOptions(context.Background(), "fill.gooo", source, suite, options)
	if err == nil || !strings.Contains(err.Error(), "load retained body-fill model") {
		t.Fatal("fill model ignored", err)
	}
	_, err = GenerateComposition(context.Background(), "fill.gooo", source, suite, "missing.json")
	if err == nil || !strings.Contains(err.Error(), "choice-based") {
		t.Fatal("wrong profile accepted", err)
	}
	search, searchCases := sourceSearchCompositionFixture(t)
	_, err = GenerateCompositionWithOptions(context.Background(), "search.gooo", search, searchCases, options)
	if err == nil || !strings.Contains(err.Error(), "requires a source_fill") {
		t.Fatal("unused fill model ignored", err)
	}
}

func TestCompositionSourceFillDerivedRecord(t *testing.T) {
	source, err := os.ReadFile("../../examples/body-codegen/source-ir-fill-record-tiny.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	source = append(source, []byte(`
entity Boolean id "recordfilltiny://boolean"
activity Accepted(Review) -> Boolean computes "return input.decision == \"accepted\""
bind ReviewCandidate.result -> Accepted.input
`)...)
	suite, err := DecodeCompositionCases([]byte(`{"schema":"gooo/body-composition-cases/v1","cases":[
{"inputs":{"ReviewCandidate":{"reviewed":true,"state":"ready"}},"expected":{"ReviewCandidate":{"decision":"accepted"},"Accepted":true}},
{"inputs":{"ReviewCandidate":{"reviewed":true,"state":"queued"}},"expected":{"ReviewCandidate":{"decision":"rejected"},"Accepted":false}}
]}`))
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateComposition(context.Background(), "record.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.MarshalIndent(prior, "", "  ")
	decoded, err := DecodeComposition(raw)
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(context.Background(), "record.gooo", source, decoded, suite, nativeTool())
	if err != nil || run.FiniteTotal != 4 || !run.RuntimeReplayed {
		t.Fatal("derived record graph", err, run)
	}
	if prior.Steps[0].Generation.Report.BodyFill.CandidateGeneration == nil {
		t.Fatal("lost derived assignment evidence")
	}
	decoded.Steps[0].Generation.Report.BodyFill.SelectedValueHoldoutCaseResults[0].Passed = !decoded.Steps[0].Generation.Report.BodyFill.SelectedValueHoldoutCaseResults[0].Passed
	if err := VerifyComposition(context.Background(), "record.gooo", source, decoded); err == nil {
		t.Fatal("changed record holdout replayed")
	}
}
