package verticalsliceclosureshadow

type artifactSummary struct {
	Satisfied            int `json:"satisfied"`
	Total                int `json:"total"`
	Executed             int `json:"executed"`
	NotSatisfied         int `json:"not_satisfied"`
	Unresolved           int `json:"unresolved"`
	CapabilitySatisfied  int `json:"capability_satisfied"`
	CapabilityTotal      int `json:"capability_total"`
	CapabilityExecuted   int `json:"capability_executed"`
	CapabilityUnresolved int `json:"capability_unresolved"`
	ReadinessBPS         int `json:"readiness_bps"`
	Coordinates          int `json:"coordinates"`
	BoundCoordinates     int `json:"bound_coordinates"`
	ReadinessCompleted   int `json:"readiness_completed"`
	ReadinessTotal       int `json:"readiness_total"`
	SemanticSatisfied    int `json:"semantic_satisfied"`
	SemanticTotal        int `json:"semantic_total"`
	SurfacesSatisfied    int `json:"surfaces_satisfied"`
	SurfacesTotal        int `json:"surfaces_total"`
	CasesSatisfied       int `json:"cases_satisfied"`
	CasesTotal           int `json:"cases_total"`
	ExecutedCases        int `json:"executed_cases"`
	CaseReadinessBPS     int `json:"case_readiness_bps"`
	IndicatorsSatisfied  int `json:"indicators_satisfied"`
	IndicatorsTotal      int `json:"indicators_total"`
	ProofsPassed         int `json:"proofs_passed"`
	ProofsTotal          int `json:"proofs_total"`
	TamperRejections     int `json:"tamper_rejections"`
	TamperTotal          int `json:"tamper_total"`
	PlatformReceipts     int `json:"platform_receipts"`
	OperatingSystems     int `json:"operating_systems"`
	ToolchainBindings    int `json:"toolchain_bindings"`
	RepositoryWrites     int `json:"repository_writes"`
	MutationAuthorities  int `json:"mutation_authorities"`
	StageOrderViolations int `json:"stage_order_violations"`
	EffectfulStages      int `json:"effectful_stages"`
	RegistryDrift        int `json:"registry_drift"`
}
