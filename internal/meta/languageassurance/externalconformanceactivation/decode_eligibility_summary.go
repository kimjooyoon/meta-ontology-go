package externalconformanceactivation

type eligibilitySummary struct {
	AssuranceDenominator       int `json:"assurance_denominator"`
	BeforeOperating            int `json:"before_operating"`
	ProjectedOperating         int `json:"projected_operating"`
	OfficialOperating          int `json:"official_operating"`
	BeforeCoverageBPS          int `json:"before_coverage_bps"`
	ProjectedCoverageBPS       int `json:"projected_coverage_bps"`
	OfficialCoverageBPS        int `json:"official_coverage_bps"`
	EvidenceTotal              int `json:"evidence_total"`
	EvidenceExact              int `json:"evidence_exact"`
	ParentCompleted            int `json:"parent_completed"`
	ParentTotal                int `json:"parent_total"`
	ParentCoverageBPS          int `json:"parent_coverage_bps"`
	ParentKnownFailures        int `json:"parent_known_failures"`
	CapabilityCompleted        int `json:"capability_completed"`
	CapabilityTotal            int `json:"capability_total"`
	CapabilityCoverageBPS      int `json:"capability_coverage_bps"`
	CapabilityOutcomes         int `json:"capability_outcomes"`
	CapabilityOutcomeTotal     int `json:"capability_outcome_total"`
	CapabilitySuitePassed      int `json:"capability_suite_passed"`
	CapabilitySuiteTotal       int `json:"capability_suite_total"`
	CapabilitySuiteCoverageBPS int `json:"capability_suite_coverage_bps"`
	ExternalExecutions         int `json:"external_executions"`
	EligiblePaths              int `json:"eligible_paths"`
	UnknownPaths               int `json:"unknown_paths"`
	RepositoryWrites           int `json:"repository_writes"`
	ExternalRepositoryWrites   int `json:"external_repository_writes"`
	OfficialMutations          int `json:"official_mutations"`
	Promotions                 int `json:"promotions"`
	IndicatorCompleted         int `json:"indicator_completed"`
	IndicatorTotal             int `json:"indicator_total"`
	IndicatorCoverageBPS       int `json:"indicator_coverage_bps"`
	DriverCompleted            int `json:"driver_completed"`
	DriverTotal                int `json:"driver_total"`
	OutcomeCompleted           int `json:"outcome_completed"`
	OutcomeTotal               int `json:"outcome_total"`
	GuardrailCompleted         int `json:"guardrail_completed"`
	GuardrailTotal             int `json:"guardrail_total"`
}
