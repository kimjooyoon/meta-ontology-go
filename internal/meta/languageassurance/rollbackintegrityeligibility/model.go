package rollbackintegrityeligibility

type ArtifactBinding struct {
	Name           string `json:"name"`
	ArtifactID     int64  `json:"artifact_id"`
	ArtifactDigest string `json:"artifact_digest"`
	CapsuleDigest  string `json:"capsule_digest"`
	ObservedDigest string `json:"observed_digest"`
	Exact          bool   `json:"exact"`
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
	DenominatorTotal         int `json:"denominator_total"`
	BeforeOperating          int `json:"before_operating"`
	AfterOperating           int `json:"after_operating"`
	BeforeCoverageBPS        int `json:"before_coverage_bps"`
	AfterCoverageBPS         int `json:"after_coverage_bps"`
	CapsulesTotal            int `json:"capsules_total"`
	CapsulesExact            int `json:"capsules_exact"`
	CapsuleCoverageBPS       int `json:"capsule_coverage_bps"`
	ShadowCasesTotal         int `json:"shadow_cases_total"`
	ShadowCasesPassed        int `json:"shadow_cases_passed"`
	ShadowReplaysTotal       int `json:"shadow_replays_total"`
	ShadowReplaysExact       int `json:"shadow_replays_exact"`
	ShadowReplayCoverageBPS  int `json:"shadow_replay_coverage_bps"`
	MetaOperationsRequired   int `json:"meta_operations_required"`
	MetaOperationsObserved   int `json:"meta_operations_observed"`
	MetaOperationCoverageBPS int `json:"meta_operation_coverage_bps"`
	EligiblePaths            int `json:"eligible_paths"`
	UnknownPaths             int `json:"unknown_paths"`
	BlockedPaths             int `json:"blocked_paths"`
}

type Definition struct {
	ID                 string `json:"id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedReason     string `json:"expected_reason"`
}

type CaseResult struct {
	Definition Definition `json:"definition"`
	Passed     bool       `json:"passed"`
	Report     Report     `json:"report"`
}

type Suite struct {
	Schema            string       `json:"schema"`
	SubjectSHA        string       `json:"subject_sha"`
	DenominatorID     string       `json:"denominator_id"`
	DenominatorDigest string       `json:"denominator_digest"`
	Decision          string       `json:"decision"`
	Resolution        string       `json:"resolution"`
	Cases             []CaseResult `json:"cases"`
	CasesTotal        int          `json:"cases_total"`
	CasesPassed       int          `json:"cases_passed"`
	CoverageBPS       int          `json:"coverage_bps"`
	SuiteDigest       string       `json:"suite_digest"`
}
