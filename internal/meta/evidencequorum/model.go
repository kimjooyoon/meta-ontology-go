package evidencequorum

const (
	ContractSchema = "gooo/meta-evidence-quorum-contract/v1"
	ReceiptSchema  = "gooo/meta-evidence-quorum-receipt/v1"
	ReportSchema   = "gooo/meta-evidence-quorum-report/v1"
	Scope          = "GOOO_CLAIM_JUSTIFICATION_ONLY"

	DecisionPass    = "PASS"
	DecisionClosed  = "FAIL_CLOSED"
	DecisionUnknown = "UNKNOWN"

	ResolutionExact     = "EXACT"
	ResolutionLower     = "LOWER_RESOLUTION"
	ResolutionInvariant = "INVARIANT_ONLY"

	StatusDischarged = "DISCHARGED"
	StatusOpen       = "OPEN"
	StatusRefuted    = "REFUTED"
	StatusUnknown    = "UNKNOWN"
)

type Contract struct {
	Schema                   string           `json:"schema"`
	Scope                    string           `json:"scope"`
	SourcePath               string           `json:"source_path"`
	SourceEntry              string           `json:"source_entry"`
	FixedCaseDenominator     int              `json:"fixed_case_denominator"`
	MinimumIndependentGroups int              `json:"minimum_independent_groups"`
	RequiredRoles            []string         `json:"required_roles"`
	Claim                    ClaimDefinition  `json:"claim"`
	Cases                    []CaseDefinition `json:"cases"`
}

type ClaimDefinition struct {
	ID            string `json:"id"`
	Statement     string `json:"statement"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
}

type CaseDefinition struct {
	ID                 string `json:"id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedReason     string `json:"expected_reason"`
	ExpectedStatus     string `json:"expected_status"`
	ProducerDecision   string `json:"producer_decision"`
}

type Evidence struct {
	ID            string `json:"id"`
	ClaimID       string `json:"claim_id"`
	OriginGroup   string `json:"origin_group"`
	Role          string `json:"role"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	ProofChoice   string `json:"proof_choice"`
	Value         string `json:"value"`
	ConfidenceBPS int    `json:"confidence_bps"`
	SourcePath    string `json:"source_path"`
	SourceDigest  string `json:"source_digest"`
}

type Receipt struct {
	Schema            string     `json:"schema"`
	HeadSHA           string     `json:"head_sha"`
	SourcePath        string     `json:"source_path"`
	SourceDigest      string     `json:"source_digest"`
	Producer          string     `json:"producer"`
	Consumer          string     `json:"consumer"`
	MetaOperation     string     `json:"meta_operation"`
	ProofChoice       string     `json:"proof_choice"`
	Decision          string     `json:"decision"`
	Resolution        string     `json:"resolution"`
	Evidence          []Evidence `json:"evidence"`
	RepositoryWrites  int        `json:"repository_writes"`
	MutationAuthority bool       `json:"mutation_authority"`
	Digest            string     `json:"digest"`
}

type Input struct {
	Contract               Contract
	HeadSHA                string
	SourcePath             string
	Source                 []byte
	Receipts               [][]byte
	CaseReceipts           [][][]byte
	ProducerReceipt        []byte
	UnknownProducerReceipt []byte
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type ClaimTransition struct {
	From       string     `json:"from"`
	To         string     `json:"to"`
	Coordinate Coordinate `json:"coordinate"`
}

type ClaimResult struct {
	ID              string            `json:"id"`
	Producer        string            `json:"producer"`
	Consumer        string            `json:"consumer"`
	MetaOperation   string            `json:"meta_operation"`
	ProofChoice     string            `json:"proof_choice"`
	Status          string            `json:"status"`
	UnknownClass    string            `json:"unknown_class,omitempty"`
	Reason          string            `json:"reason"`
	Coordinate      Coordinate        `json:"coordinate"`
	EvidenceDigests []string          `json:"evidence_digests"`
	Transitions     []ClaimTransition `json:"transitions"`
}

type GroupResult struct {
	OriginGroup string   `json:"origin_group"`
	EvidenceIDs []string `json:"evidence_ids"`
	Roles       []string `json:"roles"`
	Values      []string `json:"values"`
	Independent bool     `json:"independent"`
}

type CaseResult struct {
	ID                 string        `json:"id"`
	Status             string        `json:"status"`
	ExpectedDecision   string        `json:"expected_decision"`
	ExpectedResolution string        `json:"expected_resolution"`
	ExpectedReason     string        `json:"expected_reason"`
	ObservedDecision   string        `json:"observed_decision"`
	ObservedResolution string        `json:"observed_resolution"`
	ObservedReason     string        `json:"observed_reason"`
	Coordinate         Coordinate    `json:"coordinate"`
	RawEvidence        int           `json:"raw_evidence"`
	IndependentGroups  int           `json:"independent_groups"`
	DuplicateEvidence  int           `json:"duplicate_evidence"`
	ConflictGroups     int           `json:"conflict_groups"`
	Groups             []GroupResult `json:"groups"`
	Claims             []ClaimResult `json:"claims"`
}

type Summary struct {
	CasesSatisfied           int  `json:"cases_satisfied"`
	CasesTotal               int  `json:"cases_total"`
	ClaimsTotal              int  `json:"claims_total"`
	DischargedClaims         int  `json:"discharged_claims"`
	OpenClaims               int  `json:"open_claims"`
	RefutedClaims            int  `json:"refuted_claims"`
	UnknownClaims            int  `json:"unknown_claims"`
	RawEvidenceTotal         int  `json:"raw_evidence_total"`
	IndependentGroupsTotal   int  `json:"independent_groups_total"`
	DuplicateEvidenceTotal   int  `json:"duplicate_evidence_total"`
	ConflictCases            int  `json:"conflict_cases"`
	QuorumSatisfiedCases     int  `json:"quorum_satisfied_cases"`
	LowerResolutionCases     int  `json:"lower_resolution_cases"`
	MinimumIndependentGroups int  `json:"minimum_independent_groups"`
	ConfidenceAggregated     bool `json:"confidence_aggregated"`
	RepositoryWrites         int  `json:"repository_writes"`
	MutationAuthority        bool `json:"mutation_authority"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type Proof struct {
	Choice         string `json:"choice"`
	MetaOperation  string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed         bool   `json:"passed"`
}

type Report struct {
	Schema                       string       `json:"schema"`
	Scope                        string       `json:"scope"`
	HeadSHA                      string       `json:"head_sha"`
	SourcePath                   string       `json:"source_path"`
	SourceEntry                  string       `json:"source_entry"`
	SourceDigest                 string       `json:"source_digest"`
	Decision                     string       `json:"decision"`
	Resolution                   string       `json:"resolution"`
	Reason                       string       `json:"reason"`
	ContractDigest               string       `json:"contract_digest"`
	ProducerReceiptDigest        string       `json:"producer_receipt_digest"`
	UnknownProducerReceiptDigest string       `json:"unknown_producer_receipt_digest"`
	ReceiptDigests               []string     `json:"receipt_digests"`
	Cases                        []CaseResult `json:"cases"`
	Summary                      Summary      `json:"summary"`
	Indicators                   []Indicator  `json:"indicators"`
	Proofs                       []Proof      `json:"proofs"`
	NotClaimed                   []string     `json:"not_claimed"`
	RepositoryWrites             int          `json:"repository_writes"`
	MutationAuthority            bool         `json:"mutation_authority"`
	Digest                       string       `json:"digest"`
}
