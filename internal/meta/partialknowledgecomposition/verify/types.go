package verify

type Input struct {
	Repository string
	HeadSHA    string
	SourcePath string
	Source     []byte
	Fixture    []byte
	Receipt    []byte
}

type Report struct {
	Schema                 string `json:"schema"`
	Repository             string `json:"repository"`
	HeadSHA                string `json:"head_sha"`
	ReceiptDigest          string `json:"receipt_digest"`
	Status                 string `json:"status"`
	Decision               string `json:"decision"`
	FixedDenominator       int    `json:"fixed_denominator"`
	ExactCases             int    `json:"exact_cases"`
	DirectUnknownCases     int    `json:"direct_unknown_cases"`
	DependencyBlockedCases int    `json:"dependency_blocked_cases"`
	InvariantOnlyCases     int    `json:"invariant_only_cases"`
	MixedUnresolvedCases   int    `json:"mixed_unresolved_cases"`
	TopSuccessCases        int    `json:"top_success_cases"`
	NonExactNotPromoted    int    `json:"non_exact_not_promoted"`
	ClaimTransitionTotal   int    `json:"claim_transition_total"`
	RepositoryWrites       int    `json:"repository_writes"`
	PromotionAuthorized    bool   `json:"promotion_authorized"`
	IndependentEvaluator   bool   `json:"independent_evaluator"`
	Digest                 string `json:"digest"`
}

type fixture struct {
	Schema           string      `json:"schema"`
	SourcePath       string      `json:"source_path"`
	FixedDenominator int         `json:"fixed_denominator"`
	Cases            []caseInput `json:"cases"`
}

type caseInput struct {
	ID               string  `json:"id"`
	Producer         string  `json:"producer"`
	Consumer         string  `json:"consumer"`
	MetaOperation    string  `json:"meta_operation"`
	ProofChoice      string  `json:"proof_choice"`
	Left             operand `json:"left"`
	Right            operand `json:"right"`
	ExpectedState    string  `json:"expected_state"`
	ExpectedDecision string  `json:"expected_decision"`
	ExpectedReason   string  `json:"expected_reason"`
}

type operand struct {
	Operation         string   `json:"operation"`
	State             string   `json:"state"`
	BlockedDependency string   `json:"blocked_dependency,omitempty"`
	Invariants        []string `json:"invariants,omitempty"`
}

type value struct {
	State               string   `json:"state"`
	Contributors        []string `json:"contributors"`
	DirectUnknowns      []string `json:"direct_unknowns,omitempty"`
	BlockedDependencies []string `json:"blocked_dependencies,omitempty"`
	PreservedInvariants []string `json:"preserved_invariants,omitempty"`
}

type caseResult struct {
	ID             string  `json:"id"`
	Producer       string  `json:"producer"`
	Consumer       string  `json:"consumer"`
	MetaOperation  string  `json:"meta_operation"`
	ProofChoice    string  `json:"proof_choice"`
	Left           operand `json:"left"`
	Right          operand `json:"right"`
	Result         value   `json:"result"`
	Decision       string  `json:"decision"`
	Reason         string  `json:"reason"`
	TopSuccess     bool    `json:"top_success"`
	EvidenceDigest string  `json:"evidence_digest"`
}

type summary struct {
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

type indicator struct {
	ID                string `json:"id"`
	Class             string `json:"class"`
	ProofChoice       string `json:"proof_choice"`
	MetaOperation     string `json:"meta_operation"`
	Producer          string `json:"producer"`
	Consumer          string `json:"consumer"`
	Observed          int    `json:"observed"`
	Denominator       int    `json:"denominator"`
	BasisPoints       int    `json:"basis_points"`
	TargetBasisPoints int    `json:"target_basis_points"`
	Satisfied         bool   `json:"satisfied"`
}

type claimTransition struct {
	Sequence       int    `json:"sequence"`
	ClaimID        string `json:"claim_id"`
	From           string `json:"from"`
	To             string `json:"to"`
	MetaOperation  string `json:"meta_operation"`
	ProofChoice    string `json:"proof_choice"`
	EvidenceDigest string `json:"evidence_digest"`
	PreviousDigest string `json:"previous_digest,omitempty"`
	Digest         string `json:"digest"`
}

type receipt struct {
	Schema              string            `json:"schema"`
	Repository          string            `json:"repository"`
	HeadSHA             string            `json:"head_sha"`
	SourcePath          string            `json:"source_path"`
	SourceDigest        string            `json:"source_digest"`
	Producer            string            `json:"producer"`
	Consumer            string            `json:"consumer"`
	MetaOperation       string            `json:"meta_operation"`
	ProofChoice         string            `json:"proof_choice"`
	Resolution          string            `json:"resolution"`
	Decision            string            `json:"decision"`
	Reason              string            `json:"reason"`
	FixedDenominator    int               `json:"fixed_denominator"`
	DenominatorDigest   string            `json:"denominator_digest"`
	Cases               []caseResult      `json:"cases"`
	Claims              []claimTransition `json:"claims"`
	Indicators          []indicator       `json:"indicators"`
	Summary             summary           `json:"summary"`
	RepositoryWrites    int               `json:"repository_writes"`
	PromotionAuthorized bool              `json:"promotion_authorized"`
	Digest              string            `json:"digest"`
}
