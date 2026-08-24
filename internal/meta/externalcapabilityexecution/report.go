package externalcapabilityexecution

type Report struct {
	Schema                   string       `json:"schema"`
	SubjectSHA               string       `json:"subject_sha"`
	Decision                 string       `json:"decision"`
	Resolution               string       `json:"resolution"`
	EnforcementEffect        string       `json:"enforcement_effect"`
	Reason                   string       `json:"reason"`
	Completed                int          `json:"completed"`
	Total                    int          `json:"total"`
	BasisPoints              int          `json:"basis_points"`
	UnknownIndicators        int          `json:"unknown_indicators"`
	DriverCompleted          int          `json:"driver_completed"`
	DriverTotal              int          `json:"driver_total"`
	OutcomeCompleted         int          `json:"outcome_completed"`
	OutcomeTotal             int          `json:"outcome_total"`
	GuardrailCompleted       int          `json:"guardrail_completed"`
	GuardrailTotal           int          `json:"guardrail_total"`
	ExternalExecutions       int          `json:"external_executions"`
	RepositoryWrites         int          `json:"repository_writes"`
	ExternalRepositoryWrites int          `json:"external_repository_writes"`
	OfficialMutationCount    int          `json:"official_mutation_count"`
	PromotionCount           int          `json:"promotion_count"`
	Parent                   ParentReport `json:"parent"`
	Indicators               []Indicator  `json:"indicators"`
	Proofs                   []Proof      `json:"proofs"`
	ObservationDigest        string       `json:"observation_digest"`
	ReportDigest             string       `json:"report_digest"`
}

type SuiteCase struct {
	CaseID             string `json:"case_id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ActualDecision     string `json:"actual_decision"`
	ActualResolution   string `json:"actual_resolution"`
	Passed             bool   `json:"passed"`
}
