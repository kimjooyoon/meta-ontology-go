package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func TestTypedPathBatchesPreserveLegacyBodyAndPartialResults(t *testing.T) {
	for _, language := range []string{"en", "ko"} {
		for _, partial := range []bool{false, true} {
			source, document := conditionalPathFixture(t, language)
			if partial {
				document.TestCases[6].Expected = 999
			}
			legacy, err := GenerateWithTypedPaths(context.Background(), "fixture.gooo", source,
				"ConditionalAssign", document, "")
			if err != nil {
				t.Fatal(err)
			}
			batched, err := GenerateWithTypedPathBatches(context.Background(), "fixture.gooo", source,
				"ConditionalAssign", document, "", 8)
			if err != nil {
				t.Fatal(err)
			}
			left, right := legacy.Report.BodyPaths, batched.Report.BodyPaths
			if legacy.Source != batched.Source || !reflect.DeepEqual(left.Search, right.Search) ||
				left.FunctionalCompleteness != right.FunctionalCompleteness ||
				!right.SourceBaseMatched || !batched.Report.TypecheckPassed ||
				!batched.Report.DeterministicReplay || batched.Report.RepositoryWrites != 0 {
				t.Fatal("batching changed source, finite search, or native verification")
			}
			assertPathProgress(t, right, 8, 0)
			if partial && (right.Search.Evaluated != 64 || right.FunctionalCompleteness != 600.0/7 ||
				right.Search.Status != "PARTIAL" || right.Search.Unattempted != 0) {
				t.Fatal("partial observations lost their denominator")
			}
		}
	}
}

func assertPathProgress(t *testing.T, receipt *BodyPathReceipt, step, calls int) {
	t.Helper()
	if len(receipt.Progress) < 2 || receipt.Progress[0].Attempted != 0 ||
		receipt.Progress[0].Selection.ModelCalls != calls || !receipt.Progress[0].Initialized {
		t.Fatal("initial ranking was not observed before tests")
	}
	seen := map[string]bool{}
	previous := ""
	total := 0
	for index, progress := range receipt.Progress {
		if progress.Sequence != index+1 || progress.PreviousSHA != previous || progress.SHA == "" ||
			progress.PredictionsThisAdvance != 0 || progress.Selection.ModelCalls != calls ||
			len(progress.NewAttempts) > step || progress.ScheduledBytes != 8 {
			t.Fatalf("invalid owned progress: %+v", progress)
		}
		previous = progress.SHA
		for _, attempt := range progress.NewAttempts {
			raw, err := json.Marshal(attempt.Choices)
			if err != nil || seen[string(raw)] {
				t.Fatal("candidate was repeated between batches")
			}
			seen[string(raw)] = true
			total++
		}
		if progress.Attempted != total || progress.Unattempted != progress.Declared-total {
			t.Fatal("cumulative progress does not reconcile")
		}
	}
	if total != len(receipt.Search.Attempts) {
		t.Fatal("aggregate search lost a new attempt")
	}
}

func TestTypedPathBatchesRankOnceAndKeepSourceBinding(t *testing.T) {
	source, document := typedPathFixture(t)
	model := writeTypedPathContractModel(t, false)
	result, err := GenerateWithTypedPathBatches(context.Background(), "fixture.gooo", source,
		"Combined", document, model, 1)
	if err != nil {
		t.Fatal(err)
	}
	assertPathProgress(t, result.Report.BodyPaths, 1, 3)
	// Source binding must still reject before trying to load an absent model.
	document.Plan.Base.Expressions[1].Int = 999
	_, err = GenerateWithTypedPathBatches(context.Background(), "fixture.gooo", source,
		"Combined", document, "absent-model.json", 1)
	var failure *BodyPathError
	if !errors.As(err, &failure) || failure.Receipt.SourceBaseMatched ||
		failure.Receipt.Search.Selection.ModelCalls != 0 || failure.Receipt.Timing.ModelLoadMS != 0 {
		t.Fatalf("source binding did not precede ranking: %v", err)
	}
}

func TestTypedPathBatchesRejectInvalidBudgetAndCancellation(t *testing.T) {
	for _, step := range []int{-1, 0, 65} {
		if _, err := GenerateWithTypedPathBatches(context.Background(), "fixture.gooo", nil,
			"Combined", pathplan.Document{}, "absent-model.json", step); err == nil {
			t.Fatal("invalid batch budget was accepted")
		}
	}
	source, document := typedPathFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := GenerateWithTypedPathBatches(ctx, "fixture.gooo", source,
		"Combined", document, "absent-model.json", 1)
	var failure *BodyPathError
	if !errors.Is(err, context.Canceled) || !errors.As(err, &failure) ||
		failure.Receipt.Search.Selection.ModelCalls != 0 || len(failure.Receipt.Progress) != 0 {
		t.Fatal("canceled call performed ranking or candidates")
	}
}
