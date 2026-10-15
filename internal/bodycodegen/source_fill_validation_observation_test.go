package bodycodegen

import (
	"context"
	"errors"
	"strings"
	"testing"
)

const fillValidationSource = `package fillvalidation
namespace fillvalidation
entity Integer id "fillvalidation://integer"
activity Step(Integer) -> Integer computes "return 12 / input"
activity Build(Integer) -> Integer computes "return Step(__GOOO_BODY_HOLE_value__) + __GOOO_BODY_HOLE_bias__" assembling {
 source_fill intent "Choose a working divisor." {
  hole "value"
  hole "bias"
  candidate "fails" { fill "value" "input - 2" fill "bias" "0" }
  candidate "works" { fill "value" "input" fill "bias" "0" }
 }
 case "4" -> "3"
 case "2" -> "6"
 holdout_case "3" -> "4"
}
`

func TestSourceFillValidationRetainsFailedAndCompletedPasses(t *testing.T) {
	source := []byte(fillValidationSource)
	r, err := ObserveSourceIRBodyFillValidation(context.Background(), "fill.gooo", source, "Build")
	if err != nil || !r.Completed || r.Error != "" || r.SourceSHA256 != digest(source) || r.ActivityID == "" ||
		r.LocalCalls == nil || len(r.LocalCalls.Evaluations) != 5 {
		t.Fatal("fill preflight did not retain original evaluation passes", r, err)
	}
	failed := r.LocalCalls.Evaluations[0]
	if failed.EvaluationCompleted || len(failed.Calls) != 2 || string(failed.Calls[1].Inputs) != "[0]" ||
		failed.CasesCompleted != 1 || failed.CandidateID != "fails" {
		t.Fatal("failed prefix changed", failed)
	}
	last := r.LocalCalls.Evaluations[4]
	if !last.EvaluationCompleted || last.Phase != "fill-selected-holdout" || string(last.Calls[0].Inputs) != "[3]" {
		t.Fatal("selected holdout pass was lost", last)
	}
}

func TestSourceFillValidationRetainsAllRejectedAndCanceledPrefixes(t *testing.T) {
	rejectedSource := []byte(strings.Replace(fillValidationSource, `fill "value" "input"`, `fill "value" "0"`, 1))
	r, err := ObserveSourceIRBodyFillValidation(context.Background(), "rejected.gooo", rejectedSource, "Build")
	var rejected *NoValidBodyFillCandidates
	if !errors.As(err, &rejected) || r.Completed || r.Error == "" || r.LocalCalls == nil ||
		len(r.LocalCalls.Evaluations) != 2 || len(r.LocalCalls.Evaluations[0].Calls) != 2 {
		t.Fatal("all-rejected preflight lost its entered prefix", r, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	observed := &fillEvaluationCancelContext{Context: ctx, cancel: cancel, at: 2}
	r, err = ObserveSourceIRBodyFillValidation(observed, "cancel.gooo", []byte(fillValidationSource), "Build")
	if !errors.Is(err, context.Canceled) || r.Completed || r.LocalCalls == nil ||
		len(r.LocalCalls.Evaluations) != 2 || len(r.LocalCalls.Evaluations[0].Calls) != 2 ||
		r.LocalCalls.Evaluations[1].EvaluationCompleted {
		t.Fatal("canceled preflight lost its original prefix or gained completion", r, err)
	}
}

func TestSourceFillValidationRecordCallsKeepExactValues(t *testing.T) {
	source := []byte(`package recordvalidation
namespace recordvalidation
entity Box id "recordvalidation://box" fields {
 field amount id "recordvalidation://amount" type integer required one
 field note id "recordvalidation://note" type string optional one
}
activity Step(Box) -> Box computes "return input"
activity Build(Box) -> Box computes "if __GOOO_BODY_HOLE_keep__ { return __GOOO_BODY_HOLE_value__ }; return input" assembling {
 source_fill intent "Preserve the record." {
  hole "value"
  hole "keep"
  candidate "zero" { fill "value" "Box{amount:0}" fill "keep" "true" }
  candidate "helper" { fill "value" "Step(input)" fill "keep" "true" }
 }
 value_case "[{\"amount\":9007199254740993}]" -> "{\"amount\":9007199254740993}"
 holdout_value_case "[{\"amount\":5,\"note\":\"\"}]" -> "{\"amount\":5,\"note\":\"\"}"
}
`)
	r, err := ObserveSourceIRBodyFillValidation(context.Background(), "record.gooo", source, "Build")
	if err != nil || !r.Completed || r.LocalCalls == nil || len(r.LocalCalls.Evaluations) != 4 {
		t.Fatal("record preflight lost its passes", r, err)
	}
	if string(r.LocalCalls.Evaluations[1].Calls[0].Inputs) != `[{"amount":9007199254740993,"note":null}]` ||
		string(r.LocalCalls.Evaluations[3].Calls[0].Inputs) != `[{"amount":5,"note":""}]` {
		t.Fatal("record preflight changed integer precision or optional field presence")
	}
}

func TestSourceFillValidationBeforeEvaluatorDoesNotInventCalls(t *testing.T) {
	for name, ctx := range map[string]context.Context{"missing-context": nil, "wrong-activity": context.Background()} {
		t.Run(name, func(t *testing.T) {
			r, err := ObserveSourceIRBodyFillValidation(ctx, "fill.gooo", []byte(fillValidationSource), "Step")
			if err == nil || r.Completed || r.Error == "" || r.LocalCalls != nil {
				t.Fatal("unentered validation gained evaluation evidence", r, err)
			}
		})
	}
}

func TestSourceFillValidationCompletionIsSeparateFromScores(t *testing.T) {
	source := strings.Replace(fillValidationSource, `case "4" -> "3"`, `case "4" -> "99"`, 1)
	source = strings.Replace(source, `case "2" -> "6"`, `case "2" -> "99"`, 1)
	r, err := ObserveSourceIRBodyFillValidation(context.Background(), "mismatch.gooo", []byte(source), "Build")
	if err != nil || !r.Completed || r.LocalCalls == nil {
		t.Fatal("a finite score mismatch changed the existing validation contract", r, err)
	}
	source = strings.Replace(fillValidationSource, `holdout_case "3" -> "4"`, `holdout_case "0" -> "0"`, 1)
	r, err = ObserveSourceIRBodyFillValidation(context.Background(), "holdout.gooo", []byte(source), "Build")
	if err == nil || r.Completed || r.LocalCalls == nil || len(r.LocalCalls.Evaluations) != 5 {
		t.Fatal("selected holdout error lost the original preflight history", r, err)
	}
	last := r.LocalCalls.Evaluations[4]
	if last.EvaluationCompleted || last.Phase != "fill-selected-holdout" ||
		len(last.Calls) != 1 || string(last.Calls[0].Inputs) != "[0]" {
		t.Fatal("failed selected holdout lost its entered call", last)
	}
}
