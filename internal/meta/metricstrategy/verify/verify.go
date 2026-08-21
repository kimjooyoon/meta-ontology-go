package metricstrategyverify

import (
	"fmt"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	strategy "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy"
)

func Replay(metricsPath, ledgerPath, interventionReceiptPath string, plan strategy.Plan) (Receipt, error) {
	if !validPlan(plan) || plan.Schema != strategy.PlanSchema || plan.ExecutionPolicy != strategy.ExecutionPolicy || plan.RepositoryWorkspaceWrites || plan.PromotionAuthorized {
		return Receipt{}, fmt.Errorf("metric strategy plan boundary is invalid")
	}
	inputs, err := loadInputs(metricsPath, ledgerPath, interventionReceiptPath, plan)
	if err != nil {
		return Receipt{}, err
	}
	bindings, err := replayBindings(inputs.ledger.Indicators)
	if err != nil {
		return Receipt{}, err
	}
	candidates, err := replayCandidates(bindings)
	if err != nil {
		return Receipt{}, err
	}
	if !artifact.Equal(bindings, plan.Bindings) || !artifact.Equal(candidates, plan.Candidates) || !artifact.Equal(replaySelection(candidates, inputs.ledger.Projections), plan.Selection) {
		return Receipt{}, fmt.Errorf("metric strategy independent synthesis diverged")
	}
	expectedInput := strategy.InputEvidence{SourceIndicatorSchema: inputs.baseline.SourceIndicatorSchema, SourcePolicySchema: inputs.baseline.SourcePolicySchema, SourceMetricsDigest: inputs.baseline.SourceMetricsDigest, InterventionSchema: inputs.ledger.Schema, InterventionDigest: inputs.ledger.Digest, VerificationSchema: inputs.receipt.Schema, VerificationDigest: inputs.receipt.Digest, IndicatorCount: len(bindings), ProjectionCount: len(inputs.ledger.Projections)}
	expectedRoot := strategy.RootPolicy{CountsApplicability: "OBSERVED", TopologyApplicability: "NOT_APPLICABLE", TopologyReason: "ROOT_TOPOLOGY_EXEMPT", READMERequirement: "NOT_APPLICABLE"}
	expectedPolicy := strategy.StrategyPolicy{Schema: strategy.PolicySchema, Choices: replayChoices(), FailureRule: "FIRST_UNSATISFIED_CANONICAL_FAMILY", FixedPointRule: "REGRESSION_TERMINATES_AT_VERIFIED_ZERO_RESIDUAL"}
	if !artifact.Equal(plan.Input, expectedInput) || !artifact.Equal(plan.RootPolicy, expectedRoot) || !artifact.Equal(plan.Policy, expectedPolicy) {
		return Receipt{}, fmt.Errorf("metric strategy policy or evidence binding diverged")
	}
	receipt := Receipt{Schema: ReceiptSchema, PlanDigest: plan.Digest, SourceMetricsDigest: inputs.baseline.SourceMetricsDigest, InterventionDigest: inputs.ledger.Digest, BindingCount: len(bindings), CandidateCount: len(candidates), SelectedProofChoice: plan.Selection.ProofChoice, Status: "VERIFIED", RepositoryWorkspaceWrites: false, PromotionAuthorized: false}
	return sealReceipt(receipt)
}

