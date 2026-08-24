package sourceauthoritypromotion

type Input struct {
	SubjectSHA    string
	AssuranceJSON []byte
	UpstreamJSON  []byte
}

type Report struct {
	Schema                       string      `json:"schema"`
	SubjectSHA                   string      `json:"subject_sha"`
	EligibilityDenominatorID     string      `json:"eligibility_denominator_id"`
	EligibilityDenominatorDigest string      `json:"eligibility_denominator_digest"`
	Baseline                     Baseline    `json:"baseline"`
	Evidence                     Evidence    `json:"evidence"`
	Transition                   Transition  `json:"transition"`
	Decision                     string      `json:"decision"`
	Resolution                   string      `json:"resolution"`
	Enforcement                  string      `json:"enforcement"`
	Reason                       string      `json:"reason"`
	Summary                      Summary     `json:"summary"`
	Indicators                   []Indicator `json:"indicators"`
	RepositoryWrites             int         `json:"repository_writes"`
	PromotionApplied             int         `json:"promotion_applied"`
	ReportDigest                 string      `json:"report_digest"`
}

type Baseline struct {
	DenominatorID     string `json:"denominator_id"`
	DenominatorDigest string `json:"denominator_digest"`
	Total             int    `json:"total"`
	Operating         int    `json:"operating"`
	NotImplemented    int    `json:"not_implemented"`
	CoverageBPS       int    `json:"coverage_bps"`
}

type Evidence struct {
	DenominatorID     string `json:"denominator_id"`
	DenominatorDigest string `json:"denominator_digest"`
	CasesPassed       int    `json:"cases_passed"`
	CasesTotal        int    `json:"cases_total"`
	CoverageBPS       int    `json:"coverage_bps"`
	Repository        string `json:"repository"`
	Revision          string `json:"revision"`
	SnapshotDigest    string `json:"snapshot_digest"`
}

type Transition struct {
	MetricID           string `json:"metric_id"`
	MetaOperation      string `json:"meta_operation"`
	FromStatus         string `json:"from_status"`
	FromResolution     string `json:"from_resolution"`
	EligibleStatus     string `json:"eligible_status"`
	EligibleResolution string `json:"eligible_resolution"`
}

type Summary struct {
	DenominatorTotal  int `json:"denominator_total"`
	BeforeOperating   int `json:"before_operating"`
	AfterOperating    int `json:"after_operating"`
	BeforeCoverageBPS int `json:"before_coverage_bps"`
	AfterCoverageBPS  int `json:"after_coverage_bps"`
	EligiblePaths     int `json:"eligible_paths"`
	BlockedPaths      int `json:"blocked_paths"`
}
