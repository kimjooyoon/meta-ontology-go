package metricprogram_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
)

func fixturePayloads(t *testing.T) ([]byte, []byte) {
	t.Helper()
	bindings := []metricprogram.StrategyBinding{
		fixtureBinding("MIV-COHERENCE-DIRECT-FILES", "COHERENCE", "COHERENCE", "project-algebraic-root-state"),
		fixtureBinding("MIV-COHERENCE-DIRECT-FOLDERS", "COHERENCE", "COHERENCE", "project-algebraic-root-state"),
		fixtureBinding("MIV-COHERENCE-GO-FILES", "COHERENCE", "COHERENCE", "project-algebraic-root-state"),
		fixtureBinding("MIV-COHERENCE-GO-LINES", "COHERENCE", "COHERENCE", "project-algebraic-root-state"),
		fixtureBinding("MIV-COHERENCE-GOOO-FILES", "COHERENCE", "COHERENCE", "project-algebraic-root-state"),
		fixtureBinding("MIV-COHERENCE-GOOO-LINES", "COHERENCE", "COHERENCE", "project-algebraic-root-state"),
		fixtureBinding("MIV-COHERENCE-RECURSIVE-FILES", "COHERENCE", "COHERENCE", "project-algebraic-root-state"),
		fixtureBinding("MIV-COHERENCE-RECURSIVE-FOLDERS", "COHERENCE", "COHERENCE", "project-algebraic-root-state"),
		fixtureBinding("MIV-FOUNDATION-REGISTRY-003", "FOUNDATION", "AXIOM", "interpret-dimension-registry"),
		fixtureBinding("MIV-FOUNDATION-ROOT-002", "FOUNDATION", "AXIOM", "exempt-project-root-topology"),
		fixtureBinding("MIV-FOUNDATION-SUBJECT-001", "FOUNDATION", "AXIOM", "bind-exact-source-metrics"),
		fixtureBinding("MIV-REGRESSION-BOUNDARY-002", "REGRESSION", "REGRESS", "preserve-repository-workspace"),
		fixtureBinding("MIV-REGRESSION-CHANGED-DIRECTORIES", "REGRESSION", "REGRESS", "observe-counterfactual-boundary"),
		fixtureBinding("MIV-REGRESSION-CHANGED-FILES", "REGRESSION", "REGRESS", "observe-counterfactual-boundary"),
		fixtureBinding("MIV-REGRESSION-NON-PROMOTING-TERMINAL", "REGRESSION", "REGRESS", "preserve-non-promoting-terminal"),
		fixtureBinding("MIV-REGRESSION-REPLAY-001", "REGRESSION", "REGRESS", "replay-counterfactual"),
	}
	foundationDigest := fixtureDigest("foundation")
	coherenceDigest := fixtureDigest("coherence")
	regressionDigest := fixtureDigest("regression")
	planDigest := fixtureDigest("plan")
	plan := metricprogram.StrategyPlan{
		Schema: metricprogram.StrategySchemaVersion, Repository: "kimjooyoon/meta-ontology-go", SubjectSHA: "b8626224d7f5ce425f3132a354ec2fd480c659cb",
		ExecutionPolicy: metricprogram.StrategyExecutionPolicy,
		Input:           metricprogram.StrategyInput{SourceIndicatorSchema: "gooo/indicator-report/v3", SourcePolicySchema: "gooo/source-policy/v1", SourceMetricsDigest: fixtureDigest("metrics"), InterventionSchema: "gooo/metric-intervention-ledger/v1", InterventionDigest: fixtureDigest("intervention"), VerificationSchema: "gooo/metric-intervention-verification/v1", VerificationDigest: fixtureDigest("intervention-verification"), IndicatorCount: len(bindings), ProjectionCount: 10},
		RootPolicy:      metricprogram.RootPolicy{CountsApplicability: "OBSERVED", TopologyApplicability: "NOT_APPLICABLE", TopologyReason: "ROOT_TOPOLOGY_EXEMPT", ReadmeRequirement: "NOT_APPLICABLE"},
		Policy:          metricprogram.StrategyPolicy{Schema: "gooo/munchhausen-strategy-policy/v1", Choices: []string{"FOUNDATION", "COHERENCE", "REGRESSION"}, FailureRule: "FIRST_UNSATISFIED_CANONICAL_FAMILY", FixedPointRule: "REGRESSION_TERMINATES_AT_VERIFIED_ZERO_RESIDUAL"},
		Bindings:        bindings,
		Candidates: []metricprogram.StrategyCandidate{
			{ProofChoice: "FOUNDATION", IndicatorIDs: []string{"MIV-FOUNDATION-REGISTRY-003", "MIV-FOUNDATION-ROOT-002", "MIV-FOUNDATION-SUBJECT-001"}, MetaOperations: []string{"bind-exact-source-metrics", "exempt-project-root-topology", "interpret-dimension-registry"}, IndicatorCount: 3, Admissible: true, EvidenceDigest: foundationDigest},
			{ProofChoice: "COHERENCE", IndicatorIDs: []string{"MIV-COHERENCE-DIRECT-FILES", "MIV-COHERENCE-DIRECT-FOLDERS", "MIV-COHERENCE-GO-FILES", "MIV-COHERENCE-GO-LINES", "MIV-COHERENCE-GOOO-FILES", "MIV-COHERENCE-GOOO-LINES", "MIV-COHERENCE-RECURSIVE-FILES", "MIV-COHERENCE-RECURSIVE-FOLDERS"}, MetaOperations: []string{"project-algebraic-root-state"}, IndicatorCount: 8, Admissible: true, EvidenceDigest: coherenceDigest},
			{ProofChoice: "REGRESSION", IndicatorIDs: []string{"MIV-REGRESSION-BOUNDARY-002", "MIV-REGRESSION-CHANGED-DIRECTORIES", "MIV-REGRESSION-CHANGED-FILES", "MIV-REGRESSION-REPLAY-001", "MIV-REGRESSION-NON-PROMOTING-TERMINAL"}, MetaOperations: []string{"observe-counterfactual-boundary", "preserve-non-promoting-terminal", "preserve-repository-workspace", "replay-counterfactual"}, IndicatorCount: 5, Admissible: true, EvidenceDigest: regressionDigest},
		},
		Selection: metricprogram.StrategySelection{ProofChoice: "REGRESSION", Decision: "HOLD_FIXED_POINT", MetaOperation: "terminate-at-fixed-point", Reason: "ALL_INDICATORS_SATISFIED_AND_RESIDUALS_ZERO", CandidateDigest: regressionDigest, SourceMetaOperations: []string{"observe-counterfactual-boundary", "preserve-non-promoting-terminal", "preserve-repository-workspace", "replay-counterfactual"}},
		Digest:    planDigest,
	}
	verification := metricprogram.StrategyVerification{Schema: metricprogram.StrategyVerificationSchemaVersion, PlanDigest: planDigest, SourceMetricsDigest: plan.Input.SourceMetricsDigest, InterventionDigest: plan.Input.InterventionDigest, BindingCount: len(bindings), CandidateCount: 3, SelectedProofChoice: "REGRESSION", Status: "VERIFIED", Digest: fixtureDigest("strategy-verification")}
	return fixtureJSON(t, plan), fixtureJSON(t, verification)
}

func fixtureBinding(id, family, trilemma, operation string) metricprogram.StrategyBinding {
	return metricprogram.StrategyBinding{IndicatorID: id, Family: family, Trilemma: trilemma, MetaOperation: operation, Expected: "expected", Actual: "actual", Status: "SATISFIED", EvidenceDigest: fixtureDigest(id)}
}

func fixtureDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func fixtureJSON(t *testing.T, value any) []byte {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}
