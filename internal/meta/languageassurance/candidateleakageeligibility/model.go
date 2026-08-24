package candidateleakageeligibility

type ArtifactBinding struct {
	Name           string `json:"name"`
	ArtifactID     int64  `json:"artifact_id"`
	ArtifactDigest string `json:"artifact_digest"`
	CapsuleDigest  string `json:"capsule_digest"`
	ObservedDigest string `json:"observed_digest"`
	Exact          bool   `json:"exact"`
}

type Transition struct {
	MetricID          string `json:"metric_id"`
	MetaOperation     string `json:"meta_operation"`
	FromStatus        string `json:"from_status"`
	FromResolution    string `json:"from_resolution"`
	EligibleStatus    string `json:"eligible_status"`
	EligibleResolution string `json:"eligible_resolution"`
}

type Summary struct {
	DenominatorTotal  int `json:"denominator_total"`
	BeforeOperating   int `json:"before_operating"`
	AfterOperating    int `json:"after_operating"`
	BeforeCoverageBPS int `json:"before_coverage_bps"`
	AfterCoverageBPS  int `json:"after_coverage_bps"`
	CapsulesTotal     int `json:"capsules_total"`
	CapsulesExact     int `json:"capsules_exact"`
	CapsuleCoverageBPS int `json:"capsule_coverage_bps"`
	ShadowCasesTotal  int `json:"shadow_cases_total"`
	ShadowCasesPassed int `json:"shadow_cases_passed"`
	EligiblePaths     int `json:"eligible_paths"`
	UnknownPaths      int `json:"unknown_paths"`
	BlockedPaths      int `json:"blocked_paths"`
}

type Indicator struct {
	MetricID, Class, ProofChoice, Producer, Consumer string
	MetaOperation, Unit, Relation, Resolution        string
	Value, Target                                    int
	Satisfied                                        bool
}

type Report struct {
	Schema, SubjectSHA, EvidenceSubjectSHA string
	Decision, Resolution, EnforcementEffect, Reason string
	DenominatorID, DenominatorDigest string
	Artifacts []ArtifactBinding
	Transition Transition
	Summary Summary
	Indicators []Indicator
	MetaOperations []MetaOperationBinding
	RepositoryWrites int
	PromotionApplied int
	ReportDigest string
}

type Definition struct {
	ID, ExpectedDecision, ExpectedResolution, ExpectedReason string
}

type CaseResult struct {
	Definition Definition `json:"definition"`
	Passed bool `json:"passed"`
	Report Report `json:"report"`
}

type Suite struct {
	Schema, SubjectSHA, DenominatorID, DenominatorDigest string
	Decision, Resolution string
	Cases []CaseResult
	CasesTotal, CasesPassed, CoverageBPS int
	SuiteDigest string
}
