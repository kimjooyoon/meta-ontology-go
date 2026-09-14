package policycompilation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

func DecodePolicyRevisionObservationRequest(data []byte) (PolicyRevisionObservationRequest, error) {
	var request PolicyRevisionObservationRequest
	if err := decodeStrictJSON(data, &request); err != nil {
		return request, err
	}
	return request, validateRevisionObservationRequest(request)
}

func validateRevisionObservationRequest(request PolicyRevisionObservationRequest) error {
	if !ValidDigest(request.ExpectedSourceDigest) || !knownCondition(request.Condition) ||
		!knownDecision(request.FromDecision) || !knownDecision(request.ToDecision) ||
		request.FromDecision == request.ToDecision || len(request.Cases) == 0 {
		return errors.New("revision observation requires an exact change and at least one case pair")
	}
	seen := make(map[string]bool, len(request.Cases))
	for _, pair := range request.Cases {
		if strings.TrimSpace(pair.Baseline.ID) == "" || pair.Baseline.ID != pair.Candidate.ID || seen[pair.Baseline.ID] {
			return errors.New("revision observation requires unique, matching case identities")
		}
		for _, input := range []Case{pair.Baseline, pair.Candidate} {
			if strings.TrimSpace(input.EvidenceClass) == "" || strings.TrimSpace(input.Provenance) == "" {
				return errors.New("revision observation requires explicit caller evidence class and provenance")
			}
		}
		seen[pair.Baseline.ID] = true
	}
	return nil
}

// ObservePolicyDecisionRevision generates and executes both source versions.
// Each version is built once for its initial cases and fresh process replays.
// Caller observations and validator expectations are never rebound or repaired.
func ObservePolicyDecisionRevision(ctx context.Context, filename string, source []byte, expectedPackage, expectedNamespace string, request PolicyRevisionObservationRequest) (PolicyRevisionObservation, error) {
	if err := validateRevisionObservationRequest(request); err != nil {
		return PolicyRevisionObservation{}, err
	}
	request.Cases = append([]PolicyRevisionCasePair(nil), request.Cases...)
	proposal, err := ProposePolicyDecisionRevision(filename, source, expectedPackage, expectedNamespace, PolicyDecisionRevision{
		ExpectedSourceDigest: request.ExpectedSourceDigest,
		Condition:            request.Condition, FromDecision: request.FromDecision, ToDecision: request.ToDecision,
	})
	if err != nil {
		return PolicyRevisionObservation{}, err
	}
	requestBytes, err := canonicalJSON(request)
	if err != nil {
		return PolicyRevisionObservation{}, err
	}
	report := PolicyRevisionObservation{
		Schema: PolicyRevisionObservationSchema, SourceFile: filename,
		RequestDigest: DigestBytes(requestBytes), Request: request,
		OriginalPolicy: proposal.Original, CandidatePolicy: proposal.Candidate,
		CandidateSource:    proposal.CandidateSource,
		ChangedCoordinates: append([]string(nil), proposal.ChangedCoordinates...),
		Admission: revisionPending("INDEPENDENT_VALIDATION", "OBSERVE_REVISION_CANDIDATE",
			"INDEPENDENT_REVISION_EVIDENCE_MISSING", "RUN_INDEPENDENT_REVISION_OBSERVER"),
		Pending:               []PolicyRevisionPending{},
		InputProvenance:       "CALLER_DECLARED_NOT_VERIFIED",
		RepositoryObservation: "NOT_PERFORMED", Improvement: "UNKNOWN",
	}
	baseline := make([]Case, 0, len(request.Cases))
	candidate := make([]Case, 0, len(request.Cases))
	for _, pair := range request.Cases {
		baseline = append(baseline, pair.Baseline)
		candidate = append(candidate, pair.Candidate)
	}
	report.Baseline, err = executeRevisionPolicy(ctx, proposal.Original, baseline)
	if err != nil {
		report.Pending = append(report.Pending, revisionPending("BASELINE_EXECUTION", "BUILD_AND_EXECUTE_GENERATED_BATCH",
			"BASELINE_GENERATED_BATCH_FAILED", "RERUN_BASELINE_GENERATED_BATCH"))
		completeRevisionObservation(&report)
		return report, fmt.Errorf("observe baseline policy: %w", err)
	}
	report.Candidate, err = executeRevisionPolicy(ctx, proposal.Candidate, candidate)
	if err != nil {
		report.Pending = append(report.Pending, revisionPending("CANDIDATE_EXECUTION", "BUILD_AND_EXECUTE_GENERATED_BATCH",
			"CANDIDATE_GENERATED_BATCH_FAILED", "RERUN_CANDIDATE_GENERATED_BATCH"))
		completeRevisionObservation(&report)
		return report, fmt.Errorf("observe candidate policy: %w", err)
	}
	completeRevisionObservation(&report)
	if report.ExecutionConformance == "REFUTED" {
		return report, errors.New("generated revision execution contradicts source or replay observations")
	}
	return report, nil
}

func executeRevisionPolicy(ctx context.Context, policy CompiledPolicy, inputs []Case) (PolicyRevisionExecution, error) {
	start := time.Now()
	judge := GenerateJudge(policy)
	phase := PolicyRevisionExecution{
		GeneratedJudgeSource: string(judge), GeneratedJudgeDigest: DigestBytes(judge),
		DeclaredInputs: append([]Case(nil), inputs...),
		SourceResults:  make([]DecisionResult, 0, len(inputs)),
		FirstResults:   []DecisionResult{}, ReplayResults: []DecisionResult{},
	}
	for _, input := range inputs {
		phase.SourceResults = append(phase.SourceResults, EvaluateSourcePolicy(policy, input))
	}
	executions := make([]Case, 0, 2*len(inputs))
	executions = append(executions, inputs...)
	executions = append(executions, inputs...)
	results, err := ExecuteGeneratedBatch(ctx, judge, executions)
	phase.WallMilliseconds = time.Since(start).Milliseconds()
	firstCount := min(len(inputs), len(results))
	phase.FirstResults = append(phase.FirstResults, results[:firstCount]...)
	if len(results) > len(inputs) {
		phase.ReplayResults = append(phase.ReplayResults, results[len(inputs):]...)
	}
	if err == nil && len(results) != len(executions) {
		err = errors.New("generated revision batch did not return its exact result count")
	}
	phase.Complete = err == nil
	if err != nil {
		phase.ExecutionError = err.Error()
	}
	return phase, err
}

func revisionPending(stage, step, reason, next string) PolicyRevisionPending {
	return PolicyRevisionPending{
		State: "UNKNOWN", Stage: stage, Step: step, Reason: reason,
		UnknownClass: "DIRECT_MISSING", NextOperation: next, BlockedBy: []string{},
	}
}

func completeRevisionObservation(report *PolicyRevisionObservation) {
	counts := PolicyRevisionObservationCounts{RequestedCasePairs: len(report.Request.Cases)}
	for _, phase := range []PolicyRevisionExecution{report.Baseline, report.Candidate} {
		if phase.ExecutionError != "" {
			counts.FailedBatches++
		}
		for _, results := range [][]DecisionResult{phase.FirstResults, phase.ReplayResults} {
			for index, result := range results {
				counts.SourceComparisons++
				if index >= len(phase.SourceResults) || !sameResult(result, phase.SourceResults[index]) {
					counts.SourceMismatches++
				}
			}
		}
		for index, result := range phase.ReplayResults {
			if index < len(phase.FirstResults) {
				counts.ReplayComparisons++
				if !sameResult(result, phase.FirstResults[index]) {
					counts.ReplayMismatches++
				}
			}
		}
	}
	report.Transitions = []PolicyRevisionObservedTransition{}
	for index, pair := range report.Request.Cases {
		if index >= len(report.Baseline.FirstResults) || index >= len(report.Candidate.FirstResults) {
			continue
		}
		before, after := report.Baseline.FirstResults[index], report.Candidate.FirstResults[index]
		transition := PolicyRevisionObservedTransition{
			CaseID: pair.Baseline.ID, InputsIdentical: pair.Baseline == pair.Candidate,
			BaselineCondition: before.MatchedCondition, CandidateCondition: after.MatchedCondition,
			BaselineDecision: before.Decision, CandidateDecision: after.Decision,
			RequestedTransitionObserved: before.MatchedCondition == report.Request.Condition &&
				after.MatchedCondition == report.Request.Condition &&
				before.Decision == report.Request.FromDecision && after.Decision == report.Request.ToDecision,
			CausalAttribution: "UNASSESSED",
		}
		counts.ObservedCasePairs++
		if transition.InputsIdentical {
			counts.SameInputObservedPairs++
		}
		if transition.RequestedTransitionObserved {
			counts.RequestedTransitionsObserved++
		}
		report.Transitions = append(report.Transitions, transition)
	}
	report.Counts = counts
	report.ExecutionStatus, report.ExecutionConformance = "NOT_COMPLETED", "UNKNOWN"
	if counts.FailedBatches != 0 {
		report.ExecutionStatus = "FAILED"
	} else if report.Baseline.Complete && report.Candidate.Complete {
		report.ExecutionStatus, report.ExecutionConformance = "COMPLETED", "PASS"
	}
	if counts.SourceMismatches != 0 || counts.ReplayMismatches != 0 {
		report.ExecutionConformance = "REFUTED"
	}
	if counts.RequestedTransitionsObserved == 0 {
		report.Pending = append(report.Pending, revisionPending("REQUEST_COVERAGE", "OBSERVE_REQUESTED_TRANSITION",
			"REQUESTED_TRANSITION_NOT_OBSERVED", "SUPPLY_CASE_PAIR_EXERCISING_REQUESTED_CONDITION"))
	}
}
