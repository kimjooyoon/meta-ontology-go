package assuranceeligibility

type assuranceReport struct {
	Schema            string `json:"schema"`
	SubjectSHA        string `json:"subject_sha"`
	DenominatorID     string `json:"denominator_id"`
	DenominatorDigest string `json:"denominator_digest"`
	ReportDigest      string `json:"report_digest"`
	Summary           struct {
		DenominatorTotal           int `json:"denominator_total"`
		Operating                  int `json:"operating"`
		NotImplemented             int `json:"not_implemented"`
		ImplementationCoverageBPS  int `json:"implementation_coverage_bps"`
		EvidenceGroupsObserved     int `json:"evidence_groups_observed"`
		EvidenceGroupsTotal        int `json:"evidence_groups_total"`
		UnknownTopDecisions        int `json:"unknown_top_decisions"`
		SnapshotBindingsObserved   int `json:"snapshot_bindings_observed"`
		SnapshotBindingsRequired   int `json:"snapshot_bindings_required"`
		RawReconstructionsObserved int `json:"raw_reconstructions_observed"`
		RawReconstructionsRequired int `json:"raw_reconstructions_required"`
		UnresolvedIndicators       int `json:"unresolved_indicators"`
		ViolatedGuardrails         int `json:"violated_guardrails"`
		RepositoryWrites           int `json:"repository_writes"`
	} `json:"summary"`
	Obligations []struct {
		MetricID   string `json:"metric_id"`
		Status     string `json:"status"`
		Resolution string `json:"resolution"`
	} `json:"obligations"`
}

type parentReport struct {
	Schema                string `json:"schema"`
	ContractVersion       string `json:"contract_version"`
	Decision              string `json:"decision"`
	Resolution            string `json:"resolution"`
	Reason                string `json:"reason"`
	DenominatorVersion    string `json:"denominator_version"`
	DenominatorDigest     string `json:"denominator_digest"`
	Completed             int    `json:"completed"`
	Total                 int    `json:"total"`
	BasisPoints           int    `json:"basis_points"`
	ExternalExecutions    int    `json:"external_executions"`
	OfficialMutationCount int    `json:"official_mutation_count"`
	PromotionCount        int    `json:"promotion_count"`
	RepositoryWrites      int    `json:"repository_writes"`
	UnknownIndicators     int    `json:"unknown_indicators"`
}

type reference struct {
	Available    bool   `json:"available"`
	BindingExact bool   `json:"binding_exact"`
	URL          string `json:"url"`
	Commit       string `json:"commit"`
	Tree         string `json:"tree"`
}

type parentObservation struct {
	Schema                string    `json:"schema"`
	GoVersion             string    `json:"go_version"`
	Reference             reference `json:"reference"`
	OfficialMutationCount int       `json:"official_mutation_count"`
	PromotionCount        int       `json:"promotion_count"`
}

type parentSuite struct {
	Schema            string `json:"schema"`
	Passed            int    `json:"passed"`
	Total             int    `json:"total"`
	UnknownExpected   int    `json:"unknown_expected"`
	InvariantExpected int    `json:"invariant_expected"`
	Unresolved        int    `json:"unresolved"`
}

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
	UnknownEvents            int                 `json:"unknown_events"`
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

type evidence struct {
	Assurance             assuranceReport
	ParentReport          parentReport
	ParentObservation     parentObservation
	ParentSuite           parentSuite
	CapabilityReport      capabilityReport
	CapabilityObservation capabilityObservation
	CapabilitySuite       capabilitySuite
}
