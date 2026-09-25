package assuranceeligibility

type Report struct {
	Schema                string            `json:"schema"`
	SubjectSHA            string            `json:"subject_sha"`
	AssuranceSubjectSHA   string            `json:"assurance_subject_sha"`
	Decision              string            `json:"decision"`
	Resolution            string            `json:"resolution"`
	EnforcementEffect     string            `json:"enforcement_effect"`
	Reason                string            `json:"reason"`
	Artifacts             []ArtifactBinding `json:"artifacts"`
	Transition            Transition        `json:"transition"`
	Summary               Summary           `json:"summary"`
	Indicators            []Indicator       `json:"indicators"`
	Proofs                []Proof           `json:"proofs"`
	RepositoryWrites      int               `json:"repository_writes"`
	OfficialMutationCount int               `json:"official_mutation_count"`
	PromotionApplied      int               `json:"promotion_applied"`
	ReportDigest          string            `json:"report_digest"`
}
