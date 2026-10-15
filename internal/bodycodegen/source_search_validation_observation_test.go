package bodycodegen

import (
	"context"
	"errors"
	"testing"
)

func TestSourceSearchValidationCapturesOriginalResidualProbes(t *testing.T) {
	source := searchCallsSource(`activity Step(Integer) -> Integer computes "return input * 2"`,
		"return Step(input + __GOOO_BODY_HOLE_value__)", "integer-hole-residual/v1", `case "2" -> "6"`)
	r, err := ObserveSourceIRSearchValidation(context.Background(), "preflight.gooo", source, "Build")
	if err != nil || !r.Completed || r.SourceSHA256 != digest(source) || r.ActivityID != "searchcalls://activity/build" ||
		r.LocalCalls == nil || r.LocalCalls.Status != "RECORDED" || len(r.LocalCalls.Evaluations) != 2 {
		t.Fatal("preflight did not retain its original pure probes", r, err)
	}
	for i, want := range []string{"[2]", "[3]"} {
		got := r.LocalCalls.Evaluations[i]
		if got.Phase != "candidate-generation" || len(got.Calls) != 1 || string(got.Calls[0].Inputs) != want ||
			got.Calls[0].ActivityID != "searchcalls://activity/step" || !got.EvaluationCompleted {
			t.Fatal("preflight changed or invented helper inputs", got)
		}
	}
}

func TestSourceSearchValidationRetainsFailedProbeAndCancellation(t *testing.T) {
	source := searchCallsSource(`activity Step(Integer) -> Integer computes "return input + 1"`,
		"let denominator = __GOOO_BODY_HOLE_value__; return Step(input) / denominator", "integer-hole-residual/v1", `case "2" -> "3"`)
	r, err := ObserveSourceIRSearchValidation(context.Background(), "failed-probe.gooo", source, "Build")
	if err != nil || !r.Completed || len(r.LocalCalls.Evaluations) != 2 {
		t.Fatal("one failed probe prevented valid grammar validation", r, err)
	}
	failed := r.LocalCalls.Evaluations[0]
	if failed.EvaluationCompleted || failed.CasesCompleted != 0 || len(failed.Calls) != 1 || string(failed.Calls[0].Inputs) != "[2]" {
		t.Fatal("failed preflight lost its actual helper prefix", failed)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r, err = ObserveSourceIRSearchValidation(ctx, "cancelled.gooo", source, "Build")
	if !errors.Is(err, context.Canceled) || r.Completed || r.LocalCalls != nil {
		t.Fatal("cancelled validation gained unobserved calls", r, err)
	}
}
