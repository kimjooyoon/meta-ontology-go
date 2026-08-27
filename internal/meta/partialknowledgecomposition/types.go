package partialknowledgecomposition

type Input struct {
	Repository string  `json:"repository"`
	HeadSHA    string  `json:"head_sha"`
	SourcePath string  `json:"source_path"`
	Source     []byte  `json:"-"`
	Fixture    Fixture `json:"fixture"`
}

type Fixture struct {
	Schema           string `json:"schema"`
	SourcePath       string `json:"source_path"`
	FixedDenominator int    `json:"fixed_denominator"`
	Cases            []Case `json:"cases"`
}

type Case struct {
	ID               string      `json:"id"`
	Producer         string      `json:"producer"`
	Consumer         string      `json:"consumer"`
	MetaOperation    string      `json:"meta_operation"`
	ProofChoice      ProofChoice `json:"proof_choice"`
	Left             Operand     `json:"left"`
	Right            Operand     `json:"right"`
	ExpectedState    State       `json:"expected_state"`
	ExpectedDecision string      `json:"expected_decision"`
	ExpectedReason   string      `json:"expected_reason"`
}

type Operand struct {
	Operation         string   `json:"operation"`
	State             State    `json:"state"`
	BlockedDependency string   `json:"blocked_dependency,omitempty"`
	Invariants        []string `json:"invariants,omitempty"`
}

type Value struct {
	State               State    `json:"state"`
	Contributors        []string `json:"contributors"`
	DirectUnknowns      []string `json:"direct_unknowns,omitempty"`
	BlockedDependencies []string `json:"blocked_dependencies,omitempty"`
	PreservedInvariants []string `json:"preserved_invariants,omitempty"`
}

type CaseResult struct {
	ID             string      `json:"id"`
	Producer       string      `json:"producer"`
	Consumer       string      `json:"consumer"`
	MetaOperation  string      `json:"meta_operation"`
	ProofChoice    ProofChoice `json:"proof_choice"`
	Left           Operand     `json:"left"`
	Right          Operand     `json:"right"`
	Result         Value       `json:"result"`
	Decision       string      `json:"decision"`
	Reason         string      `json:"reason"`
	TopSuccess     bool        `json:"top_success"`
	EvidenceDigest string      `json:"evidence_digest"`
}

type Summary struct {
	FixedDenominator       int `json:"fixed_denominator"`
	CaseTotal              int `json:"case_total"`
	ExactCases             int `json:"exact_cases"`
	DirectUnknownCases     int `json:"direct_unknown_cases"`
	DependencyBlockedCases int `json:"dependency_blocked_cases"`
	InvariantOnlyCases     int `json:"invariant_only_cases"`
	MixedUnresolvedCases   int `json:"mixed_unresolved_cases"`
	TopSuccessCases        int `json:"top_success_cases"`
	NonExactCases          int `json:"non_exact_cases"`
	NonExactNotPromoted    int `json:"non_exact_not_promoted"`
	ClaimTransitionTotal   int `json:"claim_transition_total"`
	RepositoryWrites       int `json:"repository_writes"`
}

type Indicator struct {
	ID                string      `json:"id"`
	Class             string      `json:"class"`
	ProofChoice       ProofChoice `json:"proof_choice"`
	MetaOperation     string      `json:"meta_operation"`
	Producer          string      `json:"producer"`
	Consumer          string      `json:"consumer"`
	Observed          int         `json:"observed"`
	Denominator       int         `json:"denominator"`
	BasisPoints       int         `json:"basis_points"`
	TargetBasisPoints int         `json:"target_basis_points"`
	Satisfied         bool        `json:"satisfied"`
}

type ClaimTransition struct {
	Sequence       int         `json:"sequence"`
	ClaimID        string      `json:"claim_id"`
	From           string      `json:"from"`
	To             string      `json:"to"`
	MetaOperation  string      `json:"meta_operation"`
	ProofChoice    ProofChoice `json:"proof_choice"`
	EvidenceDigest string      `json:"evidence_digest"`
	PreviousDigest string      `json:"previous_digest,omitempty"`
	Digest         string      `json:"digest"`
}

type Receipt struct {
	Schema              string            `json:"schema"`
	Repository          string            `json:"repository"`
	HeadSHA             string            `json:"head_sha"`
	SourcePath          string            `json:"source_path"`
	SourceDigest        string            `json:"source_digest"`
	Producer            string            `json:"producer"`
	Consumer            string            `json:"consumer"`
	MetaOperation       string            `json:"meta_operation"`
	ProofChoice         ProofChoice       `json:"proof_choice"`
	Resolution          string            `json:"resolution"`
	Decision            string            `json:"decision"`
	Reason              string            `json:"reason"`
	FixedDenominator    int               `json:"fixed_denominator"`
	DenominatorDigest   string            `json:"denominator_digest"`
	Cases               []CaseResult      `json:"cases"`
	Claims              []ClaimTransition `json:"claims"`
	Indicators          []Indicator       `json:"indicators"`
	Summary             Summary           `json:"summary"`
	RepositoryWrites    int               `json:"repository_writes"`
	PromotionAuthorized bool              `json:"promotion_authorized"`
	Digest              string            `json:"digest"`
}
