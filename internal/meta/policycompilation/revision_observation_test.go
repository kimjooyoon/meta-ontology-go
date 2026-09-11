package policycompilation

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func revisionObservationFixture(t *testing.T) ([]byte, PolicyRevisionObservationRequest) {
	t.Helper()
	source := declaredCaseSourceFixture(t)
	revision := PolicyDecisionRevision{
		ExpectedSourceDigest: DigestBytes(source), Condition: ConditionSemanticEquivalence,
		FromDecision: DecisionPass, ToDecision: DecisionFailClosed,
	}
	proposal, err := ProposePolicyDecisionRevision("policy.gooo", source, "metapolicycompilation", "metapolicycompilation", revision)
	if err != nil {
		t.Fatal(err)
	}
	input := func(id string, policy CompiledPolicy) Case {
		return Case{
			ID: id, ValidatorExpectation: DecisionPass,
			EvidenceClass: EvidenceSyntheticFixture, Provenance: "synthetic caller-owned revision fixture",
			ProducerAvailable: true, ConsumerAvailable: true,
			ObservedSourceDigest: policy.SourceDigest, ObservedArtifactSourceDigest: policy.SourceDigest,
			ObservedGeneratedJudgeDigest: DigestBytes(GenerateJudge(policy)), ObservedIndependentDigest: policy.SemanticDigest,
		}
	}
	before, after := input("fresh-pair", proposal.Original), input("fresh-pair", proposal.Candidate)
	after.ValidatorExpectation = DecisionFailClosed
	stale := input("unchanged-predecessor", proposal.Original)
	missing := Case{ID: "missing", EvidenceClass: EvidenceSyntheticFixture, Provenance: "synthetic absent evidence"}
	return source, PolicyRevisionObservationRequest{
		ExpectedSourceDigest: revision.ExpectedSourceDigest, Condition: revision.Condition,
		FromDecision: revision.FromDecision, ToDecision: revision.ToDecision,
		Cases: []PolicyRevisionCasePair{
			{Baseline: before, Candidate: after},
			{Baseline: stale, Candidate: stale},
			{Baseline: missing, Candidate: missing},
		},
	}
}

func TestObservePolicyDecisionRevisionExecutesSourceBoundPairsWithoutRebinding(t *testing.T) {
	source, request := revisionObservationFixture(t)
	sourceBefore := append([]byte(nil), source...)
	requestBefore := request
	requestBefore.Cases = append([]PolicyRevisionCasePair(nil), request.Cases...)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	report, err := ObservePolicyDecisionRevision(ctx, "policy.gooo", source, "metapolicycompilation", "metapolicycompilation", request)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(source, sourceBefore) || !reflect.DeepEqual(request, requestBefore) ||
		!reflect.DeepEqual(report.Request, requestBefore) {
		t.Fatal("revision execution rewrote caller source or observations")
	}
	want := PolicyRevisionObservationCounts{
		RequestedCasePairs: 3, ObservedCasePairs: 3, SameInputObservedPairs: 2,
		RequestedTransitionsObserved: 1, SourceComparisons: 12, ReplayComparisons: 6,
	}
	if report.Counts != want || report.ExecutionStatus != "COMPLETED" || report.ExecutionConformance != "PASS" {
		t.Fatalf("unexpected exact execution accounting: %+v status=%s conformance=%s", report.Counts, report.ExecutionStatus, report.ExecutionConformance)
	}
	if report.Baseline.FirstResults[0].Decision != DecisionPass ||
		report.Candidate.FirstResults[0].Decision != DecisionFailClosed ||
		report.Candidate.FirstResults[1].MatchedCondition != ConditionSourceMismatch ||
		report.Candidate.FirstResults[2].Decision != DecisionUnknown {
		t.Fatalf("requested change, stale evidence, or missing evidence was hidden: %+v", report.Transitions)
	}
	if report.Transitions[0].InputsIdentical || !report.Transitions[0].RequestedTransitionObserved ||
		report.Transitions[1].RequestedTransitionObserved || report.Transitions[2].RequestedTransitionObserved {
		t.Fatalf("observed transition was attributed to the wrong pair: %+v", report.Transitions)
	}
	for _, phase := range []PolicyRevisionExecution{report.Baseline, report.Candidate} {
		if !phase.Complete || len(phase.FirstResults) != 3 || len(phase.ReplayResults) != 3 ||
			DigestBytes([]byte(phase.GeneratedJudgeSource)) != phase.GeneratedJudgeDigest || phase.WallMilliseconds < 0 {
			t.Fatal("generated source or fresh replay evidence is incomplete")
		}
	}
	if report.Admission.State != "UNKNOWN" || report.Admission.Stage == "" || report.Admission.Step == "" ||
		report.Admission.Reason != "INDEPENDENT_REVISION_EVIDENCE_MISSING" || report.Admission.UnknownClass != "DIRECT_MISSING" ||
		report.Admission.NextOperation == "" || report.Admission.BlockedBy == nil ||
		report.Improvement != "UNKNOWN" || report.MutationAuthority != 0 || report.PromotionAuthority != 0 ||
		report.InputProvenance != "CALLER_DECLARED_NOT_VERIFIED" || report.RepositoryObservation != "NOT_PERFORMED" {
		t.Fatal("execution evidence acquired unobserved admission, provenance, or authority")
	}
	requestBytes, err := canonicalJSON(request)
	if err != nil || report.RequestDigest != DigestBytes(requestBytes) {
		t.Fatal("canonical request binding differs from the exact caller request")
	}
}

func TestObservePolicyDecisionRevisionRejectsInvalidInputBeforeExecution(t *testing.T) {
	source, original := revisionObservationFixture(t)
	cases := []struct {
		name string
		edit func(*PolicyRevisionObservationRequest)
	}{
		{"empty", func(r *PolicyRevisionObservationRequest) { r.Cases = nil }},
		{"duplicate", func(r *PolicyRevisionObservationRequest) { r.Cases = append(r.Cases, r.Cases[0]) }},
		{"mismatched-id", func(r *PolicyRevisionObservationRequest) { r.Cases[0].Candidate.ID = "different" }},
		{"empty-id", func(r *PolicyRevisionObservationRequest) { r.Cases[0].Baseline.ID, r.Cases[0].Candidate.ID = "", "" }},
		{"missing-provenance", func(r *PolicyRevisionObservationRequest) { r.Cases[0].Baseline.Provenance = "" }},
		{"missing-evidence-class", func(r *PolicyRevisionObservationRequest) { r.Cases[0].Candidate.EvidenceClass = "" }},
		{"stale-source", func(r *PolicyRevisionObservationRequest) { r.ExpectedSourceDigest = DigestBytes([]byte("different source")) }},
		{"unknown-condition", func(r *PolicyRevisionObservationRequest) { r.Condition = "unrecognized" }},
		{"no-change", func(r *PolicyRevisionObservationRequest) { r.ToDecision = r.FromDecision }},
	}
	for _, current := range cases {
		t.Run(current.name, func(t *testing.T) {
			request := original
			request.Cases = append([]PolicyRevisionCasePair(nil), original.Cases...)
			current.edit(&request)
			report, err := ObservePolicyDecisionRevision(context.Background(), "policy.gooo", source, "metapolicycompilation", "metapolicycompilation", request)
			if err == nil || report.Schema != "" || report.Baseline.Complete || report.Candidate.Complete {
				t.Fatalf("invalid input acquired execution evidence: report=%+v error=%v", report, err)
			}
		})
	}
}

func TestObservePolicyDecisionRevisionPreservesCancelledAttempt(t *testing.T) {
	source, request := revisionObservationFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	report, err := ObservePolicyDecisionRevision(ctx, "policy.gooo", source, "metapolicycompilation", "metapolicycompilation", request)
	if !errors.Is(err, context.Canceled) || report.ExecutionStatus != "FAILED" || report.ExecutionConformance != "UNKNOWN" ||
		report.Counts.FailedBatches != 1 || report.Counts.ObservedCasePairs != 0 || report.Candidate.Complete ||
		report.Baseline.ExecutionError == "" || len(report.Pending) == 0 || report.Pending[0].Stage != "BASELINE_EXECUTION" {
		t.Fatalf("cancelled attempt was lost or reported as execution: report=%+v error=%v", report, err)
	}
}

func TestDecodePolicyRevisionObservationRequestRejectsAmbiguousDocuments(t *testing.T) {
	_, request := revisionObservationFixture(t)
	document, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodePolicyRevisionObservationRequest(document); err != nil {
		t.Fatal(err)
	}
	for name, document := range map[string]string{
		"duplicate": strings.Replace(string(document), "\"condition\":", "\"condition\":\"unknown\",\"condition\":", 1),
		"unknown": strings.TrimSuffix(string(document), "}") + ",\"authority\":true}",
		"trailing": string(document) + "{}",
		"null": "null",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodePolicyRevisionObservationRequest([]byte(document)); err == nil {
				t.Fatal("ambiguous request was accepted")
			}
		})
	}
}

func TestRevisionObservationAccountingRefutesReplayCounterexample(t *testing.T) {
	_, request := revisionObservationFixture(t)
	result := DecisionResult{CaseID: request.Cases[0].Baseline.ID, Decision: DecisionPass}
	phase := PolicyRevisionExecution{
		SourceResults: []DecisionResult{result}, FirstResults: []DecisionResult{result},
		ReplayResults: []DecisionResult{result}, Complete: true,
	}
	report := PolicyRevisionObservation{Request: request, Baseline: phase, Candidate: phase}
	report.Candidate.ReplayResults = append([]DecisionResult(nil), phase.ReplayResults...)
	report.Candidate.ReplayResults[0].Decision = DecisionFailClosed
	completeRevisionObservation(&report)
	if report.ExecutionConformance != "REFUTED" || report.Counts.SourceMismatches != 1 || report.Counts.ReplayMismatches != 1 {
		t.Fatalf("known counterexample became success or missing evidence: %+v", report.Counts)
	}
}
