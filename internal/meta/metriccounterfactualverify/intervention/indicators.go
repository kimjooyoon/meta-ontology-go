package metricintervention

import (
	"fmt"
	"strings"

	metric "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactual"
	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
	verify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify"
)

func EvaluateIndicators(baseline Baseline, baselineDigest string, registry Registry, registryDigest string, counter metric.Ledger, replay verify.Receipt, projections []Projection) ([]Indicator, error) {
	subject := baseline.Repository + "@" + baseline.SubjectSHA
	indicators := []Indicator{
		indicator("MIV-FOUNDATION-SUBJECT-001", "FOUNDATION", "AXIOM", "bind-exact-source-metrics", subject, counter.Repository+"@"+counter.SubjectSHA, subject == counter.Repository+"@"+counter.SubjectSHA, baselineDigest),
		indicator("MIV-FOUNDATION-ROOT-002", "FOUNDATION", "AXIOM", "exempt-project-root-topology", "counts=OBSERVED topology=NOT_APPLICABLE readme=NOT_APPLICABLE", rootPolicyText(baseline.RootPolicy), rootPolicySatisfied(baseline.RootPolicy), baselineDigest),
		indicator("MIV-FOUNDATION-REGISTRY-003", "FOUNDATION", "AXIOM", "interpret-dimension-registry", "schema="+RegistrySchema+" dimensions=10", fmt.Sprintf("schema=%s dimensions=%d", registry.Schema, len(registry.Dimensions)), ValidateRegistry(registry) == nil, registryDigest),
	}
	for index, projection := range projections {
		dimension := registry.Dimensions[index]
		id := "MIV-" + dimension.Family + "-" + strings.ToUpper(strings.ReplaceAll(dimension.ID, "_", "-"))
		expected := fmt.Sprintf("predicted_delta=%d", projection.PredictedDelta)
		actual := fmt.Sprintf("observed_delta=%d residual=%d", projection.ObservedDelta, projection.Residual)
		indicators = append(indicators, indicator(id, dimension.Family, dimension.Trilemma, dimension.MetaOperation, expected, actual, projection.Residual == 0, projection.EvidenceDigest))
	}
	replayOK := replay.Status == "VERIFIED" && replay.LedgerDigest == counter.Digest && !replay.PromotionAuthorized
	indicators = append(indicators,
		indicator("MIV-REGRESSION-REPLAY-001", "REGRESSION", "REGRESS", "replay-counterfactual", "VERIFIED", replay.Status, replayOK, replay.Digest),
		indicator("MIV-REGRESSION-BOUNDARY-002", "REGRESSION", "REGRESS", "preserve-repository-workspace", "writes=false promotion=false", fmt.Sprintf("writes=%t promotion=%t", counter.RepositoryWorkspaceWrites, counter.PromotionAuthorized), !counter.RepositoryWorkspaceWrites && !counter.PromotionAuthorized, counter.Digest),
	)
	return indicators, nil
}

func indicator(id, family, trilemma, operation, expected, actual string, satisfied bool, evidence string) Indicator {
	value := Indicator{ID: id, Family: family, Trilemma: trilemma, MetaOperation: operation, Expected: expected, Actual: actual, Status: metricStatus(satisfied), EvidenceDigest: evidence}
	if digest, err := artifact.Digest(value); err == nil {
		value.EvidenceDigest = digest
	}
	return value
}

func rootPolicyText(policy RootPolicy) string {
	return fmt.Sprintf("counts=%s topology=%s readme=%s", policy.CountsApplicability, policy.TopologyApplicability, policy.READMERequirement)
}

func rootPolicySatisfied(policy RootPolicy) bool {
	return policy.CountsApplicability == "OBSERVED" && policy.TopologyApplicability == "NOT_APPLICABLE" && policy.TopologyReason == "ROOT_TOPOLOGY_EXEMPT" && policy.READMERequirement == "NOT_APPLICABLE"
}
