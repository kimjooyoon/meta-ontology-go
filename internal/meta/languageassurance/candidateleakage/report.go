package candidateleakage

type Summary struct {
	BoundaryPaths         int `json:"boundary_paths"`
	LeakagePaths          int `json:"leakage_paths"`
	UnknownPaths          int `json:"unknown_paths"`
	BlockedPaths          int `json:"blocked_paths"`
	AuthorizedPaths       int `json:"authorized_paths"`
	BoundaryBindingBPS    int `json:"boundary_binding_bps"`
	PromotionAuthorityBPS int `json:"promotion_authority_bps"`
	RepositoryWrites      int `json:"repository_writes"`
	PromotionCreditBPS    int `json:"promotion_credit_bps"`
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

type Report struct {
	Schema            string          `json:"schema"`
	SubjectSHA        string          `json:"subject_sha"`
	Decision          string          `json:"decision"`
	Resolution        string          `json:"resolution"`
	EnforcementEffect string          `json:"enforcement_effect"`
	Reason            string          `json:"reason"`
	DenominatorID     string          `json:"denominator_id"`
	DenominatorDigest string          `json:"denominator_digest"`
	Input             Input           `json:"input"`
	Summary           Summary         `json:"summary"`
	Indicators        []Indicator     `json:"indicators"`
	MetaOperations    []MetaOperation `json:"meta_operations"`
	ReportDigest      string          `json:"report_digest"`
}

type CaseResult struct {
	Definition Definition `json:"definition"`
	Passed     bool       `json:"passed"`
	Report     Report     `json:"report"`
}

type SuiteSummary struct {
	CasesTotal          int `json:"cases_total"`
	CasesPassed         int `json:"cases_passed"`
	ExactPass           int `json:"exact_pass"`
	ExactFailClosed     int `json:"exact_fail_closed"`
	InvariantFailClosed int `json:"invariant_fail_closed"`
	CoverageBPS         int `json:"coverage_bps"`
}

type Suite struct {
	Schema             string       `json:"schema"`
	SubjectSHA         string       `json:"subject_sha"`
	DenominatorID      string       `json:"denominator_id"`
	DenominatorDigest  string       `json:"denominator_digest"`
	Decision           string       `json:"decision"`
	Resolution         string       `json:"resolution"`
	Cases              []CaseResult `json:"cases"`
	Summary            SuiteSummary `json:"summary"`
	RepositoryWrites   int          `json:"repository_writes"`
	PromotionCreditBPS int          `json:"promotion_credit_bps"`
	SuiteDigest        string       `json:"suite_digest"`
}
