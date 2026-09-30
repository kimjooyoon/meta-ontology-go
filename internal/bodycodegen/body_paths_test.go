package bodycodegen

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func typedPathFixture(t *testing.T) ([]byte, pathplan.Document) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/typed-path-compound.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/body-codegen/typed-path-compound-plan.json")
	if err != nil {
		t.Fatal(err)
	}
	document, err := pathplan.DecodeDocument(raw)
	if err != nil {
		t.Fatal(err)
	}
	return source, document
}

func TestTypedPathsCompoundTDDPreservesOwnerAndReplay(t *testing.T) {
	source, document := typedPathFixture(t)
	original := append([]byte(nil), source...)
	result, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", document, "")
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyPaths
	if receipt == nil || !receipt.SourceBaseMatched || receipt.Search.Status != "TRAINING_COMPLETE" ||
		receipt.Search.Selection.ModelCalls != 0 || receipt.Search.Selection.ExternalCalls != 0 ||
		!receipt.Search.Selection.ExternalCallsKnown || receipt.Search.DeclaredCombinations != 8 ||
		receipt.FunctionalCompleteness != 100 || len(receipt.NativeCases) != 3 ||
		result.Report.ActivityID != "sample://activity/combined" || !result.Report.TypecheckPassed ||
		!result.Report.DeterministicReplay || receipt.OriginalSourceSHA256 == receipt.SelectedSourceSHA256 {
		t.Fatalf("compound receipt: %+v", result.Report)
	}
	if string(source) != string(original) || result.Report.RepositoryWrites != 0 {
		t.Fatal("source or repository was mutated")
	}
	replayed, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", document, "")
	if err != nil || replayed.Source != result.Source ||
		replayed.Report.BodyPaths.Search.Selection.Choices["reference"] != "reference_second" ||
		replayed.Report.BodyPaths.Search.Selection.Choices["branch"] != "layout_reverse" {
		t.Fatal("offline TDD was not deterministic")
	}
	// The same source-span replacement helper must preserve unrelated declarations.
	if !strings.Contains(string(source), `activity Unrelated(Integer) -> Integer computes "return input"`) {
		t.Fatal("fixture lost unrelated identity")
	}
}

func TestTypedPathsPartialFunctionalCompletenessIsSeparateFromLowering(t *testing.T) {
	source, document := typedPathFixture(t)
	document.TestCases[2].Expected = 999
	result, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", document, "")
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyPaths
	if receipt.Search.Status != "PARTIAL" || receipt.Search.TypeRejected != 2 ||
		receipt.Search.Evaluated != 6 || receipt.Search.Unattempted != 0 ||
		receipt.FunctionalCompleteness != 200.0/3 || result.Report.CompletenessPercent != 100 {
		t.Fatalf("partial receipt: %+v", receipt)
	}
}

func TestTypedPathsRejectSourceMismatchBeforeLoadingModel(t *testing.T) {
	for _, change := range []string{"body", "name", "effect", "no-source", "seed"} {
		t.Run(change, func(t *testing.T) {
			source, document := typedPathFixture(t)
			switch change {
			case "body":
				source = []byte(strings.Replace(string(source), "input + 2", "input + 99", 1))
			case "name":
				document.Plan.Base.Name = "Other"
			case "effect":
				source = []byte(strings.Replace(string(source), "input + 2", "helper(input)", 1))
			case "no-source":
				source = nil
			case "seed":
				document.Seed = "seed"
			}
			modelPath := "missing-model.json"
			if change == "seed" {
				modelPath = ""
			}
			_, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source, "Combined", document, modelPath)
			var failure *BodyPathError
			if !errors.As(err, &failure) || strings.Contains(err.Error(), "load explicit") ||
				failure.Receipt.Search.Selection.ModelCalls != 0 || failure.Receipt.Timing.ModelLoadMS != 0 {
				t.Fatalf("pre-model rejection: %v", err)
			}
		})
	}
}

func TestTypedPathsCanceledAndMissingContexts(t *testing.T) {
	source, document := typedPathFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, candidate := range []context.Context{nil, ctx} {
		_, err := GenerateWithTypedPaths(candidate, "fixture.gooo", source, "Combined", document, "")
		var failure *BodyPathError
		if !errors.As(err, &failure) || failure.Receipt.Search.Selection.ModelCalls != 0 {
			t.Fatalf("context was not rejected: %v", err)
		}
	}
}
