package bodycodegen

import (
	"context"
	"os"
	"testing"
)

func sourceSearchReplayFixture(t *testing.T) ([]byte, Result) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/ir-search-source.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	spec, err := SourceAssembly(context.Background(), "search.gooo", source, "ClampNegativeToZero")
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateWithSourceIRSearch(context.Background(), "search.gooo", source, "ClampNegativeToZero", spec, "", "")
	if err != nil {
		t.Fatal(err)
	}
	return source, prior
}

func TestSourceIRSearchReplayRejectsUndeclaredExpression(t *testing.T) {
	source, prior := sourceSearchReplayFixture(t)
	if err := VerifyIRBodySearchProjection(context.Background(), "search.gooo", source, prior); err != nil {
		t.Fatalf("original generated projection did not replay: %v", err)
	}
	_, activity, _, body, err := prepareBodySearch("search.gooo", source, prior.Report.Activity, "floor")
	if err != nil {
		t.Fatal(err)
	}
	filled, err := replaceIdentifier(body, bodyFillHoleToken("floor"), "input + 1234567")
	if err != nil {
		t.Fatal(err)
	}
	changed, err := replaceActivityProgram(source, activity.ValueProgramSpan, filled)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := Generate("search.gooo", changed, prior.Report.Activity)
	if err != nil {
		t.Fatal(err)
	}
	projection.Report.BodySearch = prior.Report.BodySearch
	projection.Report.BodySearch.SelectedExpression = "input + 1234567"
	populateCompletenessReceipt(&projection.Report, "")
	if err := VerifyIRBodySearchProjection(context.Background(), "search.gooo", source, projection); err == nil {
		t.Fatal("replay accepted a typed expression outside the source-declared candidate set")
	}
}

func TestSourceIRSearchReplayRejectsChangedFiniteEvidence(t *testing.T) {
	for _, field := range []string{"plan", "candidate", "training", "holdout"} {
		t.Run(field, func(t *testing.T) {
			source, prior := sourceSearchReplayFixture(t)
			search := prior.Report.BodySearch
			switch field {
			case "plan":
				search.IRPlanSHA256 = digest([]byte("another plan"))
			case "candidate":
				search.SelectedCandidateID = "another candidate"
			case "training":
				search.TrainingCaseResults[0].Input = 100
			case "holdout":
				search.HoldoutCaseResults[0].Actual = 100
			}
			if err := VerifyIRBodySearchProjection(context.Background(), "search.gooo", source, prior); err == nil {
				t.Fatal("source replay accepted changed " + field + " evidence")
			}
		})
	}
}

func TestSourceIRSearchReplayPreservesObservedZeroMatches(t *testing.T) {
	source := []byte(`package zero
namespace zero
entity Integer id "zero://integer"
activity Constant(Integer) -> Integer computes "return 42 + (__GOOO_BODY_HOLE_value__ * 0)" assembling {
    search hole "value" grammar "integer-offset-constant/v1" intent "Try to return zero." max_candidates "4"
    case "1" -> "0"
    holdout_case "2" -> "0"
    attempts "2"
}
`)
	spec, err := SourceAssembly(context.Background(), "zero.gooo", source, "Constant")
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateWithSourceIRSearch(context.Background(), "zero.gooo", source, "Constant", spec, "", "")
	if err != nil {
		t.Fatal(err)
	}
	search := prior.Report.BodySearch
	if search.TrainingPassed != 0 || search.TrainingTotal != 1 || search.HoldoutPassed != 0 || search.HoldoutTotal != 1 {
		t.Fatalf("expected measured 0/1 for both finite suites: %#v", search)
	}
	for _, d := range prior.Report.CompletenessReceipt.Dimensions {
		if (d.ID == "search_training_accuracy" || d.ID == "search_holdout_accuracy") && d.Status != "PROGRESS" {
			t.Fatalf("measured zero matches lost their PROGRESS state: %#v", d)
		}
	}
	if err := VerifyIRBodySearchProjection(context.Background(), "zero.gooo", source, prior); err != nil {
		t.Fatalf("replay rejected a faithfully observed partial result: %v", err)
	}
}
