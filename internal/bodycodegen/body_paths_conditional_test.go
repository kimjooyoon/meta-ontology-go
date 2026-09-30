package bodycodegen

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func conditionalPathFixture(t *testing.T, language string) ([]byte, pathplan.Document) {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/typed-path-conditional-assignment.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	name := "../../examples/body-codegen/typed-path-conditional-assignment-plan.json"
	if language == "ko" {
		name = "../../examples/body-codegen/typed-path-conditional-assignment-ko-plan.json"
	}
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	document, err := pathplan.DecodeDocument(raw)
	if err != nil {
		t.Fatal(err)
	}
	return source, document
}

func TestPreparedTypedPathsBooleanConditionAndAssignments(t *testing.T) {
	for _, language := range []string{"en", "ko"} {
		t.Run(language, func(t *testing.T) {
			source, document := conditionalPathFixture(t, language)
			before := string(source)
			result, err := GenerateWithTypedPaths(context.Background(), "conditional.gooo", source, "ConditionalAssign", document, "")
			if err != nil {
				t.Fatal(err)
			}
			receipt := result.Report.BodyPaths
			if receipt == nil || !receipt.SourceBaseMatched || receipt.FunctionalCompleteness != 100 ||
				receipt.Search.DeclaredCombinations != 64 || receipt.Search.TrainingTotal != 7 ||
				receipt.Search.Selection.ModelCalls != 0 || receipt.Search.Selection.ExternalCalls != 0 ||
				receipt.Timing.PlanPrepareMS <= 0 || !result.Report.TypecheckPassed || !result.Report.DeterministicReplay ||
				result.Report.RepositoryWrites != 0 || string(source) != before {
				t.Fatalf("conditional snapshot receipt: %+v", receipt)
			}
			if !strings.Contains(result.Source, "&&") || !strings.Contains(result.Source, "var guard_alt") {
				t.Fatal("Boolean locals or conjunction were not emitted")
			}
		})
	}
}

func TestPreparedTypedPathsRetainPartialBooleanContract(t *testing.T) {
	source, document := conditionalPathFixture(t, "en")
	document.TestCases[6].Expected = 999
	result, err := GenerateWithTypedPaths(context.Background(), "conditional.gooo", source, "ConditionalAssign", document, "")
	if err != nil {
		t.Fatal(err)
	}
	receipt := result.Report.BodyPaths
	if receipt.Search.Status != "PARTIAL" || receipt.Search.Evaluated != 64 || receipt.Search.TypeRejected != 0 ||
		receipt.Search.Unattempted != 0 || receipt.FunctionalCompleteness != 600.0/7 || result.Report.CompletenessPercent != 100 {
		t.Fatalf("partial Boolean contract lost its denominator: %+v", receipt)
	}
}
