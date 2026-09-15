package candidateleakageactivation

type Input struct {
	SubjectSHA  string
	Assurance   []byte
	Eligibility []byte
}

type ArtifactBinding struct {
	Name           string `json:"name"`
	ArtifactID     int64  `json:"artifact_id"`
	ArtifactDigest string `json:"artifact_digest"`
	CapsuleDigest  string `json:"capsule_digest"`
	ObservedDigest string `json:"observed_digest"`
	Bytes          int    `json:"bytes"`
	Exact          bool   `json:"exact"`
}

type Transition struct {
	MetricID       string `json:"metric_id"`
	MetaOperation  string `json:"meta_operation"`
	FromStatus     string `json:"from_status"`
	FromResolution string `json:"from_resolution"`
	ToStatus       string `json:"to_status"`
	ToResolution   string `json:"to_resolution"`
}

type Summary struct {
	DenominatorTotal        int `json:"denominator_total"`
	BeforeOperating         int `json:"before_operating"`
	AfterOperating          int `json:"after_operating"`
	BeforeCoverageBPS       int `json:"before_coverage_bps"`
	AfterCoverageBPS        int `json:"after_coverage_bps"`
	CapsulesTotal           int `json:"capsules_total"`
	CapsulesExact           int `json:"capsules_exact"`
	CapsuleCoverageBPS      int `json:"capsule_coverage_bps"`
	PredecessorSemanticsBPS int `json:"predecessor_semantics_bps"`
	UnknownPaths            int `json:"unknown_paths"`
	BlockedPaths            int `json:"blocked_paths"`
}

type Receipt struct {
	Schema                  string            `json:"schema"`
	SubjectSHA              string            `json:"subject_sha"`
	PredecessorSHA          string            `json:"predecessor_sha"`
	Decision                string            `json:"decision"`
	Resolution              string            `json:"resolution"`
	Reason                  string            `json:"reason"`
	DenominatorID           string            `json:"denominator_id"`
	DenominatorDigest       string            `json:"denominator_digest"`
	EligibilityReportDigest string            `json:"eligibility_report_digest"`
	Artifacts               []ArtifactBinding `json:"artifacts"`
	Transition              Transition        `json:"transition"`
	Summary                 Summary           `json:"summary"`
	Indicators              []Indicator       `json:"indicators"`
	RepositoryWrites        int               `json:"repository_writes"`
	TransitionApplied       int               `json:"transition_applied"`
	ReportDigest            string            `json:"report_digest"`
}
