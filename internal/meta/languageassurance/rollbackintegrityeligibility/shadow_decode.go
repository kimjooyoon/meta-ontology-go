package rollbackintegrityeligibility

type shadowReportCapsule struct {
	Schema              string        `json:"schema"`
	MetricID            string        `json:"metric_id"`
	MetaOperation       string        `json:"meta_operation"`
	Decision            string        `json:"decision"`
	Resolution          string        `json:"resolution"`
	EnforcementEffect   string        `json:"enforcement_effect"`
	AssuranceSubjectSHA string        `json:"assurance_subject_sha"`
	EvidenceDigest      string        `json:"evidence_digest"`
	Summary             shadowSummary `json:"summary"`
	Cases               []shadowCase  `json:"cases"`
	RepositoryWrites    int           `json:"repository_writes"`
	PromotionApplied    int           `json:"promotion_applied"`
	ReportDigest        string        `json:"report_digest"`
}

type shadowSummary struct {
	DenominatorTotal     int `json:"denominator_total"`
	BeforeOperating      int `json:"before_operating"`
	ProjectedOperating   int `json:"projected_operating"`
	BeforeCoverageBPS    int `json:"before_coverage_bps"`
	ProjectedCoverageBPS int `json:"projected_coverage_bps"`
	CasesTotal           int `json:"cases_total"`
	CasesPassed          int `json:"cases_passed"`
	CaseCoverageBPS      int `json:"case_coverage_bps"`
	MetaReportsValid     int `json:"meta_reports_valid"`
	CoordinatesTotal     int `json:"coordinates_total"`
	TerminalCases        int `json:"terminal_cases"`
	UnknownDecisionCases int `json:"unknown_decision_cases"`
	KnownRejectCases     int `json:"known_reject_cases"`
}

type shadowCase struct {
	Name               string `json:"name"`
	ExpectedDecision   string `json:"expected_decision"`
	ActualDecision     string `json:"actual_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ActualResolution   string `json:"actual_resolution"`
	ExpectedMode       string `json:"expected_mode"`
	ActualMode         string `json:"actual_mode"`
	Unresolved         int    `json:"unresolved"`
	RepositoryWrites   int    `json:"repository_writes"`
	Passed             bool   `json:"passed"`
}
