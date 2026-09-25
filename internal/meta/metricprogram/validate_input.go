package metricprogram

import (
	"fmt"
	"slices"
)

var proofChoices = []string{"FOUNDATION", "COHERENCE", "REGRESSION"}

func validateStrategy(plan StrategyPlan, verification StrategyVerification) error {
	if plan.Schema != StrategySchemaVersion || verification.Schema != StrategyVerificationSchemaVersion {
		return fmt.Errorf("strategy schema mismatch")
	}
	if plan.Repository == "" || !validSubjectSHA(plan.SubjectSHA) {
		return fmt.Errorf("strategy subject is not exact")
	}
	if plan.ExecutionPolicy != StrategyExecutionPolicy || plan.RepositoryWorkspaceWrites || plan.PromotionAuthorized {
		return fmt.Errorf("strategy is not read-only")
	}
	if plan.RootPolicy != (RootPolicy{CountsApplicability: "OBSERVED", TopologyApplicability: "NOT_APPLICABLE", TopologyReason: "ROOT_TOPOLOGY_EXEMPT", ReadmeRequirement: "NOT_APPLICABLE"}) {
		return fmt.Errorf("project root exception is not canonical")
	}
	if !validDigest(plan.Digest) || verification.PlanDigest != plan.Digest || !validDigest(verification.Digest) {
		return fmt.Errorf("strategy digest binding is invalid")
	}
	if verification.Status != "VERIFIED" || verification.RepositoryWorkspaceWrites || verification.PromotionAuthorized {
		return fmt.Errorf("strategy verification is not read-only VERIFIED evidence")
	}
	if verification.SourceMetricsDigest != plan.Input.SourceMetricsDigest || verification.InterventionDigest != plan.Input.InterventionDigest {
		return fmt.Errorf("strategy verification input digests do not match")
	}
	if !validDigest(plan.Input.SourceMetricsDigest) || !validDigest(plan.Input.InterventionDigest) || !validDigest(plan.Input.VerificationDigest) {
		return fmt.Errorf("strategy input digest is invalid")
	}
	if plan.Input.IndicatorCount != len(plan.Bindings) || verification.BindingCount != len(plan.Bindings) || verification.CandidateCount != len(plan.Candidates) {
		return fmt.Errorf("strategy cardinality binding is invalid")
	}
	if plan.Policy.Schema != "gooo/munchhausen-strategy-policy/v1" || !slices.Equal(plan.Policy.Choices, proofChoices) {
		return fmt.Errorf("Munchhausen policy is not canonical")
	}
	if plan.Policy.FailureRule != "FIRST_UNSATISFIED_CANONICAL_FAMILY" || plan.Policy.FixedPointRule != "REGRESSION_TERMINATES_AT_VERIFIED_ZERO_RESIDUAL" {
		return fmt.Errorf("Munchhausen selection rules are not canonical")
	}
	return validateStrategyMembers(plan, verification)
}
