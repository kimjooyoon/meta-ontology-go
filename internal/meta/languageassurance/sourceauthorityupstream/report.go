package sourceauthorityupstream

type Snapshot struct {
	SourceRef    string    `json:"source_ref"`
	AuthorityRef string    `json:"authority_ref"`
	URL          string    `json:"url"`
	Authority    Authority `json:"authority"`
	Selection    Selection `json:"selection"`
	Digest       string    `json:"digest"`
	Bytes        int       `json:"bytes"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	Resolution    string `json:"resolution"`
	Unit          string `json:"unit"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Relation      string `json:"relation"`
	Satisfied     bool   `json:"satisfied"`
}

type Receipt struct {
	Schema             string      `json:"schema"`
	SubjectSHA         string      `json:"subject_sha"`
	RequestDigest      string      `json:"request_digest"`
	Mode               string      `json:"mode"`
	Observation        string      `json:"observation"`
	Resolution         string      `json:"resolution"`
	Enforcement        string      `json:"enforcement"`
	Reason             string      `json:"reason"`
	Snapshot           *Snapshot   `json:"snapshot,omitempty"`
	Indicators         []Indicator `json:"indicators"`
	RepositoryWrites   int         `json:"repository_writes"`
	PromotionCreditBPS int         `json:"promotion_credit_bps"`
	ReceiptDigest      string      `json:"receipt_digest"`
}

type CaseResult struct {
	ID                  string  `json:"id"`
	ExpectedObservation string  `json:"expected_observation"`
	ExpectedResolution  string  `json:"expected_resolution"`
	ExpectedEnforcement string  `json:"expected_enforcement"`
	ExpectedReason      string  `json:"expected_reason"`
	Passed              bool    `json:"passed"`
	Receipt             Receipt `json:"receipt"`
}

type SuiteSummary struct {
	CasesTotal  int `json:"cases_total"`
	CasesPassed int `json:"cases_passed"`
	ExactAllow  int `json:"exact_allow"`
	FailClosed  int `json:"fail_closed"`
	CoverageBPS int `json:"coverage_bps"`
}

type Suite struct {
	Schema             string       `json:"schema"`
	SubjectSHA         string       `json:"subject_sha"`
	DenominatorID      string       `json:"denominator_id"`
	DenominatorDigest  string       `json:"denominator_digest"`
	Decision           string       `json:"decision"`
	Resolution         string       `json:"resolution"`
	Reason             string       `json:"reason"`
	Cases              []CaseResult `json:"cases"`
	Summary            SuiteSummary `json:"summary"`
	RepositoryWrites   int          `json:"repository_writes"`
	PromotionCreditBPS int          `json:"promotion_credit_bps"`
	SuiteDigest        string       `json:"suite_digest"`
}
