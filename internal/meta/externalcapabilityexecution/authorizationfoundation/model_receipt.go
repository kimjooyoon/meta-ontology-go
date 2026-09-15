package authorizationfoundation

type Receipt struct {
	Schema                      string             `json:"schema"`
	SubjectSHA                  string             `json:"subject_sha"`
	Decision                    string             `json:"decision"`
	Resolution                  string             `json:"resolution"`
	EnforcementEffect           string             `json:"enforcement_effect"`
	Reason                      string             `json:"reason"`
	Completed                   int                `json:"completed"`
	Total                       int                `json:"total"`
	BasisPoints                 int                `json:"basis_points"`
	UnknownIndicators           int                `json:"unknown_indicators"`
	OpenClaims                  int                `json:"open_claims"`
	DischargedClaims            int                `json:"discharged_claims"`
	RejectedClaims              int                `json:"rejected_claims"`
	RepositoryWrites            int                `json:"repository_writes"`
	OfficialMutationCount       int                `json:"official_mutation_count"`
	PromotionCount              int                `json:"promotion_count"`
	ExecutionAuthority          bool               `json:"execution_authority"`
	RepositoryMutationAuthority bool               `json:"repository_mutation_authority"`
	PromotionAuthority          bool               `json:"promotion_authority"`
	Indicators                  []Indicator        `json:"indicators"`
	Claims                      []Claim            `json:"claims"`
	Unknowns                    []Unknown          `json:"unknowns"`
	Proofs                      []Proof            `json:"proofs"`
	ReaderViews                 []ReaderView       `json:"reader_views"`
	NonClaims                   []string           `json:"non_claims"`
	EnvelopeDigest              string             `json:"envelope_digest"`
	SourceReportDigest          string             `json:"source_report_digest"`
	PolicySourceDigest          string             `json:"policy_source_digest"`
	PolicyGeneratedDigest       string             `json:"policy_generated_digest"`
	Foundation                  *FoundationBinding `json:"foundation,omitempty"`
	ReceiptDigest               string             `json:"receipt_digest"`
}
