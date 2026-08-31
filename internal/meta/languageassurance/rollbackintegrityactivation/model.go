package rollbackintegrityactivation

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
	DenominatorTotal         int `json:"denominator_total"`
	BeforeOperating          int `json:"before_operating"`
	AfterOperating           int `json:"after_operating"`
	BeforeCoverageBPS        int `json:"before_coverage_bps"`
	AfterCoverageBPS         int `json:"after_coverage_bps"`
	CapsulesTotal            int `json:"capsules_total"`
	CapsulesExact            int `json:"capsules_exact"`
	CapsuleCoverageBPS       int `json:"capsule_coverage_bps"`
	PredecessorSemanticsBPS  int `json:"predecessor_semantics_bps"`
	ShadowCasesTotal         int `json:"shadow_cases_total"`
	ShadowCasesPassed        int `json:"shadow_cases_passed"`
	ShadowReplaysTotal       int `json:"shadow_replays_total"`
	ShadowReplaysExact       int `json:"shadow_replays_exact"`
	MetaOperationsRequired   int `json:"meta_operations_required"`
	MetaOperationsObserved   int `json:"meta_operations_observed"`
	MetaOperationCoverageBPS int `json:"meta_operation_coverage_bps"`
	UnknownPaths             int `json:"unknown_paths"`
	BlockedPaths             int `json:"blocked_paths"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Resolution    string `json:"resolution"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type MetaOperationBinding struct {
	ID          string `json:"id"`
	ProofChoice string `json:"proof_choice"`
}
