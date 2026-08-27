package hygienicoriginidentity

const (
	SchemaVersion       = "gooo/hygienic-origin-identity/v1"
	Producer            = "gooo://hygienic-origin-identity/producer/name-generator"
	Consumer            = "gooo://hygienic-origin-identity/consumer/binding-site"
	MetaOperation       = "generate-name-preserving-origin-and-scope"
	ProofChoice         = "ORIGIN_SCOPE_EQUIVALENCE"
	DecisionPass        = "PASS"
	DecisionUnknown     = "UNKNOWN"
	ResolutionExact     = "EXACT"
	ResolutionLower     = "LOWER_RESOLUTION"
	StatusOpen          = "OPEN"
	StatusDischarged    = "DISCHARGED"
	StatusRefuted       = "REFUTED"
	CapturedOrigin      = "consumer-binding"
	CapturedScope       = "consumer-call-site"
	HygienicOrigin      = "producer-expansion-1"
	HygienicScope       = "fresh-producer-expansion-1"
	ConsumerBinding     = "gooo://hygienic-origin-identity/consumer/binding-site"
	ProducerExpansion   = "gooo://hygienic-origin-identity/producer/expansion-1"
	ConsumerCallSite    = "gooo://hygienic-origin-identity/scope/consumer-call-site"
	FreshProducerScope  = "gooo://hygienic-origin-identity/scope/fresh-producer-expansion-1"
	ExpectedCaseTotal   = 2
	ExpectedClaimTotal  = 4
	ExpectedUnknownPath = 1
)

type Report struct {
	SchemaVersion string    `json:"schema"`
	Decision      string    `json:"decision"`
	Resolution    string    `json:"resolution"`
	Producer      string    `json:"producer"`
	Consumer      string    `json:"consumer"`
	MetaOperation string    `json:"meta_operation"`
	ProofChoice   string    `json:"proof_choice"`
	Source        Subject   `json:"source"`
	Cases         []Case    `json:"cases"`
	Claims        []Claim   `json:"claims"`
	Unknowns      []Unknown `json:"unknowns"`
	Metrics       Metrics   `json:"metrics"`
	Authority     Authority `json:"authority"`
	ReceiptDigest string    `json:"receipt_digest"`
}

type Subject struct {
	Path    string `json:"path"`
	HeadSHA string `json:"head_sha"`
	Digest  string `json:"digest"`
}

type Case struct {
	ID                       string   `json:"id"`
	Label                    string   `json:"label"`
	Spelling                 string   `json:"spelling"`
	OriginIdentity           string   `json:"origin_identity"`
	ScopeProvenance          string   `json:"scope_provenance"`
	ResolvedIdentity         string   `json:"resolved_identity"`
	SameSpelling             bool     `json:"same_spelling"`
	Captured                 bool     `json:"captured"`
	OriginIdentityPreserved  bool     `json:"origin_identity_preserved"`
	ScopeProvenancePreserved bool     `json:"scope_provenance_preserved"`
	ClaimIDs                 []string `json:"claim_ids"`
}

type Claim struct {
	ID          string `json:"id"`
	CaseID      string `json:"case_id"`
	Proposition string `json:"proposition"`
	Status      string `json:"status"`
}

type Unknown struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type Metrics struct {
	FixedCaseDenominator        int `json:"fixed_case_denominator"`
	FixedClaimDenominator       int `json:"fixed_claim_denominator"`
	ObservedCaseTotal           int `json:"observed_case_total"`
	ObservedClaimTotal          int `json:"observed_claim_total"`
	SameSpellingCaseTotal       int `json:"same_spelling_case_total"`
	CapturedCaseTotal           int `json:"captured_case_total"`
	NonCapturedCaseTotal        int `json:"non_captured_case_total"`
	ClassifiedClaimTotal        int `json:"classified_claim_total"`
	DischargedClaimTotal        int `json:"discharged_claim_total"`
	RefutedClaimTotal           int `json:"refuted_claim_total"`
	OpenClaimTotal              int `json:"open_claim_total"`
	ClassificationCoverageBPS   int `json:"classification_coverage_bps"`
	PreservationSatisfactionBPS int `json:"preservation_satisfaction_bps"`
	UnknownPathTotal            int `json:"unknown_path_total"`
}

type Authority struct {
	RepositoryWrites             int  `json:"repository_writes"`
	RepositoryMutationAuthorized bool `json:"repository_mutation_authorized"`
}

type sourceCase struct {
	ID       string
	Spelling string
	Origin   string
	Scope    string
	Resolves string
	Expected string
}

type sourceDocument struct {
	Package    string
	Namespace  string
	Entities   map[string]bool
	Activities map[string]bool
	Cases      map[string]sourceCase
	Unknowns   []Unknown
}
