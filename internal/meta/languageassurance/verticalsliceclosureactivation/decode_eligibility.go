package verticalsliceclosureactivation

type eligibilityReport struct {
	Schema                     string                 `json:"schema"`
	SubjectSHA                 string                 `json:"subject_sha"`
	AssuranceSubjectSHA        string                 `json:"assurance_subject_sha"`
	ShadowEvidenceHead         string                 `json:"shadow_evidence_head"`
	Decision                   string                 `json:"decision"`
	Resolution                 string                 `json:"resolution"`
	EnforcementEffect          string                 `json:"enforcement_effect"`
	Reason                     string                 `json:"reason"`
	AssuranceDenominatorID     string                 `json:"assurance_denominator_id"`
	AssuranceDenominatorDigest string                 `json:"assurance_denominator_digest"`
	ShadowDenominatorDigest    string                 `json:"shadow_denominator_digest"`
	Artifacts                  []eligibilityArtifact  `json:"artifacts"`
	Transition                 eligibilityTransition `json:"transition"`
	Summary                    eligibilitySummary    `json:"summary"`
	Indicators                 []eligibilityIndicator `json:"indicators"`
	MetaOperations             []eligibilityOperation `json:"meta_operations"`
	RepositoryWrites           int                    `json:"repository_writes"`
	PromotionApplied           int                    `json:"promotion_applied"`
	ReportDigest               string                 `json:"report_digest"`
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
	OfficialStatus     string `json:"official_status"`
	OfficialResolution string `json:"official_resolution"`
}

type eligibilitySummary struct {
	DenominatorTotal          int `json:"denominator_total"`
	BeforeOperating           int `json:"before_operating"`
	EligibleOperating         int `json:"eligible_operating"`
	OfficialOperating         int `json:"official_operating"`
	BeforeCoverageBPS         int `json:"before_coverage_bps"`
	EligibleCoverageBPS       int `json:"eligible_coverage_bps"`
	OfficialCoverageBPS       int `json:"official_coverage_bps"`
	CapsulesTotal             int `json:"capsules_total"`
	CapsulesExact             int `json:"capsules_exact"`
	CapsuleCoverageBPS        int `json:"capsule_coverage_bps"`
	BoundariesTotal           int `json:"boundaries_total"`
	BoundariesSatisfied       int `json:"boundaries_satisfied"`
	LinksTotal                int `json:"links_total"`
	LinksSatisfied            int `json:"links_satisfied"`
	EligiblePaths             int `json:"eligible_paths"`
	UnknownPaths              int `json:"unknown_paths"`
	BlockedPaths              int `json:"blocked_paths"`
	ObservedRepositoryWrites  int `json:"observed_repository_writes"`
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
