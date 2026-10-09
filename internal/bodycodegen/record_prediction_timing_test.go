package bodycodegen

import (
	"context"
	"encoding/json"
	"os"
	"testing"
)

func recordOwnPredictionFixture(t *testing.T) ([]byte, Result) {
	t.Helper()
	source, err := os.ReadFile("../../examples/scalar-identity/source.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	g, err := NewTypedPathGenerator("../../examples/scalar-identity/model/model.json")
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.GenerateSourceAssembly(context.Background(), "r.gooo", source, "Describe")
	if err != nil {
		t.Fatal(err)
	}
	if r.Report.RecordAssembly.ModelCalls != 1 || r.Report.RecordAssembly.Prediction == nil {
		t.Fatal("real prediction missing")
	}
	return source, r
}

func TestRecordPredictionReplayAcceptsZeroResolutionTiming(t *testing.T) {
	source, r := recordOwnPredictionFixture(t)
	r.Report.RecordAssembly.PredictNS = 0
	populateCompletenessReceipt(&r.Report, "")
	replayed, err := RealizeSourceAssembly(context.Background(), "r.gooo", source, r)
	if err != nil {
		t.Fatal("zero-resolution timing rejected despite actual prediction:", err)
	}
	if replayed.ModelCalls != 0 || replayed.FinitePassed != 4 || r.Report.RecordAssembly.PredictNS != 0 {
		t.Fatal("replay changed measured timing or made a new prediction")
	}
}

func TestRecordPredictionTimingDoesNotReplaceInferenceEvidence(t *testing.T) {
	source, original := recordOwnPredictionFixture(t)
	for _, mutate := range []func(*RecordAssemblyReceipt){
		func(r *RecordAssemblyReceipt) { r.PredictNS = -1 },
		func(r *RecordAssemblyReceipt) { r.PredictNS = 0; r.ModelCalls = 0 },
		func(r *RecordAssemblyReceipt) { r.PredictNS = 0; r.Prediction = nil },
		func(r *RecordAssemblyReceipt) { r.PredictNS = 0; r.Ranking[0] = 7 },
		func(r *RecordAssemblyReceipt) { r.PredictNS = 0; r.Prediction.Probabilities[0] = 2 },
	} {
		raw, err := json.Marshal(original)
		if err != nil {
			t.Fatal(err)
		}
		var r Result
		if err = json.Unmarshal(raw, &r); err != nil {
			t.Fatal(err)
		}
		mutate(r.Report.RecordAssembly)
		populateCompletenessReceipt(&r.Report, "")
		if _, err = RealizeSourceAssembly(context.Background(), "r.gooo", source, r); err == nil {
			t.Fatal("invalid inference evidence accepted with zero timing")
		}
	}
}
