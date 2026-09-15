package externalecosystemconformance

type Summary struct {
	ReferenceDenominator    int `json:"reference_denominator"`
	BoundCapabilities       int `json:"bound_capabilities"`
	ImplementedCapabilities int `json:"implemented_capabilities"`
	CapabilityMappings      int `json:"capability_mappings"`
	MappingCoverageBPS      int `json:"mapping_coverage_bps"`
	DocumentsTotal          int `json:"documents_total"`
	DocumentsExact          int `json:"documents_exact"`
	CommitExact             int `json:"commit_exact"`
	ModuleExact             int `json:"module_exact"`
	UnknownPaths            int `json:"unknown_paths"`
	BlockedPaths            int `json:"blocked_paths"`
	LanguageDenominator     int `json:"language_denominator"`
	LanguageBeforeOperating int `json:"language_before_operating"`
	LanguageOfficialAfter   int `json:"language_official_after"`
	OfficialMutations       int `json:"official_mutations"`
	ObservedWrites          int `json:"observed_repository_writes"`
	ObservedExecutions      int `json:"observed_external_executions"`
}

type Report struct {
	Schema             string      `json:"schema"`
	SubjectSHA         string      `json:"subject_sha"`
	ReferenceID        string      `json:"reference_id"`
	Decision           string      `json:"decision"`
	Resolution         string      `json:"resolution"`
	EnforcementEffect  string      `json:"enforcement_effect"`
	Reason             string      `json:"reason"`
	Summary            Summary     `json:"summary"`
	Indicators         []Indicator `json:"indicators"`
	RepositoryWrites   int         `json:"repository_writes"`
	ExternalExecutions int         `json:"external_executions"`
	PromotionApplied   int         `json:"promotion_applied"`
	ReportDigest       string      `json:"report_digest"`
}

type SuiteCase struct {
	CaseID             string `json:"case_id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ActualDecision     string `json:"actual_decision"`
	ActualResolution   string `json:"actual_resolution"`
	Passed             bool   `json:"passed"`
}

type Suite struct {
	Schema              string      `json:"schema"`
	SubjectSHA          string      `json:"subject_sha"`
	Decision            string      `json:"decision"`
	Resolution          string      `json:"resolution"`
	CasesTotal          int         `json:"cases_total"`
	CasesPassed         int         `json:"cases_passed"`
	CoverageBPS         int         `json:"coverage_bps"`
	ReferenceBoundExact int         `json:"reference_bound_exact"`
	UnknownFailClosed   int         `json:"unknown_fail_closed"`
	InvariantFailClosed int         `json:"invariant_fail_closed"`
	RepositoryWrites    int         `json:"repository_writes"`
	ExternalExecutions  int         `json:"external_executions"`
	OfficialMutations   int         `json:"official_mutations"`
	Cases               []SuiteCase `json:"cases"`
	SuiteDigest         string      `json:"suite_digest"`
}
