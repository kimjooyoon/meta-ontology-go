package authorization

type Indicator struct {
	MetricID       string `json:"metric_id"`
	Class          string `json:"class"`
	ProofChoice    string `json:"proof_choice"`
	Stage          string `json:"stage"`
	MetaOperation  string `json:"meta_operation"`
	Value          int    `json:"value"`
	Target         int    `json:"target"`
	Status         string `json:"status"`
	UnknownReason  string `json:"unknown_reason,omitempty"`
	EvidenceDigest string `json:"evidence_digest,omitempty"`
}

type Claim struct {
	ClaimID        string `json:"claim_id"`
	Stage          string `json:"stage"`
	Statement      string `json:"statement"`
	Status         string `json:"status"`
	EvidenceDigest string `json:"evidence_digest,omitempty"`
}

type Proof struct {
	Mode       string `json:"mode"`
	Status     string `json:"status"`
	Completed  int    `json:"completed"`
	Total      int    `json:"total"`
	Resolution string `json:"resolution"`
}

type ReaderView struct {
	Reader      string `json:"reader"`
	Completed   int    `json:"completed"`
	Total       int    `json:"total"`
	BasisPoints int    `json:"basis_points"`
	Resolution  string `json:"resolution"`
}

type Receipt struct {
	Schema                      string            `json:"schema"`
	SubjectSHA                  string            `json:"subject_sha"`
	Decision                    string            `json:"decision"`
	Resolution                  string            `json:"resolution"`
	EnforcementEffect           string            `json:"enforcement_effect"`
	Reason                      string            `json:"reason"`
	Completed                   int               `json:"completed"`
	Total                       int               `json:"total"`
	BasisPoints                 int               `json:"basis_points"`
	UnknownIndicators           int               `json:"unknown_indicators"`
	OpenClaims                  int               `json:"open_claims"`
	DischargedClaims            int               `json:"discharged_claims"`
	RejectedClaims              int               `json:"rejected_claims"`
	RepositoryWrites            int               `json:"repository_writes"`
	OfficialMutationCount       int               `json:"official_mutation_count"`
	PromotionCount              int               `json:"promotion_count"`
	ExecutionAuthority          bool              `json:"execution_authority"`
	RepositoryMutationAuthority bool              `json:"repository_mutation_authority"`
	PromotionAuthority          bool              `json:"promotion_authority"`
	Indicators                  []Indicator       `json:"indicators"`
	Claims                      []Claim           `json:"claims"`
	Unknowns                    []UnknownEvidence `json:"unknowns"`
	Proofs                      []Proof           `json:"proofs"`
	ReaderViews                 []ReaderView      `json:"reader_views"`
	NonClaims                   []string          `json:"non_claims"`
	EnvelopeDigest              string            `json:"envelope_digest"`
	SourceReportDigest          string            `json:"source_report_digest"`
	PolicySourceDigest          string            `json:"policy_source_digest"`
	PolicyGeneratedDigest       string            `json:"policy_generated_digest"`
	ReceiptDigest               string            `json:"receipt_digest"`
}
