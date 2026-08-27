package languagesourcebindingpromotion

type ClaimResult struct {
	ID              string     `json:"id"`
	Status          string     `json:"status"`
	UnknownClass    string     `json:"unknown_class,omitempty"`
	Reason          string     `json:"reason"`
	Coordinate      Coordinate `json:"coordinate"`
	EvidenceDigests []string   `json:"evidence_digests"`
}

type CaseResult struct {
	ID                 string        `json:"id"`
	Status             string        `json:"status"`
	ExpectedDecision   string        `json:"expected_decision"`
	ObservedDecision   string        `json:"observed_decision"`
	ObservedResolution string        `json:"observed_resolution"`
	ObservedReason     string        `json:"observed_reason"`
	Coordinate         Coordinate    `json:"coordinate"`
	Claims             []ClaimResult `json:"claims"`
}

type Summary struct {
	CasesSatisfied       int `json:"cases_satisfied"`
	CasesTotal           int `json:"cases_total"`
	ExactPromotions      int `json:"exact_promotions"`
	ExactClaims          int `json:"exact_claims"`
	DirectUnknowns       int `json:"direct_unknowns"`
	DependencyBlocked    int `json:"dependency_blocked"`
	LinkRefutations      int `json:"link_refutations"`
	PolicyReplays        int `json:"policy_replays"`
	ProducerDependencies int `json:"producer_dependencies"`
	SemanticClaims       int `json:"semantic_correctness_claims"`
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
	Choice        string `json:"choice"`
	MetaOperation string `json:"meta_operation"`
	EvidenceDigest string `json:"evidence_digest"`
	Passed        bool   `json:"passed"`
}
