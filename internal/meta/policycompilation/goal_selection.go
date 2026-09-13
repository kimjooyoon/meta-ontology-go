package policycompilation

import "context"

const GoalSelectionSchema = "gooo/meta-policy-goal-selection/v1"

type GoalSelectionReport struct {
	Schema                  string                     `json:"schema"`
	Decision                string                     `json:"decision"`
	Reason                  string                     `json:"reason"`
	GoalSourceDigest        string                     `json:"goal_source_digest"`
	GoalSemanticDigest      string                     `json:"goal_semantic_digest"`
	GoalPolicyID            string                     `json:"goal_policy_id"`
	CandidateSourceDigest   string                     `json:"candidate_source_digest"`
	CandidateSemanticDigest string                     `json:"candidate_semantic_digest"`
	GoalSemanticMatch       bool                       `json:"goal_semantic_match"`
	CriterionIndependence   string                     `json:"criterion_independence"`
	Execution               *NextPolicyExecutionReport `json:"execution,omitempty"`
	SelectedSource          string                     `json:"selected_source,omitempty"`
	SelectedSourceDigest    string                     `json:"selected_source_digest,omitempty"`
	MaterializedSource      string                     `json:"materialized_source,omitempty"`
	MaterializationBoundary string                     `json:"materialization_boundary"`
	Pending                 *PolicyRevisionPending     `json:"pending,omitempty"`
	Diagnostic              string                     `json:"diagnostic,omitempty"`
	GeneralAdmission        string                     `json:"general_admission"`
	Improvement             string                     `json:"improvement"`
	MutationAuthority       int                        `json:"mutation_authority"`
	PromotionAuthority      int                        `json:"promotion_authority"`
}

// SelectNextPolicyForGoal fixes a separate Gooo objective before executing the
// candidate. Independence is by source authority, not a diverse compiler. The
// candidate cannot replace the goal with its own caller-declared expectations.
func SelectNextPolicyForGoal(ctx context.Context, filename string, source, predecessor, request []byte, pkg, namespace string, goal []byte, expectedGoalDigest string) GoalSelectionReport {
	goal, source = append([]byte(nil), goal...), append([]byte(nil), source...)
	report := GoalSelectionReport{
		Schema: GoalSelectionSchema, Decision: "UNKNOWN", GoalSourceDigest: DigestBytes(goal),
		CandidateSourceDigest: DigestBytes(source), GeneralAdmission: "UNKNOWN", Improvement: "UNKNOWN",
		CriterionIndependence:   "SEPARATE_GOOO_GOAL_SHARED_COMPILER",
		MaterializationBoundary: "NEW_CALLER_OWNED_TEMP_DIRECTORY_ONLY",
	}
	if !ValidDigest(expectedGoalDigest) {
		return unknownGoalSelection(report, "GOAL", "PIN_FROZEN_GOOO_GOAL", "FROZEN_GOAL_PIN_MISSING", "DIRECT_MISSING")
	}
	if expectedGoalDigest != report.GoalSourceDigest {
		report.Decision, report.Reason = "REFUTED", "FROZEN_GOAL_DIGEST_MISMATCH"
		return report
	}
	target, err := CompileForIdentity("frozen-goal.gooo", goal, pkg, namespace)
	if err != nil {
		report.Diagnostic = err.Error()
		return unknownGoalSelection(report, "GOAL", "COMPILE_FROZEN_GOOO_GOAL", "FROZEN_GOAL_NOT_COMPILED", "AMBIGUOUS")
	}
	report.GoalSemanticDigest, report.GoalPolicyID = target.SemanticDigest, target.PolicyID
	candidate, err := CompileForIdentity(filename, source, pkg, namespace)
	if err != nil {
		report.Diagnostic = err.Error()
		return unknownGoalSelection(report, "CANDIDATE", "COMPILE_SELECTED_GOOO", "CANDIDATE_NOT_COMPILED", "AMBIGUOUS")
	}
	report.CandidateSemanticDigest = candidate.SemanticDigest
	report.GoalSemanticMatch = candidate.PolicyID == target.PolicyID && candidate.SemanticDigest == target.SemanticDigest
	if !report.GoalSemanticMatch {
		report.Decision, report.Reason = "REFUTED", "CANDIDATE_DOES_NOT_MATCH_FROZEN_GOOO_GOAL"
		return report
	}
	execution := ObserveNextPolicyExecution(ctx, filename, source, predecessor, request, pkg, namespace)
	report.Execution = &execution
	switch execution.Decision {
	case "REFUTED":
		report.Decision, report.Reason = "REFUTED", "GOAL_MATCHED_BUT_NEXT_EXECUTION_REFUTED"
	case "NEXT_EXECUTION_OBSERVED":
		report.Decision, report.Reason = "SELECTED_FOR_FROZEN_GOOO_GOAL", "FROZEN_GOAL_AND_NEXT_EXECUTION_SATISFIED"
		report.SelectedSource, report.SelectedSourceDigest = string(source), DigestBytes(source)
	default:
		report = unknownGoalSelection(report, "EXECUTION", "OBSERVE_GOAL_BOUND_NEXT_EXECUTION", "GOAL_MATCHED_BUT_EXECUTION_UNAVAILABLE", "DIRECT_MISSING")
		if execution.Pending != nil {
			report.Pending = execution.Pending
		}
	}
	return report
}

func unknownGoalSelection(report GoalSelectionReport, stage, step, reason, class string) GoalSelectionReport {
	report.Decision, report.Reason = "UNKNOWN", reason
	report.Pending = &PolicyRevisionPending{
		State: "UNKNOWN", Stage: stage, Step: step, Reason: reason, UnknownClass: class,
		NextOperation: "RESOLVE_GOAL_SELECTION_INPUT_AND_REPEAT", BlockedBy: []string{},
	}
	return report
}
