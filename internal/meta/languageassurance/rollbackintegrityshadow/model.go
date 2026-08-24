package rollbackintegrityshadow

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type CaseResult struct {
	Name               string `json:"name"`
	ExpectedDecision   string `json:"expected_decision"`
	ActualDecision     string `json:"actual_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ActualResolution   string `json:"actual_resolution"`
	ExpectedMode       string `json:"expected_mode"`
	ActualMode         string `json:"actual_mode"`
	Satisfied          int    `json:"satisfied"`
	CoordinatesTotal   int    `json:"coordinates_total"`
	Unresolved         int    `json:"unresolved"`
	NotSatisfied       int    `json:"not_satisfied"`
	RepositoryWrites   int    `json:"repository_writes"`
	ReportDigest       string `json:"report_digest"`
	ValidationError    string `json:"validation_error,omitempty"`
	Passed             bool   `json:"passed"`
}

type Summary struct {
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

type Report struct {
	Schema              string       `json:"schema"`
	MetricID            string       `json:"metric_id"`
	MetaOperation       string       `json:"meta_operation"`
	Decision            string       `json:"decision"`
	Reason              string       `json:"reason"`
	Resolution          string       `json:"resolution"`
	EnforcementEffect   string       `json:"enforcement_effect"`
	AssuranceSubjectSHA string       `json:"assurance_subject_sha,omitempty"`
	EvidenceDigest      string       `json:"evidence_digest"`
	Summary             Summary      `json:"summary"`
	Cases               []CaseResult `json:"cases"`
	Indicators          []Indicator  `json:"indicators"`
	RepositoryWrites    int          `json:"repository_writes"`
	PromotionApplied    int          `json:"promotion_applied"`
	ReportDigest        string       `json:"report_digest"`
}
