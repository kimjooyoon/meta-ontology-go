package sourceauthoritypromotion

type upstreamDocument struct {
	Schema             string          `json:"schema"`
	SubjectSHA         string          `json:"subject_sha"`
	Decision           string          `json:"decision"`
	Resolution         string          `json:"resolution"`
	DenominatorID      string          `json:"denominator_id"`
	DenominatorDigest  string          `json:"denominator_digest"`
	RepositoryWrites   int             `json:"repository_writes"`
	PromotionCreditBPS int             `json:"promotion_credit_bps"`
	Summary            upstreamSummary `json:"summary"`
	Cases              []upstreamCase  `json:"cases"`
}

type upstreamSummary struct {
	CasesTotal  int `json:"cases_total"`
	CasesPassed int `json:"cases_passed"`
	ExactAllow  int `json:"exact_allow"`
	FailClosed  int `json:"fail_closed"`
	CoverageBPS int `json:"coverage_bps"`
}

type upstreamCase struct {
	ID                  string          `json:"id"`
	ExpectedObservation string          `json:"expected_observation"`
	ExpectedResolution  string          `json:"expected_resolution"`
	ExpectedEnforcement string          `json:"expected_enforcement"`
	ExpectedReason      string          `json:"expected_reason"`
	Passed              bool            `json:"passed"`
	Receipt             upstreamReceipt `json:"receipt"`
}

type upstreamReceipt struct {
	SubjectSHA         string              `json:"subject_sha"`
	Observation        string              `json:"observation"`
	Resolution         string              `json:"resolution"`
	Enforcement        string              `json:"enforcement"`
	Reason             string              `json:"reason"`
	RepositoryWrites   int                 `json:"repository_writes"`
	PromotionCreditBPS int                 `json:"promotion_credit_bps"`
	Snapshot           *upstreamSnapshot   `json:"snapshot"`
	Indicators         []upstreamIndicator `json:"indicators"`
}

type upstreamSnapshot struct {
	Digest       string            `json:"digest"`
	SourceRef    string            `json:"source_ref"`
	AuthorityRef string            `json:"authority_ref"`
	Bytes        int               `json:"bytes"`
	Authority    upstreamAuthority `json:"authority"`
	Selection    upstreamSelection `json:"selection"`
}

type upstreamAuthority struct {
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
	Path       string `json:"path"`
}

type upstreamSelection struct {
	StartLine int `json:"start_line"`
	EndLine   int `json:"end_line"`
}

type upstreamIndicator struct {
	Class       string `json:"class"`
	ProofChoice string `json:"proof_choice"`
	Satisfied   bool   `json:"satisfied"`
}
