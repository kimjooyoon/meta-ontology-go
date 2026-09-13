package policycompilation

import "context"

const PolicyGoalCounterexampleSchema = "gooo/meta-policy-goal-counterexample/v1"

type PolicyGoalCounterexampleRequest struct {
	ExpectedSourceDigest string `json:"expected_source_digest"`
	ExpectedGoalDigest string `json:"expected_goal_digest"`
	SourceCase Case `json:"source_case"`
	GoalCase Case `json:"goal_case"`
}

type PolicyGoalCounterexampleObservation struct {
	Schema string `json:"schema"`
	State string `json:"state"`
	Reason string `json:"reason"`
	SourceDigest string `json:"source_digest"`
	GoalDigest string `json:"goal_digest"`
	GoalSemanticDigest string `json:"goal_semantic_digest,omitempty"`
	RequestDigest string `json:"request_digest"`
	Request *PolicyGoalCounterexampleRequest `json:"declared_request,omitempty"`
	SourceResults []DecisionResult `json:"source_results"`
	GoalResults []DecisionResult `json:"goal_results"`
	Proposal *PolicyCounterexampleProposal `json:"proposal,omitempty"`
	Pending *PolicyRevisionPending `json:"pending,omitempty"`
	Diagnostic string `json:"diagnostic,omitempty"`
	GeneratedBatches int `json:"generated_batches"`
	ExpectationAuthority string `json:"expectation_authority"`
	CriterionIndependence string `json:"criterion_independence"`
	GeneralAdmission string `json:"general_admission"`
	Improvement string `json:"improvement"`
	MutationAuthority int `json:"mutation_authority"`
	PromotionAuthority int `json:"promotion_authority"`
}

// ObserveGoalPolicyCounterexample derives the expected decision from a fresh
// generated execution of a caller-pinned Gooo goal, not a caller's expectation.
// Source-specific input snapshots remain explicit; their digests are not rebound.
func ObserveGoalPolicyCounterexample(ctx context.Context, filename string, source, goal, raw []byte, pkg, namespace string) PolicyGoalCounterexampleObservation {
	source, goal, raw = append([]byte(nil), source...), append([]byte(nil), goal...), append([]byte(nil), raw...)
	report := PolicyGoalCounterexampleObservation{
		Schema: PolicyGoalCounterexampleSchema, State: "UNKNOWN",
		SourceDigest: DigestBytes(source), GoalDigest: DigestBytes(goal), RequestDigest: DigestBytes(raw),
		SourceResults: []DecisionResult{}, GoalResults: []DecisionResult{},
		ExpectationAuthority: "PINNED_GOOO_GOAL_GENERATED_EXECUTION",
		CriterionIndependence: "SEPARATE_GOOO_GOAL_SHARED_COMPILER",
		GeneralAdmission: "UNKNOWN", Improvement: "UNKNOWN",
	}
	policies, report, ready := prepareGoalCounterexample(report, filename, source, goal, raw, pkg, namespace)
	if !ready {
		return report
	}
	report.GeneratedBatches++
	results, err := ExecuteGeneratedBatch(ctx, GenerateJudge(policies[0]), []Case{report.Request.SourceCase})
	report.SourceResults = results
	if err != nil || len(results) != 1 {
		if err != nil {
			report.Diagnostic = err.Error()
		}
		return stopGoalCounterexample(report, "UNKNOWN", "SOURCE", "SOURCE_EXECUTION_INCOMPLETE", "DIRECT_MISSING")
	}
	report.GeneratedBatches++
	results, err = ExecuteGeneratedBatch(ctx, GenerateJudge(policies[1]), []Case{report.Request.GoalCase})
	report.GoalResults = results
	if err != nil || len(results) != 1 {
		if err != nil {
			report.Diagnostic = err.Error()
		}
		return stopGoalCounterexample(report, "UNKNOWN", "GOAL", "GOAL_EXECUTION_INCOMPLETE", "DIRECT_MISSING")
	}
	if !sameResult(results[0], EvaluateSourcePolicy(policies[1], report.Request.GoalCase)) {
		return stopGoalCounterexample(report, "REFUTED", "GOAL", "GENERATED_GOAL_CONTRADICTS_SOURCE", "")
	}
	if results[0].Decision == DecisionUnknown {
		return stopGoalCounterexample(report, "UNKNOWN", "GOAL", "GOAL_DECISION_UNKNOWN", "DEPENDENCY_BLOCKED")
	}
	if results[0].MatchedCondition != report.SourceResults[0].MatchedCondition {
		return stopGoalCounterexample(report, "REFUTED", "INPUT", "SOURCE_AND_GOAL_CONDITIONS_DIFFER", "")
	}
	input := report.Request.SourceCase
	input.ValidatorExpectation = results[0].Decision
	proposal, err := ProposePolicyRevisionFromCounterexample(filename, source, pkg, namespace, PolicyRevisionCounterexample{
		ExpectedSourceDigest: report.SourceDigest, Input: input, Observed: report.SourceResults[0],
	})
	report.Proposal = &proposal
	report.State, report.Reason, report.Pending = proposal.State, proposal.Reason, proposal.Pending
	if err != nil {
		report.Diagnostic = err.Error()
	}
	return report
}
