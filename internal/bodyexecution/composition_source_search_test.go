package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func sourceSearchCompositionFixture(t *testing.T) ([]byte, CompositionCases) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/source-search-composition.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/source-search-composition-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	return source, suite
}

func TestCompositionSourceSearchGeneratesAndReplaysNativeGraph(t *testing.T) {
	source, suite := sourceSearchCompositionFixture(t)
	prior, err := GenerateComposition(context.Background(), "search.gooo", source, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	if prior.Model != nil || len(prior.Steps) != 3 || strings.Contains(prior.GoooSource, "__GOOO_BODY_HOLE_") || strings.Contains(prior.GoooSource, "assembling") {
		t.Fatal("search did not produce a closed reusable checkpoint", prior.GoooSource)
	}
	for _, step := range prior.Steps[:2] {
		search := step.Generation.Report.BodySearch
		if search == nil || search.ProviderOperations != 0 || search.TrainingPassed != 3 || search.HoldoutPassed != 2 {
			t.Fatal("source search evidence differs", search)
		}
	}
	raw, _ := json.Marshal(prior)
	decoded, err := DecodeComposition(raw)
	if err != nil {
		t.Fatal(err)
	}
	run, err := ExecuteComposition(context.Background(), "search.gooo", source, decoded, suite, nativeTool())
	if err != nil || !run.ProjectionReplayed || !run.RuntimeReplayed || run.ModelCalls != 0 || run.FinitePassed != 15 {
		t.Fatalf("source search native graph: %v %+v", err, run)
	}
	next, err := GenerateComposition(context.Background(), "search.gooo", []byte(prior.GoooSource), suite, "")
	if err != nil || next.Source != prior.Source || next.GoooSource != prior.GoooSource {
		t.Fatal("completed search graph cannot continue as ordinary Gooo", err)
	}
}

func TestCompositionSourceSearchRejectsChangedStageEvidence(t *testing.T) {
	for _, field := range []string{"expression", "holdout", "checkpoint", "source"} {
		t.Run(field, func(t *testing.T) {
			source, suite := sourceSearchCompositionFixture(t)
			prior, err := GenerateComposition(context.Background(), "search.gooo", source, suite, "")
			if err != nil {
				t.Fatal(err)
			}
			switch field {
			case "expression":
				prior.Steps[0].Generation.Report.BodySearch.SelectedExpression = "999999"
			case "holdout":
				prior.Steps[1].Generation.Report.BodySearch.HoldoutCaseResults[0].Actual++
			case "checkpoint":
				prior.Steps[1].InputSourceSHA256 = prior.Steps[0].InputSourceSHA256
			case "source":
				source = []byte(strings.Replace(string(source), "Add two", "Add three", 1))
			}
			if err := VerifyComposition(context.Background(), "search.gooo", source, prior); err == nil {
				t.Fatal("changed source search evidence replayed", field)
			}
		})
	}
}

func TestCompositionSourceSearchPreflightAndModelRoute(t *testing.T) {
	source, suite := sourceSearchCompositionFixture(t)
	bad := []byte(strings.Replace(string(source), "return input > 0", "return missing > 0", 1))
	failed, err := GenerateComposition(context.Background(), "search.gooo", bad, suite, "missing-model.json")
	if err == nil || failed.Stage != "BODY_PREFLIGHT" || failed.Model != nil || !strings.Contains(err.Error(), "missing") {
		t.Fatal("invalid later body was not checked before model loading", err, failed.Stage)
	}
	failed, err = GenerateComposition(context.Background(), "search.gooo", source, suite, "missing-model.json")
	if err == nil || failed.Model != nil || !strings.Contains(err.Error(), "choice-based") {
		t.Fatal("unsupported model route was silently discarded or loaded", err)
	}
}

func TestCompositionSourceSearchConnectsToExistingChoiceAssembly(t *testing.T) {
	source, suite := compositionFixture(t)
	source = append(source, []byte(`
activity Search(Integer) -> Integer computes "return __GOOO_BODY_HOLE_result__" assembling {
    search hole "result" grammar "integer-offset-constant/v1" intent "Preserve the input." max_candidates "8"
    case "-2" -> "-2"
    case "3" -> "3"
    attempts "8"
}
bind Search.result -> Assemble.input
`)...)
	for i := range suite.Cases {
		suite.Cases[i].Inputs["Search"] = suite.Cases[i].Inputs["Assemble"]
		delete(suite.Cases[i].Inputs, "Assemble")
	}
	prior, err := GenerateComposition(context.Background(), "mixed.gooo", source, suite, "")
	if err != nil || prior.Model == nil || prior.Model.Loaded {
		t.Fatal("mixed source search and choice assembly failed", err)
	}
	run, err := ExecuteComposition(context.Background(), "mixed.gooo", source, prior, suite, nativeTool())
	if err != nil || run.FinitePassed != 49 || !run.RuntimeReplayed {
		t.Fatal("mixed graph did not preserve ordinary choice assembly", err, run.FinitePassed)
	}
	_, err = GenerateComposition(context.Background(), "mixed.gooo", source, suite, "missing-model.json")
	if err == nil || !strings.Contains(err.Error(), "load retained structural model") {
		t.Fatal("mixed graph did not route its explicit model to choice assembly", err)
	}
}
