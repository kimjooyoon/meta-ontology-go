package bodycodegen

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestRetainedSearchCandidatesBindOriginalPoolAndCurrentExecution(t *testing.T) {
	source := searchCallsSource(`activity Step(Integer) -> Integer computes "return input * 2"`,
		"return Step(input + __GOOO_BODY_HOLE_value__)", "integer-hole-residual/v1", `case "2" -> "6"`)
	ctx := context.Background()
	spec, err := SourceAssembly(ctx, "retained.gooo", source, "Build")
	if err != nil {
		t.Fatal(err)
	}
	r, err := GenerateWithSourceIRSearch(ctx, "retained.gooo", source, "Build", spec, "", "")
	if err != nil || r.Report.BodySearch == nil || r.Report.BodySearch.CandidateSet == nil {
		t.Fatal("initial candidate pool missing", err)
	}
	if err := VerifyIRBodySearchProjection(ctx, "retained.gooo", source, r); err != nil {
		t.Fatal("initial pool was not verified against its source", err)
	}
	search := r.Report.BodySearch
	current := []byte(strings.Replace(string(source), "input * 2", "input * 3", 1))
	replanned, err := PlanSourceSearchCandidates(ctx, "retained.gooo", current, "Build")
	if err != nil || replanned.PlanSHA256 == search.IRPlanSHA256 {
		t.Fatal("fixture did not distinguish helper-dependent pools", err)
	}
	retained, err := RetainedSourceSearchCandidates(ctx, "retained.gooo", current, "Build", search)
	if err != nil || retained.PlanSHA256 != search.CandidateSet.PlanSHA256 ||
		!equalLocalCallEvidence(retained.Candidates, search.CandidateSet.Candidates) ||
		!equalLocalCallEvidence(retained.Generation, search.CandidateGeneration) {
		t.Fatal("current helper changed the retained pool", err)
	}
	retained.Candidates[0].Expression = "99"
	if search.CandidateSet.Candidates[0].Expression == "99" {
		t.Fatal("returned slice aliases the initial pool")
	}
	observedContext, stop := context.WithCancel(ctx)
	defer stop()
	observed := &fillEvaluationCancelContext{Context: observedContext, cancel: stop, at: 1}
	if _, err := RetainedSourceSearchCandidates(observed, "retained.gooo", current, "Build", search); err != nil || observed.evaluations != 0 {
		t.Fatal("reading the saved candidate set executed a new evaluator pass", err, observed.evaluations)
	}
	got, _, err := RealizeRetainedSourceSearchCandidate(ctx, "retained.gooo", current, "Build", search, search.SelectedCandidateID)
	if err != nil || got.PlanningSourceSHA256 != digest(source) || got.InputSourceSHA256 != digest(current) ||
		!got.Attempt.ScoringCompleted || got.Attempt.TestCasesPassed != 0 || got.Attempt.CaseResults[0].Actual != 9 ||
		got.LocalCalls == nil || len(got.LocalCalls.Evaluations) != 1 {
		t.Fatal("current helper was not evaluated independently of the original score", got, err)
	}
	changedContract := []byte(strings.Replace(string(current), `case "2" -> "6"`, `case "2" -> "7"`, 1))
	if _, _, err := RealizeRetainedSourceSearchCandidate(ctx, "retained.gooo", changedContract, "Build", search, search.SelectedCandidateID); err == nil {
		t.Fatal("changed source cases accepted the old pool")
	}
	if _, _, err := RealizeRetainedSourceSearchCandidate(ctx, "retained.gooo", current, "Build", search, "missing"); err == nil {
		t.Fatal("candidate outside the retained prefix was accepted")
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := RetainedSourceSearchCandidates(canceled, "retained.gooo", current, "Build", search); err == nil {
		t.Fatal("canceled planning succeeded")
	}
	raw, _ := json.MarshalIndent(r, "", "  ")
	var saved Result
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if err := VerifyIRBodySearchProjection(ctx, "retained.gooo", source, saved); err != nil {
		t.Fatal("saved pool did not replay", err)
	}
	saved.Report.BodySearch.CandidateSet.Candidates[0].Expression = "99"
	if err := VerifyIRBodySearchProjection(ctx, "retained.gooo", source, saved); err == nil {
		t.Fatal("modified pool was not rederived from the original source")
	}
	saved.Report.BodySearch.CandidateSet = nil
	if err := VerifyIRBodySearchProjection(ctx, "retained.gooo", source, saved); err != nil || saved.Report.BodySearch.CandidateSet != nil {
		t.Fatal("historical missing pool became a newly retained one", err)
	}
}
