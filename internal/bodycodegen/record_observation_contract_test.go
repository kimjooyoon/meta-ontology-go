package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestRecordObservationUnansweredAndBoundedQuestions(t *testing.T) {
	for _, tc := range []struct {
		name, from, to, status string
		cases, attempts        int
	}{
		{"no oracle", `oracle "Expected"`, "", "ORACLE_UNAVAILABLE", 1, 1},
		{"candidate cap", `max_candidates "2"`, `max_candidates "1"`, "CANDIDATE_BUDGET", 1, 1},
		{"no resolving", `resolve_unique "true"`, `resolve_unique "false"`, "ONE_SURVIVING_CANDIDATE", 2, 2},
		{"without cache", `reuse_outputs "true"`, `reuse_outputs "false"`, "ONE_SURVIVING_CANDIDATE", 2, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := []byte(strings.Replace(sourceRecordObservationFixture, tc.from, tc.to, 1))
			g, _ := NewTypedPathGenerator("")
			result, err := g.GenerateSourceAssembly(context.Background(), "questions.gooo", source, "Decide")
			if err != nil {
				t.Fatal(err)
			}
			r := result.Report.RecordAssembly
			if r.Observation.Status != tc.status || r.Total != tc.cases || len(r.Attempts) != tc.attempts || r.ModelCalls != 0 {
				t.Fatal(r)
			}
			if tc.name == "without cache" && r.Observation.ReusedOutputs != 0 {
				t.Fatal("disabled cache used", r.Observation)
			}
			if _, err = RealizeSourceAssembly(context.Background(), "questions.gooo", source, result); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRecordObservationContradictionAndTypedProbeErrors(t *testing.T) {
	wrong := strings.Replace(sourceRecordObservationFixture, `result.amount = 0 - input0.amount; result.approved = true`, `result.amount = 1 - input0.amount; result.approved = true`, 1)
	g, _ := NewTypedPathGenerator("")
	_, err := g.GenerateSourceAssembly(context.Background(), "contradiction.gooo", []byte(wrong), "Decide")
	var failure *RecordAssemblyError
	if !errors.As(err, &failure) || failure.Receipt.Observation.Status != "FAILED" || len(failure.Receipt.Attempts) != 0 ||
		failure.Receipt.Observation.OracleInitialCases[0].Passed {
		t.Fatal("oracle contradiction was not preserved", err)
	}
	bad := strings.Replace(sourceRecordObservationFixture, `17,\"approved\":false,\"label\":\"확인\"},true]`, `17,\"approved\":false,\"label\":\"확인\"},7]`, 1)
	if bad == sourceRecordObservationFixture {
		t.Fatal("bad fixture was not changed")
	}
	if err := ValidateSourceAssembly(context.Background(), "bad.gooo", []byte(bad), "Decide"); err == nil {
		t.Fatal("incorrect Boolean question accepted")
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := g.GenerateSourceAssembly(canceled, "cancel.gooo", []byte(sourceRecordObservationFixture), "Decide"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestRecordObservationSavedReplayRejectsForgedQuestionsAndAnswers(t *testing.T) {
	ctx := context.Background()
	source := []byte(sourceRecordObservationFixture)
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(ctx, "saved.gooo", source, "Decide")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly.Observation
	if len(r.Rounds) != 2 || r.Rounds[0].RecommendedIndex == nil || *r.Rounds[0].RecommendedIndex != 0 ||
		r.Rounds[0].Probes[0].SeparatedPairs != 1 || r.Rounds[0].Probes[1].SeparatedPairs != 0 ||
		r.OracleEvaluations != 2 || r.ReusedOutputs < 1 || r.ResolvedMask == nil || *r.ResolvedMask != 1 {
		t.Fatal(r)
	}
	raw, _ := json.MarshalIndent(result, "", "  ")
	var saved Result
	if err = json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if _, err = RealizeSourceAssembly(ctx, "saved.gooo", source, saved); err != nil {
		t.Fatal("formatted saved observation", err)
	}
	for _, change := range []func(*RecordAssemblyReceipt){
		func(x *RecordAssemblyReceipt) { x.Observation = nil },
		func(x *RecordAssemblyReceipt) { x.Observation.Rounds[0].Observation.Expected = `{\"amount\":17}` },
		func(x *RecordAssemblyReceipt) { x.Observation.Options.MaxCandidates = 1 },
		func(x *RecordAssemblyReceipt) { x.Observation.ResolvedMask = new(uint16(0)) },
		func(x *RecordAssemblyReceipt) { x.Observation.CandidateEvaluations++ },
		func(x *RecordAssemblyReceipt) { x.Observation.SourceSHA256 = strings.Repeat("0", 64) },
	} {
		if err = json.Unmarshal(raw, &saved); err != nil {
			t.Fatal(err)
		}
		change(saved.Report.RecordAssembly)
		if _, err = RealizeSourceAssembly(ctx, "saved.gooo", source, saved); err == nil {
			t.Fatal("changed observation replayed")
		}
	}
}

func TestRecordObservationRetainedByCallerCandidates(t *testing.T) {
	ctx := context.Background()
	source := []byte(sourceRecordObservationFixture)
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(ctx, "caller.gooo", source, "Decide")
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err = RealizeRecordCandidate(ctx, "caller.gooo", source, "Decide", 0); err == nil {
		t.Fatal("legacy candidate discarded observation")
	}
	prior := result.Report.RecordAssembly
	before, _ := json.Marshal(prior)
	for mask := range 2 {
		candidate, _, err := RealizeObservedRecordCandidate(ctx, "caller.gooo", source, "Decide", prior, uint16(mask))
		if err != nil {
			t.Fatal(err)
		}
		if candidate.Schema != "gooo/record-candidate/v2" || candidate.Attempt.Total != 2 || candidate.Attempt.Passed != mask+1 ||
			candidate.Observation.AddedCases != 1 || candidate.Observation.EffectiveCasesSHA256 != prior.Observation.EffectiveCasesSHA256 {
			t.Fatal(candidate)
		}
	}
	changed := []byte(strings.Replace(sourceRecordObservationFixture, `result.amount = 0 - input0.amount; result.approved = true`, `result.amount = input0.amount; result.approved = true`, 1))
	if _, _, err = RealizeObservedRecordCandidate(ctx, "caller.gooo", changed, "Decide", prior, 0); err == nil {
		t.Fatal("caller oracle relabeled a retained answer")
	}
	after, _ := json.Marshal(prior)
	if string(before) != string(after) {
		t.Fatal("retained original receipt changed")
	}
}

func TestRecordObservationResumeRetainsAnswersAndChecksChangedDependencies(t *testing.T) {
	text := strings.Replace(sourceRecordObservationFixture, `let result = input0; if input1`, `let result = Copy(input0); if input1`, 1)
	// Affect an unconditionally retained field when the fixed helper changes.
	text += `activity Copy(Order) -> Order computes "return input"` + "\n"
	ctx := context.Background()
	source := []byte(text)
	g, _ := NewTypedPathGenerator("")
	prior, err := g.GenerateSourceAssembly(ctx, "resume.gooo", source, "Decide")
	if err != nil {
		t.Fatal(err)
	}
	changed := []byte(strings.Replace(text, `activity Copy(Order) -> Order computes "return input"`,
		`activity Copy(Order) -> Order computes "let result = input; result.label = input.label + \"!\"; return result"`, 1))
	result, err := ResumeRecordAssembly(ctx, "resume.gooo", source, changed, prior, assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if !reflect.DeepEqual(r.Observation, prior.Report.RecordAssembly.Observation) || r.Total != 2 || r.Passed != 0 ||
		r.Continuation.RecheckedAttempts != 1 || r.Continuation.AddedAttempts != 1 || r.ModelCalls != 0 {
		t.Fatal(r)
	}
	if _, err = RealizeSourceAssembly(ctx, "resume.gooo", changed, result); err != nil {
		t.Fatal(err)
	}
	wrong := []byte(strings.Replace(text, `result.amount = 0 - input0.amount; result.approved = true`, `result.amount = input0.amount; result.approved = true`, 1))
	if _, err = ResumeRecordAssembly(ctx, "resume.gooo", source, wrong, prior, assemblyPolicyFixture(t)); err == nil {
		t.Fatal("continuation oracle relabeled acquired answer")
	}
}
