package verify

type Input struct {
	Repository       string
	HeadSHA          string
	SourcePath       string
	Source           []byte
	InterventionMode string
	Receipt          []byte
}

type Evidence struct {
	Operation         string `json:"operation"`
	Required          string `json:"required"`
	Observed          string `json:"observed"`
	ObservedAvailable bool   `json:"observed_available"`
	DependencyClaimID string `json:"dependency_claim_id,omitempty"`
	PriorState        string `json:"prior_state"`
	InvariantEvidence string `json:"invariant_evidence,omitempty"`
}

type Case struct {
	ID               string   `json:"id"`
	SourceActivity   string   `json:"source_activity"`
	SourceActivityID string   `json:"source_activity_id"`
	Producer         string   `json:"producer"`
	Consumer         string   `json:"consumer"`
	MetaOperation    string   `json:"meta_operation"`
	ProofChoice      string   `json:"proof_choice"`
	Left             Evidence `json:"left"`
	Right            Evidence `json:"right"`
}

type Value struct {
	State               string   `json:"state"`
	Contributors        []string `json:"contributors"`
	DirectUnknowns      []string `json:"direct_unknowns,omitempty"`
	BlockedDependencies []string `json:"blocked_dependencies,omitempty"`
	PreservedInvariants []string `json:"preserved_invariants,omitempty"`
}

type Provenance struct {
	SourcePath        string `json:"source_path"`
	SourceActivity    string `json:"source_activity"`
	Producer          string `json:"producer"`
	Consumer          string `json:"consumer"`
	MetaOperation     string `json:"meta_operation"`
	ProofChoice       string `json:"proof_choice"`
	SemanticIRDigest  string `json:"semantic_ir_digest"`
	ObservationDigest string `json:"observation_digest"`
}

type CaseResult struct {
	ID               string     `json:"id"`
	SourceActivity   string     `json:"source_activity"`
	SourceActivityID string     `json:"source_activity_id"`
	Producer         string     `json:"producer"`
	Consumer         string     `json:"consumer"`
	MetaOperation    string     `json:"meta_operation"`
	ProofChoice      string     `json:"proof_choice"`
	Left             Evidence   `json:"left"`
	Right            Evidence   `json:"right"`
	Result           Value      `json:"result"`
	Decision         string     `json:"decision"`
	Resolution       string     `json:"resolution"`
	Stage            string     `json:"stage"`
	Step             string     `json:"step"`
	Reason           string     `json:"reason"`
	TopSuccess       bool       `json:"top_success"`
	Provenance       Provenance `json:"provenance"`
	EvidenceDigest   string     `json:"evidence_digest"`
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
	OpenClaims             int `json:"open_claims"`
	DischargedClaims       int `json:"discharged_claims"`
	ClaimTransitionTotal   int `json:"claim_transition_total"`
	RepositoryWrites       int `json:"repository_writes"`
}

type Indicator struct {
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

type ClaimTransition struct {
	Sequence       int        `json:"sequence"`
	ClaimID        string     `json:"claim_id"`
	From           string     `json:"from"`
	To             string     `json:"to"`
	Stage          string     `json:"stage"`
	Step           string     `json:"step"`
	Reason         string     `json:"reason"`
	Provenance     Provenance `json:"provenance"`
	EvidenceDigest string     `json:"evidence_digest"`
	PreviousDigest string     `json:"previous_digest,omitempty"`
	Digest         string     `json:"digest"`
}

type Intervention struct {
	Mode     string `json:"mode"`
	Semantic bool   `json:"semantic"`
	Target   string `json:"target,omitempty"`
	From     string `json:"from,omitempty"`
	To       string `json:"to,omitempty"`
	Comment  string `json:"comment,omitempty"`
}

type Receipt struct {
	Schema                   string            `json:"schema"`
	Repository               string            `json:"repository"`
	HeadSHA                  string            `json:"head_sha"`
	SourcePath               string            `json:"source_path"`
	SourceDigest             string            `json:"source_digest"`
	SemanticIRDigest         string            `json:"semantic_ir_digest"`
	SourceCases              int               `json:"source_cases"`
	SourceCasesTotal         int               `json:"source_cases_total"`
	Producer                 string            `json:"producer"`
	Consumer                 string            `json:"consumer"`
	MetaOperation            string            `json:"meta_operation"`
	ProofChoice              string            `json:"proof_choice"`
	Resolution               string            `json:"resolution"`
	SubjectResolution        string            `json:"subject_resolution"`
	Decision                 string            `json:"decision"`
	Reason                   string            `json:"reason"`
	FixedDenominator         int               `json:"fixed_denominator"`
	Cases                    []CaseResult      `json:"cases"`
	Claims                   []ClaimTransition `json:"claims"`
	Indicators               []Indicator       `json:"indicators"`
	Summary                  Summary           `json:"summary"`
	Intervention             Intervention      `json:"intervention"`
	RepositoryWrites         int               `json:"repository_writes"`
	PromotionAuthorized      bool              `json:"promotion_authorized"`
	SemanticProjectionDigest string            `json:"semantic_projection_digest"`
	Digest                   string            `json:"digest"`
}

type Report struct {
	Schema                   string `json:"schema"`
	Repository               string `json:"repository"`
	HeadSHA                  string `json:"head_sha"`
	Status                   string `json:"status"`
	Decision                 string `json:"decision"`
	Resolution               string `json:"resolution"`
	SubjectResolution        string `json:"subject_resolution"`
	FixedDenominator         int    `json:"fixed_denominator"`
	SourceCases              int    `json:"source_cases"`
	SourceCasesTotal         int    `json:"source_cases_total"`
	ExactCases               int    `json:"exact_cases"`
	DirectUnknownCases       int    `json:"direct_unknown_cases"`
	DependencyBlockedCases   int    `json:"dependency_blocked_cases"`
	InvariantOnlyCases       int    `json:"invariant_only_cases"`
	MixedUnresolvedCases     int    `json:"mixed_unresolved_cases"`
	TopSuccessCases          int    `json:"top_success_cases"`
	NonExactNotPromoted      int    `json:"non_exact_not_promoted"`
	OpenClaims               int    `json:"open_claims"`
	ClaimTransitionTotal     int    `json:"claim_transition_total"`
	RepositoryWrites         int    `json:"repository_writes"`
	PromotionAuthorized      bool   `json:"promotion_authorized"`
	IndependentEvaluator     bool   `json:"independent_evaluator"`
	SourceSemanticDigest     string `json:"source_semantic_digest"`
	ReceiptDigest            string `json:"receipt_digest"`
	SemanticProjectionDigest string `json:"semantic_projection_digest"`
	Digest                   string `json:"digest"`
}
