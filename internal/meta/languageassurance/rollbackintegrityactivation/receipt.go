package rollbackintegrityactivation

type Receipt struct {
	Schema                  string                 `json:"schema"`
	SubjectSHA              string                 `json:"subject_sha"`
	PredecessorSHA          string                 `json:"predecessor_sha"`
	Decision                string                 `json:"decision"`
	Resolution              string                 `json:"resolution"`
	EnforcementEffect       string                 `json:"enforcement_effect"`
	Reason                  string                 `json:"reason"`
	DenominatorID           string                 `json:"denominator_id"`
	DenominatorDigest       string                 `json:"denominator_digest"`
	EligibilityReportDigest string                 `json:"eligibility_report_digest"`
	Artifacts               []ArtifactBinding      `json:"artifacts"`
	Transition              Transition             `json:"transition"`
	Summary                 Summary                `json:"summary"`
	Indicators              []Indicator            `json:"indicators"`
	MetaOperations          []MetaOperationBinding `json:"meta_operations"`
	RepositoryWrites        int                    `json:"repository_writes"`
	TransitionApplied       int                    `json:"transition_applied"`
	ReportDigest            string                 `json:"report_digest"`
}
