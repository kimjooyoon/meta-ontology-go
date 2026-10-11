package bodycodegen

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Cancel at an evaluator boundary without timers or a goroutine race. The
// evaluator's existing context lookup marks entry into each actual pass.
type fillEvaluationCancelContext struct {
	context.Context
	cancel          context.CancelFunc
	evaluations, at int
}

func (c *fillEvaluationCancelContext) Value(key any) any {
	if _, ok := key.(localCaptureKey); ok {
		c.evaluations++
		if c.evaluations == c.at {
			c.cancel()
		}
	}
	return c.Context.Value(key)
}

func TestBodyFillProbesRespectCallerContext(t *testing.T) {
	for _, vector := range []bool{false, true} {
		name, source := "scalar", "package p; func Pick(input int64) int64 { return input }"
		training := []IRBodyFillTestCase{{Input: 0}}
		if vector {
			name, source = "vector", "package p; func Pick(input0, input1 int64) int64 { return input0 + input1 }"
			training[0].Inputs = []int64{0, 0}
		}
		t.Run(name, func(t *testing.T) {
			sources := map[string][]byte{"first": []byte(source), "second": []byte(source)}
			if r, err := measureIRBodyFillBehavioralProbes(nil, "Pick", sources, training, nil); err == nil || r != nil {
				t.Fatal("missing context produced a probe score", r, err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			observed := &fillEvaluationCancelContext{Context: ctx, cancel: cancel, at: 2}
			r, err := measureIRBodyFillBehavioralProbes(observed, "Pick", sources, training, nil)
			if !errors.Is(err, context.Canceled) || r != nil || observed.evaluations != 2 {
				t.Fatal("mid-probe cancellation became a failed candidate score", r, err, observed.evaluations)
			}
			if r, err = measureIRBodyFillBehavioralProbes(ctx, "Pick", sources, nil, nil); !errors.Is(err, context.Canceled) || r != nil {
				t.Fatal("empty probe set hid caller cancellation", r, err)
			}
		})
	}
}

func TestBodyFillSelectedEvaluationsKeepCallerContext(t *testing.T) {
	source := tinyBodyFillFixture("return __GOOO_BODY_HOLE_value__")
	plan := tinyBodyFillPlan("Add one.", "value",
		IRBodyFillCandidate{ID: "identity", Expression: "input"},
		IRBodyFillCandidate{ID: "add", Expression: "input + 1"})
	plan.TestCases = []IRBodyFillTestCase{{Input: 2, Expected: 3}}
	plan.HoldoutTestCases = []IRBodyFillTestCase{{Input: 5, Expected: 6}}
	// Two candidate training passes, two behavioral passes, then the selected
	// training and holdout passes. Cancellation must prevent a completed result.
	for _, stage := range []struct {
		name string
		at   int
	}{{"selected-training", 5}, {"selected-holdout", 6}} {
		t.Run(stage.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			observed := &fillEvaluationCancelContext{Context: ctx, cancel: cancel, at: stage.at}
			r, err := GenerateWithIRBodyFill(observed, "cancel.gooo", source, "Choose", plan, "", "")
			if !errors.Is(err, context.Canceled) || r.Source != "" || r.Report.BodyFill != nil || observed.evaluations != stage.at {
				t.Fatal("selected evaluation ignored cancellation", err, observed.evaluations, r.Report.BodyFill)
			}
		})
	}
}

func TestSourceRecordFillReturnsCallerCancellationAfterProvider(t *testing.T) {
	source, activity := readRecordBodyFillFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		cancel()
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	r, err := GenerateWithSourceIRBodyFill(ctx, "record-cancel.gooo", source, activity.Name,
		&activity.Assembly.Spec, server.URL+"/v1/systemone", "", IRBodyFillOptions{})
	if !errors.Is(err, context.Canceled) || r.Source != "" || r.Report.BodyFill != nil {
		t.Fatal("record cancellation became deterministic completion", err, r.Report.BodyFill)
	}
}

func TestBodyFillProbeEvaluationFailureRemainsAnObservation(t *testing.T) {
	sources := map[string][]byte{
		"good": []byte("package p; func Pick(input int64) int64 { return input }"),
		"bad":  []byte("package p; func Pick(input int64) int64 { return 1 / input }"),
	}
	r, err := measureIRBodyFillBehavioralProbes(context.Background(), "Pick", sources,
		[]IRBodyFillTestCase{{Input: 1}}, nil)
	if err != nil || r == nil || r.CandidateRunsCompleted != 1 || r.CandidateRunsFailed != 1 {
		t.Fatal("ordinary probe failure became cancellation", r, err)
	}
}
