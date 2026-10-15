package bodycodegen

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
)

func TestSourceIRSearchRetainsOriginalLocalCallsAndReplays(t *testing.T) {
	source := searchCallsSource(`activity Step(Integer) -> Integer computes "return input + 1"`,
		"let denominator = __GOOO_BODY_HOLE_value__; return Step(input + denominator) / denominator",
		"integer-offset-constant/v1", `case "1" -> "3" case "0" -> "2" holdout_case "5" -> "7"`)
	ctx := context.Background()
	spec, err := SourceAssembly(ctx, "calls.gooo", source, "Build")
	if err != nil {
		t.Fatal(err)
	}
	r, err := GenerateWithSourceIRSearch(ctx, "calls.gooo", source, "Build", spec, "", "")
	if err != nil {
		t.Fatal(err)
	}
	history := r.Report.BodySearch.LocalCalls
	if history == nil || history.Status != "RECORDED" || len(history.Evaluations) < 3 {
		t.Fatal("actual search passes were not observed", history)
	}
	first := history.Evaluations[0]
	if first.Phase != "candidate" || first.CandidateID != r.Report.BodySearch.Attempts[0].CandidateID ||
		first.EvaluationCompleted || first.CasesStarted != 2 || first.CasesCompleted != 1 ||
		len(first.Calls) != 2 || string(first.Calls[1].Inputs) != "[0]" {
		t.Fatal("failed candidate helper exposure was lost", first)
	}
	for i, phase := range []string{"selected-training", "selected-holdout"} {
		if history.Evaluations[len(history.Evaluations)-2+i].Phase != phase {
			t.Fatal("selected suites have no explicit phase", history)
		}
	}
	if err := VerifyIRBodySearchProjection(ctx, "calls.gooo", source, r); err != nil {
		t.Fatal("existing replay pass lost call observations", err)
	}
	raw, _ := json.MarshalIndent(r, "", "  ")
	var modified Result
	if err := json.Unmarshal(raw, &modified); err != nil {
		t.Fatal(err)
	}
	if err := VerifyIRBodySearchProjection(ctx, "calls.gooo", source, modified); err != nil {
		t.Fatal("saved indented JSON must preserve exact typed evidence", err)
	}
	modified.Report.BodySearch.LocalCalls.Evaluations[0].Calls[1].Inputs = json.RawMessage("[999]")
	if VerifyIRBodySearchProjection(ctx, "calls.gooo", source, modified) == nil {
		t.Fatal("invented historical argument accepted")
	}
	modified.Report.BodySearch.LocalCalls = nil
	if err := VerifyIRBodySearchProjection(ctx, "calls.gooo", source, modified); err != nil || modified.Report.BodySearch.LocalCalls != nil {
		t.Fatal("legacy absence must stay unobserved", err)
	}
}

func TestSourceIRSearchCapturesQuadraticProbesAndExplicitCandidate(t *testing.T) {
	source := searchCallsSource(`activity Step(Integer) -> Integer computes "return input * input"`,
		"return input + Step(__GOOO_BODY_HOLE_value__)", "integer-hole-quadratic/v1",
		`case "2" -> "11" case "10" -> "19" holdout_case "100" -> "109"`)
	ctx := context.Background()
	spec, err := SourceAssembly(ctx, "calls.gooo", source, "Build")
	if err != nil {
		t.Fatal(err)
	}
	r, err := GenerateWithSourceIRSearch(ctx, "calls.gooo", source, "Build", spec, "", "")
	if err != nil {
		t.Fatal(err)
	}
	h := r.Report.BodySearch.LocalCalls
	if h.Status != "RECORDED" || len(h.Evaluations) < 8 {
		t.Fatal(h)
	}
	for i, expected := range []string{"[0]", "[1]", "[-1]", "[0]", "[1]", "[-1]"} {
		probe := h.Evaluations[i]
		if probe.Phase != "candidate-generation" || len(probe.Calls) != 1 || string(probe.Calls[0].Inputs) != expected {
			t.Fatal("original training probe missing", i, probe)
		}
	}
	if err := VerifyIRBodySearchProjection(ctx, "calls.gooo", source, r); err != nil {
		t.Fatal(err)
	}
	candidate, _, err := RealizeSourceSearchCandidate(ctx, "calls.gooo", source, "Build", r.Report.BodySearch.SelectedCandidateID)
	if err != nil || candidate.LocalCalls.Status != "RECORDED" || len(candidate.LocalCalls.Evaluations) != 7 {
		t.Fatal("caller-selected candidate lost its probe or scoring history", candidate.LocalCalls, err)
	}
	legacy := candidate
	legacy.LocalCalls = nil
	if !EqualSearchCandidateHistory([]SearchCandidate{legacy}, []SearchCandidate{candidate}) || !reflect.DeepEqual(legacy.LocalCalls, (*LocalCallHistory)(nil)) {
		t.Fatal("legacy comparison changed the saved receipt")
	}
}

func TestSearchLocalCallAggregateLimitsKeepEvaluation(t *testing.T) {
	source := []byte(pureCallsFixture)
	r, err := Generate("calls.gooo", source, "Score")
	if err != nil {
		t.Fatal(err)
	}
	for _, limit := range []string{"calls", "bytes", "evaluations"} {
		t.Run(limit, func(t *testing.T) {
			ctx, history := captureSearchCalls(context.Background(), "calls.gooo", source, "Score")
			capture := ctx.Value(localCaptureKey{}).(*searchCallCapture)
			switch limit {
			case "calls":
				capture.calls = localCallRecordLimit
			case "bytes":
				capture.bytes = localCallByteLimit
			case "evaluations":
				history.Evaluations = make([]*LocalCallObservation, 512)
			}
			_, passed, err := evaluateIntegerCasesContext(localCallPhase(ctx, "candidate", "test"), []byte(r.Source), "Score",
				[]IRBodyFillTestCase{{Input: 4, Expected: 11}})
			if err != nil || passed != 1 || history.Status != "UNKNOWN" || history.Reason == "" {
				t.Fatal("aggregate cap changed evaluation or implied complete capture", history, err)
			}
		})
	}
}
