package externalconformanceactivation

type eligibilityReport struct {
	Schema                string                 `json:"schema"`
	SubjectSHA            string                 `json:"subject_sha"`
	AssuranceSubjectSHA   string                 `json:"assurance_subject_sha"`
	Decision              string                 `json:"decision"`
	Resolution            string                 `json:"resolution"`
	EnforcementEffect     string                 `json:"enforcement_effect"`
	Reason                string                 `json:"reason"`
	Artifacts             []eligibilityArtifact  `json:"artifacts"`
	Transition            eligibilityTransition  `json:"transition"`
	Summary               eligibilitySummary     `json:"summary"`
	Indicators            []eligibilityIndicator `json:"indicators"`
	Proofs                []eligibilityProof     `json:"proofs"`
	RepositoryWrites      int                    `json:"repository_writes"`
	OfficialMutationCount int                    `json:"official_mutation_count"`
	PromotionApplied      int                    `json:"promotion_applied"`
	ReportDigest          string                 `json:"report_digest"`
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

type eligibilityProof struct {
	Choice    string `json:"choice"`
	Status    string `json:"status"`
	Satisfied int    `json:"satisfied"`
	Total     int    `json:"total"`
}
