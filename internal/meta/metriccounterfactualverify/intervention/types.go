package metricintervention

import (
	counter "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactual"
	counterverify "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualverify"
	transition "github.com/kimjooyoon/meta-ontology-go/internal/meta/metrictransition"
)

const (
	LedgerSchema   = "gooo/metric-intervention-ledger/v1"
	BaselineSchema = "gooo/metric-intervention-baseline/v1"
	RegistrySchema = "gooo/metric-dimension-registry/v1"
)

type RootPolicy struct {
	CountsApplicability   string `json:"counts_applicability"`
	TopologyApplicability string `json:"topology_applicability"`
	TopologyReason        string `json:"topology_reason"`
	READMERequirement     string `json:"readme_requirement"`
}
type Baseline struct {
	Schema                string            `json:"schema"`
	RepositoryStateSchema string            `json:"repository_state_schema"`
	SourceIndicatorSchema string            `json:"source_indicator_schema"`
	SourcePolicySchema    string            `json:"source_policy_schema"`
	Repository            string            `json:"repository"`
	SubjectSHA            string            `json:"subject_sha"`
	Root                  transition.Counts `json:"root"`
	RootPolicy            RootPolicy        `json:"root_policy"`
	SourceMetricsDigest   string            `json:"source_metrics_digest"`
}

type Projection struct {
	DimensionID    string `json:"dimension_id"`
	Kind           string `json:"kind"`
	Baseline       int    `json:"baseline"`
	PredictedDelta int    `json:"predicted_delta"`
	ObservedDelta  int    `json:"observed_delta"`
	Residual       int    `json:"residual"`
	Projected      int    `json:"projected"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
}

type Indicator struct {
	ID             string `json:"id"`
	Family         string `json:"family"`
	Trilemma       string `json:"trilemma"`
	MetaOperation  string `json:"meta_operation"`
	Expected       string `json:"expected"`
	Actual         string `json:"actual"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
}

type Ledger struct {
	Schema                     string                `json:"schema"`
	Repository                 string                `json:"repository"`
	SubjectSHA                 string                `json:"subject_sha"`
	ExecutionPolicy            string                `json:"execution_policy"`
	ScenarioKind               string                `json:"scenario_kind"`
	Baseline                   Baseline              `json:"baseline"`
	BaselineDigest             string                `json:"baseline_digest"`
	Registry                   Registry              `json:"registry"`
	RegistryDigest             string                `json:"registry_digest"`
	PredictedDelta             counter.Delta         `json:"predicted_delta"`
	ObservedDelta              counter.Delta         `json:"observed_delta"`
	Counterfactual             counter.Ledger        `json:"counterfactual"`
	CounterfactualVerification counterverify.Receipt `json:"counterfactual_verification"`
	Projections                []Projection          `json:"projections"`
	Indicators                 []Indicator           `json:"indicators"`
	RepositoryWorkspaceWrites  bool                  `json:"repository_workspace_writes"`
	PromotionAuthorized        bool                  `json:"promotion_authorized"`
	Digest                     string                `json:"digest"`
}
