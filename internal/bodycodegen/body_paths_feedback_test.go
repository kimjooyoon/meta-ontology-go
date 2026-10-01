package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func TestTypedPathFeedbackRetainsPartialBodyAndActualCalls(t *testing.T) {
	for _, language := range []string{"en", "ko"} {
		t.Run(language, func(t *testing.T) {
			source, document := conditionalPathFixture(t, language)
			document.TestCases[6].Expected = 999
			model := writeTypedPathContractModel(t, false)
			baseline, err := GenerateWithTypedPathBatches(context.Background(), "fixture.gooo", source,
				"ConditionalAssign", document, model, 8)
			if err != nil {
				t.Fatal(err)
			}
			ci := &pathplan.CIHint{SourceSHA: strings.Repeat("a", 40), Status: "FAIL"}
			result, err := GenerateWithTypedPathFeedback(context.Background(), "fixture.gooo", source,
				"ConditionalAssign", document, model, 8, 2, ci)
			if err != nil {
				t.Fatal(err)
			}
			receipt := result.Report.BodyPaths
			if result.Source != baseline.Source || result.Report.ActivityID != baseline.Report.ActivityID ||
				!result.Report.TypecheckPassed || !result.Report.DeterministicReplay || result.Report.RepositoryWrites != 0 ||
				!receipt.SourceBaseMatched || receipt.Search.Status != "PARTIAL" || receipt.Search.Evaluated != 64 ||
				receipt.FunctionalCompleteness != 600.0/7 || len(receipt.NativeCases) != 7 || len(receipt.Feedback) != 2 ||
				receipt.Search.Selection.ModelCalls != 18 || receipt.Search.Selection.ExternalCalls != 0 {
				t.Fatalf("partial native feedback lost source identity, cases or actual calls: %+v", receipt)
			}
			seen := map[string]bool{}
			progresses := map[string]pathplan.SessionProgress{}
			previous, attempted, calls := "", 0, 6
			for index, progress := range receipt.Progress {
				if progress.Sequence != index+1 || progress.PreviousSHA != previous || progress.SHA == "" ||
					progress.Selection.ModelCalls < calls || progress.PredictionsThisAdvance != 0 || len(progress.NewAttempts) > 8 {
					t.Fatal("feedback progress is not chained or cumulative")
				}
				previous, calls = progress.SHA, progress.Selection.ModelCalls
				progresses[progress.SHA] = progress
				for _, attempt := range progress.NewAttempts {
					raw, _ := json.Marshal(attempt.Choices)
					if seen[string(raw)] {
						t.Fatal("feedback repeated an already evaluated path")
					}
					seen[string(raw)] = true
					attempted++
				}
				if progress.Attempted != attempted {
					t.Fatal("progress changed the committed candidate count")
				}
			}
			if attempted != 64 || calls != 18 || len(receipt.Progress) != 11 {
				t.Fatal("feedback observations did not reconcile with the final search")
			}
			previous = ""
			for index, feedback := range receipt.Feedback {
				prior, exists := progresses[feedback.FromProgressSHA]
				if !exists || prior.Attempted != feedback.Attempted || feedback.Round != index+1 ||
					feedback.PreviousSHA != previous || feedback.SHA == "" || !feedback.Applied ||
					feedback.CI == nil || *feedback.CI != *ci || feedback.CIIsAuthority || feedback.FirstFailure == nil ||
					feedback.ModelCalls != 6 || feedback.CumulativeCalls != 6*(index+2) {
					t.Fatal("feedback failed to bind actual calls and finite failure to the prior observation")
				}
				previous = feedback.SHA
			}
			if len(baseline.Report.BodyPaths.Feedback) != 0 {
				t.Fatal("rank-once mode unexpectedly reconsidered candidates")
			}
		})
	}
}

func TestTypedPathFeedbackValidatesBeforeModelLoad(t *testing.T) {
	source, document := typedPathFixture(t)
	for _, options := range []struct {
		model        string
		step, rounds int
		ci           *pathplan.CIHint
	}{
		{"", 1, 1, nil}, {"absent.json", 0, 1, nil}, {"absent.json", 65, 1, nil},
		{"absent.json", 1, 0, nil}, {"absent.json", 1, 17, nil},
		{"absent.json", 1, 1, &pathplan.CIHint{SourceSHA: "invalid", Status: "PASS"}},
	} {
		_, err := GenerateWithTypedPathFeedback(context.Background(), "fixture.gooo", source, "Combined",
			document, options.model, options.step, options.rounds, options.ci)
		if err == nil || strings.Contains(err.Error(), "load explicit") {
			t.Fatal("invalid feedback options reached model loading")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := GenerateWithTypedPathFeedback(ctx, "fixture.gooo", source, "Combined", document, "absent.json", 1, 1, nil)
	var failure *BodyPathError
	if !errors.Is(err, context.Canceled) || !errors.As(err, &failure) ||
		failure.Receipt.Timing.ModelLoadMS != 0 || failure.Receipt.Search.Selection.ModelCalls != 0 {
		t.Fatal("canceled feedback performed model work")
	}
	document.Plan.Base.Expressions[1].Int = 999
	_, err = GenerateWithTypedPathFeedback(context.Background(), "fixture.gooo", source, "Combined",
		document, "absent.json", 1, 1, nil)
	if !errors.As(err, &failure) || failure.Receipt.SourceBaseMatched ||
		failure.Receipt.Timing.ModelLoadMS != 0 || failure.Receipt.Search.Selection.ModelCalls != 0 {
		t.Fatal("feedback bypassed authoritative source binding")
	}
}

func TestTypedPathFeedbackContextDeclineRetainsNativePartialBody(t *testing.T) {
	for _, language := range []string{"en", "ko"} {
		source, document := conditionalPathFixture(t, language)
		document.TestCases[6].Expected = 999
		for len(document.Plan.Decisions[0].Intent)+len(" detail") <= 480 {
			document.Plan.Decisions[0].Intent += " detail"
		}
		model := writeTypedPathContractModel(t, false)
		baseline, err := GenerateWithTypedPathBatches(context.Background(), "fixture.gooo", source,
			"ConditionalAssign", document, model, 8)
		if err != nil {
			t.Fatal(err)
		}
		result, err := GenerateWithTypedPathFeedback(context.Background(), "fixture.gooo", source,
			"ConditionalAssign", document, model, 8, 2, nil)
		if err != nil {
			t.Fatal("valid partial body blocked by optional context format", err)
		}
		receipt := result.Report.BodyPaths
		if result.Source != baseline.Source || result.Report.ActivityID != baseline.Report.ActivityID ||
			!result.Report.TypecheckPassed || !result.Report.DeterministicReplay || result.Report.RepositoryWrites != 0 ||
			receipt.FunctionalCompleteness != 600.0/7 || receipt.Search.Status != "PARTIAL" || receipt.Search.Evaluated != 64 ||
			receipt.Search.Selection.ModelCalls != 6 || len(receipt.Feedback) != 2 {
			t.Fatal("native context decline lost finite parity or stable body identity")
		}
		for _, decline := range receipt.Feedback {
			if !decline.ContextDeclined || decline.ModelCalls != 0 || decline.Applied || decline.SHA == "" || decline.DeclinedBytes <= 512 {
				t.Fatal("native receipt hid zero-call context decline")
			}
		}
		last := receipt.Progress[len(receipt.Progress)-1]
		if last.Interrupted || last.FeedbackRounds != 2 || last.FeedbackPredictions != 0 || last.Attempted != 64 {
			t.Fatal("representation decline interrupted native continuation")
		}
	}
}
