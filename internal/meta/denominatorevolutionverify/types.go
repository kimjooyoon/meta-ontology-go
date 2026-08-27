package denominatorevolutionverify

const (
	ContractSchema     = "gooo/denominator-evolution-contract/v1"
	ReportSchema       = "gooo/denominator-evolution-report/v1"
	ReceiptSchema      = "gooo/denominator-migration-receipt/v1"
	VerificationSchema = "gooo/denominator-evolution-verification/v1"
	DenominatorVersion = "gooo/measurement-denominator/v1"
	DenominatorSize    = 5
	CaseCount          = 3
)

type Input struct {
	ContractRaw        []byte
	ReportRaw          []byte
	HeadSHA            string
	SourceRaw          []byte
	RepositorySnapshot RepositorySnapshot
}

type Contract struct {
	Schema      string          `json:"schema"`
	Version     int             `json:"version"`
	Producer    string          `json:"producer"`
	Consumer    string          `json:"consumer"`
	Denominator DenominatorSpec `json:"denominator"`
	Policy      Policy          `json:"policy"`
	Cases       []CaseSpec      `json:"cases"`
	NotClaimed  []string        `json:"not_claimed"`
}

type RepositorySnapshot struct {
	BeforeDigest string `json:"before_digest"`
	AfterDigest  string `json:"after_digest"`
	ChangedPaths int    `json:"changed_paths"`
}

type DenominatorSpec struct {
	Version     string       `json:"version"`
	Obligations []Obligation `json:"obligations"`
}

type Policy struct {
	NoAggregateEstimates   bool     `json:"no_aggregate_estimates"`
	ForbiddenClaims        []string `json:"forbidden_claims"`
	AllowedAdditionReasons []string `json:"allowed_addition_reasons"`
	AllowedDeletionReasons []string `json:"allowed_deletion_reasons"`
}

type Obligation struct {
	ID            string `json:"id"`
	Claim         string `json:"claim"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
}

type CaseSpec struct {
	ID                 string `json:"id"`
	Kind               string `json:"kind"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedReason     string `json:"expected_reason"`
	FromClaim          string `json:"from_claim"`
	ToClaim            string `json:"to_claim"`
	ProofChoice        string `json:"proof_choice"`
	MetaOperation      string `json:"meta_operation"`
	Stage              string `json:"stage"`
	Step               string `json:"step"`
	Reason             string `json:"reason"`
}

type Denominator struct {
	Version     string       `json:"version"`
	Obligations []Obligation `json:"obligations"`
	Digest      string       `json:"digest"`
}

type Ref struct {
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

type Change struct {
	ObligationID string `json:"obligation_id"`
	Reason       string `json:"reason"`
}

type Receipt struct {
	Schema             string             `json:"schema"`
	ID                 string             `json:"id"`
	Producer           string             `json:"producer"`
	Consumer           string             `json:"consumer"`
	Predecessor        Ref                `json:"predecessor"`
	Successor          Ref                `json:"successor"`
	Additions          []Change           `json:"additions"`
	Deletions          []Change           `json:"deletions"`
	Decision           string             `json:"decision"`
	Reason             string             `json:"reason"`
	Coordinate         Coordinate         `json:"coordinate"`
	RepositoryWrites   int                `json:"repository_writes"`
	MutationAuthority  bool               `json:"mutation_authority"`
	RepositorySnapshot RepositorySnapshot `json:"repository_snapshot"`
	Guardrails         []Guardrail        `json:"guardrails"`
	Digest             string             `json:"digest"`
}

type Guardrail struct {
	ID                     string `json:"id"`
	Direction              string `json:"direction"`
	Observed               int    `json:"observed"`
	AllowedMax             int    `json:"allowed_max"`
	ConformanceNumerator   int    `json:"conformance_numerator"`
	ConformanceDenominator int    `json:"conformance_denominator"`
	Conforms               bool   `json:"conforms"`
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type Check struct {
	ID            string     `json:"id"`
	Status        string     `json:"status"`
	ProofChoice   string     `json:"proof_choice"`
	MetaOperation string     `json:"meta_operation"`
	Coordinate    Coordinate `json:"coordinate"`
	Expected      string     `json:"expected"`
	Observed      string     `json:"observed"`
}

type Case struct {
	ID                 string      `json:"id"`
	Kind               string      `json:"kind"`
	Status             string      `json:"status"`
	ExpectedDecision   string      `json:"expected_decision"`
	ExpectedResolution string      `json:"expected_resolution"`
	ExpectedReason     string      `json:"expected_reason"`
	ObservedDecision   string      `json:"observed_decision"`
	ObservedResolution string      `json:"observed_resolution"`
	ObservedReason     string      `json:"observed_reason"`
	FromClaim          string      `json:"from_claim"`
	ToClaim            string      `json:"to_claim"`
	Predecessor        Denominator `json:"predecessor"`
	Successor          Denominator `json:"successor"`
	Receipt            *Receipt    `json:"receipt,omitempty"`
	Coordinate         Coordinate  `json:"coordinate"`
	Checks             []Check     `json:"checks"`
}

type Summary struct {
	CasesSatisfied                   int         `json:"cases_satisfied"`
	CasesTotal                       int         `json:"cases_total"`
	FixedDenominatorNumerator        int         `json:"fixed_denominator_numerator"`
	FixedDenominatorDenominator      int         `json:"fixed_denominator_denominator"`
	LegalAdvanceNumerator            int         `json:"legal_advance_numerator"`
	LegalAdvanceDenominator          int         `json:"legal_advance_denominator"`
	UnauthorizedRejectionNumerator   int         `json:"unauthorized_rejection_numerator"`
	UnauthorizedRejectionDenominator int         `json:"unauthorized_rejection_denominator"`
	UnknownPredecessorNumerator      int         `json:"unknown_predecessor_numerator"`
	UnknownPredecessorDenominator    int         `json:"unknown_predecessor_denominator"`
	AdditionReasonNumerator          int         `json:"addition_reason_numerator"`
	AdditionReasonDenominator        int         `json:"addition_reason_denominator"`
	DeletionReasonNumerator          int         `json:"deletion_reason_numerator"`
	DeletionReasonDenominator        int         `json:"deletion_reason_denominator"`
	SourceCasesNumerator             int         `json:"source_cases_numerator"`
	SourceCasesDenominator           int         `json:"source_cases_denominator"`
	PersistentClaimsNumerator        int         `json:"persistent_claims_numerator"`
	PersistentClaimsDenominator      int         `json:"persistent_claims_denominator"`
	GuardrailObservationsNumerator   int         `json:"guardrail_observations_numerator"`
	GuardrailObservationsDenominator int         `json:"guardrail_observations_denominator"`
	Guardrails                       []Guardrail `json:"guardrails"`
}

type Indicator struct {
	MetricID          string     `json:"metric_id"`
	Class             string     `json:"class"`
	ProofChoice       string     `json:"proof_choice"`
	MetaOperation     string     `json:"meta_operation"`
	Coordinate        Coordinate `json:"coordinate"`
	Numerator         int        `json:"numerator"`
	Denominator       int        `json:"denominator"`
	ExpectedNumerator int        `json:"expected_numerator"`
	Satisfied         bool       `json:"satisfied"`
	Guardrail         *Guardrail `json:"guardrail,omitempty"`
}

type Report struct {
	Schema             string             `json:"schema"`
	Scope              string             `json:"scope"`
	HeadSHA            string             `json:"head_sha"`
	Producer           string             `json:"producer"`
	Consumer           string             `json:"consumer"`
	Decision           string             `json:"decision"`
	Resolution         string             `json:"resolution"`
	Reason             string             `json:"reason"`
	ContractDigest     string             `json:"contract_digest"`
	SourceDigest       string             `json:"source_digest"`
	Denominator        Denominator        `json:"denominator"`
	SourceProjection   SourceProjection   `json:"source_projection"`
	Cases              []Case             `json:"cases"`
	Summary            Summary            `json:"summary"`
	Indicators         []Indicator        `json:"indicators"`
	NotClaimed         []string           `json:"not_claimed"`
	RepositoryWrites   int                `json:"repository_writes"`
	MutationAuthority  bool               `json:"mutation_authority"`
	RepositorySnapshot RepositorySnapshot `json:"repository_snapshot"`
	ClaimLedger        []ClaimLedgerEntry `json:"claim_ledger"`
	EmittedClaims      []EmittedClaim     `json:"emitted_claims"`
	Digest             string             `json:"digest"`
}

type SourceProjection struct {
	EntityCount        int      `json:"entity_count"`
	ActivityCount      int      `json:"activity_count"`
	ObligationCount    int      `json:"obligation_count"`
	CaseCount          int      `json:"case_count"`
	SemanticDigest     string   `json:"semantic_digest"`
	WireDigest         string   `json:"wire_digest"`
	RequiredEntities   []string `json:"required_entities"`
	RequiredActivities []string `json:"required_activities"`
	Exact              bool     `json:"exact"`
}

type ClaimLedgerEntry struct {
	Sequence       int    `json:"sequence"`
	ClaimID        string `json:"claim_id"`
	PriorState     string `json:"prior_state"`
	NextState      string `json:"next_state"`
	Stage          string `json:"stage"`
	Step           string `json:"step"`
	Reason         string `json:"reason"`
	PreviousDigest string `json:"previous_digest"`
	Digest         string `json:"digest"`
}

type EmittedClaim struct {
	ID    string `json:"id"`
	Class string `json:"class"`
	State string `json:"state"`
}

type Verification struct {
	Schema             string             `json:"schema"`
	HeadSHA            string             `json:"head_sha"`
	Decision           string             `json:"decision"`
	Resolution         string             `json:"resolution"`
	Reason             string             `json:"reason"`
	Producer           string             `json:"producer"`
	Consumer           string             `json:"consumer"`
	ContractDigest     string             `json:"contract_digest"`
	ReportDigest       string             `json:"report_digest"`
	SourceDigest       string             `json:"source_digest"`
	Checks             []Check            `json:"checks"`
	NotClaimed         []string           `json:"not_claimed"`
	Guardrails         []Guardrail        `json:"guardrails"`
	RepositoryWrites   int                `json:"repository_writes"`
	MutationAuthority  bool               `json:"mutation_authority"`
	RepositorySnapshot RepositorySnapshot `json:"repository_snapshot"`
	ClaimLedger        []ClaimLedgerEntry `json:"claim_ledger"`
	Digest             string             `json:"digest"`
}
