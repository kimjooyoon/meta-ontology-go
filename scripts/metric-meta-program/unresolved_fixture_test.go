package main

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram"
	"github.com/kimjooyoon/meta-ontology-go/internal/projectionbundle"
)

func unresolvedInputsFresh1733(t *testing.T, operation string) ([]byte, []byte) {
	t.Helper()
	digest := projectionbundle.Digest([]byte("fresh-source-input"))
	plan := metricprogram.StrategyPlan{
		Schema: metricprogram.StrategySchemaVersion, Repository: "example/compiler",
		SubjectSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Digest: digest,
		ExecutionPolicy: metricprogram.StrategyExecutionPolicy,
		RootPolicy: metricprogram.RootPolicy{CountsApplicability: "OBSERVED", TopologyApplicability: "NOT_APPLICABLE",
			TopologyReason: "ROOT_TOPOLOGY_EXEMPT", ReadmeRequirement: "NOT_APPLICABLE"},
		Policy: metricprogram.StrategyPolicy{Schema: "gooo/munchhausen-strategy-policy/v1",
			Choices:     []string{"FOUNDATION", "COHERENCE", "REGRESSION"},
			FailureRule: "FIRST_UNSATISFIED_CANONICAL_FAMILY", FixedPointRule: "REGRESSION_TERMINATES_AT_VERIFIED_ZERO_RESIDUAL"},
		Input: metricprogram.StrategyInput{SourceMetricsDigest: digest, InterventionDigest: digest,
			VerificationDigest: digest, IndicatorCount: 1},
		Bindings: []metricprogram.StrategyBinding{{IndicatorID: "upstream://indicator/new", Family: "FOUNDATION",
			Trilemma: "AXIOM", MetaOperation: operation, Status: "SATISFIED", EvidenceDigest: digest}},
	}
	verification := metricprogram.StrategyVerification{Schema: metricprogram.StrategyVerificationSchemaVersion,
		PlanDigest: digest, SourceMetricsDigest: digest, InterventionDigest: digest, BindingCount: 1,
		Status: "VERIFIED", Digest: digest}
	first, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(verification)
	if err != nil {
		t.Fatal(err)
	}
	return first, second
}
