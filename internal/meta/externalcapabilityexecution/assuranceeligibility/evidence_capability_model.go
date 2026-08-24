package assuranceeligibility

type parentBinding struct {
	Decision              string `json:"decision"`
	Resolution            string `json:"resolution"`
	Completed             int    `json:"completed"`
	Total                 int    `json:"total"`
	BasisPoints           int    `json:"basis_points"`
	OfficialMutationCount int    `json:"official_mutation_count"`
	PromotionCount        int    `json:"promotion_count"`
}

type capabilityReport struct {
	Schema                   string        `json:"schema"`
	SubjectSHA               string        `json:"subject_sha"`
	Decision                 string        `json:"decision"`
	Resolution               string        `json:"resolution"`
	EnforcementEffect        string        `json:"enforcement_effect"`
	Reason                   string        `json:"reason"`
	Completed                int           `json:"completed"`
	Total                    int           `json:"total"`
	BasisPoints              int           `json:"basis_points"`
	DriverCompleted          int           `json:"driver_completed"`
	DriverTotal              int           `json:"driver_total"`
	OutcomeCompleted         int           `json:"outcome_completed"`
	OutcomeTotal             int           `json:"outcome_total"`
	GuardrailCompleted       int           `json:"guardrail_completed"`
	GuardrailTotal           int           `json:"guardrail_total"`
	ExternalExecutions       int           `json:"external_executions"`
	ExternalRepositoryWrites int           `json:"external_repository_writes"`
	OfficialMutationCount    int           `json:"official_mutation_count"`
	PromotionCount           int           `json:"promotion_count"`
	RepositoryWrites         int           `json:"repository_writes"`
	UnknownIndicators        int           `json:"unknown_indicators"`
	ObservationDigest        string        `json:"observation_digest"`
	Parent                   parentBinding `json:"parent"`
}

type capabilityReference struct {
	RepositoryURL string `json:"repository_url"`
	CommitSHA     string `json:"commit_sha"`
	TreeSHA       string `json:"tree_sha"`
	GoVersion     string `json:"go_version"`
}

type capabilityObservation struct {
	Schema                   string              `json:"schema"`
	SubjectSHA               string              `json:"subject_sha"`
	ObservationDigest        string              `json:"observation_digest"`
	Available                bool                `json:"available"`
	ReplayExact              bool                `json:"replay_exact"`
	ExternalExecutions       int                 `json:"external_executions"`
	ExternalRepositoryWrites int                 `json:"external_repository_writes"`
	RepositoryWrites         int                 `json:"repository_writes"`
	OfficialMutationCount    int                 `json:"official_mutation_count"`
	PromotionCount           int                 `json:"promotion_count"`
	UnknownEvents            []string            `json:"unknown_events"`
	Reference                capabilityReference `json:"reference"`
}

type capabilitySuite struct {
	Schema            string `json:"schema"`
	SubjectSHA        string `json:"subject_sha"`
	Decision          string `json:"decision"`
	Resolution        string `json:"resolution"`
	Passed            int    `json:"passed"`
	Total             int    `json:"total"`
	CoverageBPS       int    `json:"coverage_bps"`
	ExactExpected     int    `json:"exact_expected"`
	UnknownExpected   int    `json:"unknown_expected"`
	InvariantExpected int    `json:"invariant_expected"`
	OfficialMutations int    `json:"official_mutations"`
	PromotionCount    int    `json:"promotion_count"`
	RepositoryWrites  int    `json:"repository_writes"`
}
