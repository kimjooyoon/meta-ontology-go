package metricstrategy

const (
	PlanSchema      = "gooo/metric-meta-strategy/v1"
	PolicySchema    = "gooo/munchhausen-strategy-policy/v1"
	ExecutionPolicy = "READ_ONLY_ARTIFACT_SYNTHESIS"
)

type RootPolicy struct {
	CountsApplicability   string `json:"counts_applicability"`
	TopologyApplicability string `json:"topology_applicability"`
	TopologyReason        string `json:"topology_reason"`
	READMERequirement     string `json:"readme_requirement"`
}

type InputEvidence struct {
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

type Binding struct {
	IndicatorID   string `json:"indicator_id"`
	Family        string `json:"family"`
	Trilemma      string `json:"trilemma"`
	MetaOperation string `json:"meta_operation"`
	Expected      string `json:"expected"`
	Actual        string `json:"actual"`
	Status        string `json:"status"`
	EvidenceDigest string `json:"evidence_digest"`
}

