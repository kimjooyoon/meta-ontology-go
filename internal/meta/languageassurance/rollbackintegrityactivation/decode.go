package rollbackintegrityactivation

type assuranceReport struct {
	Schema              string                `json:"schema"`
	SubjectSHA          string                `json:"subject_sha"`
	DenominatorID       string                `json:"denominator_id"`
	AssuranceDecision   string                `json:"assurance_decision"`
	CandidateDecision   string                `json:"candidate_decision"`
	CandidateResolution string                `json:"candidate_resolution"`
	Summary             assuranceSummary      `json:"summary"`
	Obligations         []assuranceObligation `json:"obligations"`
	ReportDigest        string                `json:"report_digest"`
}

type assuranceSummary struct {
	DenominatorTotal          int `json:"denominator_total"`
	Operating                 int `json:"operating"`
	NotImplemented            int `json:"not_implemented"`
	ImplementationCoverageBPS int `json:"implementation_coverage_bps"`
	UnresolvedIndicators      int `json:"unresolved_indicators"`
	ViolatedGuardrails        int `json:"violated_guardrails"`
	RepositoryWrites          int `json:"repository_writes"`
}

type assuranceObligation struct {
	MetricID      string `json:"metric_id"`
	Status        string `json:"status"`
	Resolution    string `json:"resolution"`
	MetaOperation string `json:"meta_operation"`
}

type eligibilityReport struct {
	Schema             string                 `json:"schema"`
	SubjectSHA         string                 `json:"subject_sha"`
	EvidenceSubjectSHA string                 `json:"evidence_subject_sha"`
	Decision           string                 `json:"decision"`
	Resolution         string                 `json:"resolution"`
	EnforcementEffect  string                 `json:"enforcement_effect"`
	Reason             string                 `json:"reason"`
	DenominatorID      string                 `json:"denominator_id"`
	DenominatorDigest  string                 `json:"denominator_digest"`
	Artifacts          []eligibilityArtifact  `json:"artifacts"`
	Transition         eligibilityTransition  `json:"transition"`
	Summary            eligibilitySummary     `json:"summary"`
	Indicators         []eligibilityIndicator `json:"indicators"`
	MetaOperations     []eligibilityOperation `json:"meta_operations"`
	RepositoryWrites   int                    `json:"repository_writes"`
	PromotionApplied   int                    `json:"promotion_applied"`
	ReportDigest       string                 `json:"report_digest"`
}

type eligibilityArtifact struct {
	Exact bool `json:"exact"`
}

type eligibilityTransition struct {
	MetricID           string `json:"metric_id"`
	MetaOperation      string `json:"meta_operation"`
	FromStatus         string `json:"from_status"`
	FromResolution     string `json:"from_resolution"`
	EligibleStatus     string `json:"eligible_status"`
	EligibleResolution string `json:"eligible_resolution"`
}

type eligibilitySummary struct {
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

type eligibilityIndicator struct {
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Resolution    string `json:"resolution"`
	Satisfied     bool   `json:"satisfied"`
}

type eligibilityOperation struct {
	ID          string `json:"id"`
	ProofChoice string `json:"proof_choice"`
}
