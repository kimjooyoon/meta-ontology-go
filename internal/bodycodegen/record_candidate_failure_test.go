package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
)

func recordFailureFixture(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../examples/caller-record-failures/main.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestRecordAssemblyContinuesAfterEvaluationFailure(t *testing.T) {
	ctx := context.Background()
	source := recordFailureFixture(t)
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	result, err := g.GenerateSourceAssembly(ctx, "record.gooo", source, "Build")
	if err != nil {
		t.Fatal("one failed candidate discarded the remaining source budget", err)
	}
	r := result.Report.RecordAssembly
	if r.Status != "COMPLETE_FINITE" || r.SelectedMask != 1 || len(r.Attempts) != 2 || r.Passed != 2 || r.ModelCalls != 0 {
		t.Fatal("finite record search did not reach its valid alternative", r)
	}
	failed := r.Attempts[0]
	if failed.Status != "EVALUATION_FAILED" || failed.Total != 0 || failed.Passed != 0 ||
		!strings.Contains(failed.Reason, "zero divisor") || failed.LocalCalls.Status != "RECORDED" ||
		failed.LocalCalls.EvaluationCompleted || failed.LocalCalls.CasesCompleted != 1 || len(failed.LocalCalls.Calls) != 2 ||
		string(failed.LocalCalls.Calls[1].Inputs) != "[0]" {
		t.Fatal("failed local case gained a score or lost its actual input prefix", failed)
	}
	raw, _ := json.MarshalIndent(result, "", "  ")
	var saved Result
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if _, err := RealizeSourceAssembly(ctx, "record.gooo", source, saved); err != nil {
		t.Fatal("saved failure did not replay", err)
	}
	saved.Report.RecordAssembly.Attempts[0].LocalCalls.Calls[1].Inputs = json.RawMessage("[99]")
	if _, err := RealizeSourceAssembly(ctx, "record.gooo", source, saved); err == nil {
		t.Fatal("changed failure inputs were accepted")
	}
	for i := range saved.Report.RecordAssembly.Attempts {
		saved.Report.RecordAssembly.Attempts[i].LocalCalls = nil
	}
	populateCompletenessReceipt(&saved.Report, "")
	if _, err := RealizeSourceAssembly(ctx, "record.gooo", source, saved); err != nil {
		t.Fatal("missing historical call evidence must remain absent", err)
	}
}

func TestRecordAssemblyFailureBudgetAndExplicitRejection(t *testing.T) {
	ctx := context.Background()
	source := recordFailureFixture(t)
	g, err := NewTypedPathGenerator("")
	if err != nil {
		t.Fatal(err)
	}
	bounded := []byte(strings.Replace(string(source), `attempts "2"`, `attempts "1"`, 1))
	_, err = g.GenerateSourceAssembly(ctx, "record.gooo", bounded, "Build")
	var failed *RecordAssemblyError
	if !errors.As(err, &failed) || len(failed.Receipt.Attempts) != 1 || failed.Receipt.Attempts[0].Status != "EVALUATION_FAILED" {
		t.Fatal("failed budget or attempted prefix was lost", err)
	}
	candidate, selected, err := RealizeRecordCandidate(ctx, "record.gooo", source, "Build", 0)
	var rejected *SourceRecordCandidateRejection
	if !errors.As(err, &rejected) || selected != nil || candidate.Attempt.LocalCalls == nil || candidate.SelectedSourceSHA256 != "" {
		t.Fatal("explicit failed candidate lost its typed rejection", candidate, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err = RealizeRecordCandidate(ctx, "record.gooo", source, "Build", 0)
	if !errors.Is(err, context.Canceled) || errors.As(err, &rejected) {
		t.Fatal("cancellation became a consumed candidate", err)
	}
}
