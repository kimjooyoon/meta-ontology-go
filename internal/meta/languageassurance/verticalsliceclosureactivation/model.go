package verticalsliceclosureactivation

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
	DenominatorTotal               int `json:"denominator_total"`
	BeforeOperating                int `json:"before_operating"`
	AfterOperating                 int `json:"after_operating"`
	BeforeCoverageBPS              int `json:"before_coverage_bps"`
	AfterCoverageBPS               int `json:"after_coverage_bps"`
	CapsulesTotal                  int `json:"capsules_total"`
	CapsulesExact                  int `json:"capsules_exact"`
	CapsuleCoverageBPS             int `json:"capsule_coverage_bps"`
	PredecessorSemanticsBPS        int `json:"predecessor_semantics_bps"`
	BoundariesTotal                int `json:"boundaries_total"`
	BoundariesSatisfied            int `json:"boundaries_satisfied"`
	LinksTotal                     int `json:"links_total"`
	LinksSatisfied                 int `json:"links_satisfied"`
	EligibilityIndicatorsTotal     int `json:"eligibility_indicators_total"`
	EligibilityIndicatorsSatisfied int `json:"eligibility_indicators_satisfied"`
	MetaOperationsRequired         int `json:"meta_operations_required"`
	MetaOperationsObserved         int `json:"meta_operations_observed"`
	UnknownPaths                   int `json:"unknown_paths"`
	BlockedPaths                   int `json:"blocked_paths"`
}
