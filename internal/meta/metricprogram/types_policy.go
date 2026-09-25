package metricprogram

const (
	StrategySchemaVersion             = "gooo/metric-meta-strategy/v1"
	StrategyVerificationSchemaVersion = "gooo/metric-meta-strategy-verification/v1"
	ProgramSchemaVersion              = "gooo/metric-meta-program/v1"
	ProgramVerificationSchemaVersion  = "gooo/metric-meta-program-verification/v1"
	StrategyExecutionPolicy           = "READ_ONLY_ARTIFACT_SYNTHESIS"
	ProgramExecutionPolicy            = "READ_ONLY_META_PROGRAM"
	ProgramSourceFilename             = "program.gooo"
)

type StrategyInput struct {
	SourceIndicatorSchema string `json:"source_indicator_schema"`
	SourcePolicySchema    string `json:"source_policy_schema"`
	SourceMetricsDigest   string `json:"source_metrics_digest"`
	InterventionSchema    string `json:"intervention_schema"`
	InterventionDigest    string `json:"intervention_digest"`
	VerificationSchema    string `json:"verification_schema"`
	VerificationDigest    string `json:"verification_digest"`
	IndicatorCount        int    `json:"indicator_count"`
	ProjectionCount       int    `json:"projection_count"`
}

type RootPolicy struct {
	CountsApplicability   string `json:"counts_applicability"`
	TopologyApplicability string `json:"topology_applicability"`
	TopologyReason        string `json:"topology_reason"`
	ReadmeRequirement     string `json:"readme_requirement"`
}

type StrategyPolicy struct {
	Schema         string   `json:"schema"`
	Choices        []string `json:"choices"`
	FailureRule    string   `json:"failure_rule"`
	FixedPointRule string   `json:"fixed_point_rule"`
}
