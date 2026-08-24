package externalconformanceactivation

type Summary struct {
	DenominatorTotal               int `json:"denominator_total"`
	BeforeOperating                int `json:"before_operating"`
	AfterOperating                 int `json:"after_operating"`
	BeforeCoverageBPS              int `json:"before_coverage_bps"`
	AfterCoverageBPS               int `json:"after_coverage_bps"`
	CapsulesTotal                  int `json:"capsules_total"`
	CapsulesExact                  int `json:"capsules_exact"`
	CapsuleCoverageBPS             int `json:"capsule_coverage_bps"`
	PredecessorSemanticsBPS        int `json:"predecessor_semantics_bps"`
	EligibilityIndicatorsTotal     int `json:"eligibility_indicators_total"`
	EligibilityIndicatorsSatisfied int `json:"eligibility_indicators_satisfied"`
	ParentCompleted                int `json:"parent_completed"`
	ParentTotal                    int `json:"parent_total"`
	ParentKnownFailures            int `json:"parent_known_failures"`
	SelectedCompleted              int `json:"selected_completed"`
	SelectedTotal                  int `json:"selected_total"`
	ExternalExecutions             int `json:"external_executions"`
	MergeRelationsTotal            int `json:"merge_relations_total"`
	MergeRelationsSatisfied        int `json:"merge_relations_satisfied"`
	UnknownPaths                   int `json:"unknown_paths"`
	BlockedPaths                   int `json:"blocked_paths"`
}
