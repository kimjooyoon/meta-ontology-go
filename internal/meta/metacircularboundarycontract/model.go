package metacircularboundarycontract

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type SourceObservation struct {
	Path              string   `json:"path"`
	SourceDigest      string   `json:"source_digest"`
	SemanticDigest    string   `json:"semantic_digest"`
	Package           string   `json:"package"`
	Namespace         string   `json:"namespace"`
	Entities          []string `json:"entities"`
	Activities        []string `json:"activities"`
	DescriptionBound  bool     `json:"description_bound"`
	ReadOnly          bool     `json:"read_only"`
	RepositoryWrites  int      `json:"repository_writes"`
	MutationAuthority bool     `json:"mutation_authority"`
}

type Capability struct {
	Issuer        string `json:"issuer"`
	SubjectDigest string `json:"subject_digest"`
	Operation     string `json:"operation"`
	Scope         string `json:"scope"`
	Handle        string `json:"handle"`
}

type Attempt struct {
	DescriptionDigest string      `json:"description_digest"`
	Capability        *Capability `json:"capability,omitempty"`
	RequestExecution  bool        `json:"request_execution"`
}

type CaseDefinition struct {
	ID                    string `json:"id"`
	ExpectedDecision      string `json:"expected_decision"`
	ExpectedAuthorization string `json:"expected_authorization"`
	ExpectedExecution     string `json:"expected_execution"`
	ExpectedReason        string `json:"expected_reason"`
	ProofChoice           string `json:"proof_choice"`
	MetaOperation         string `json:"meta_operation"`
}

type CaseObservation struct {
	Description          string `json:"description"`
	Authorization        string `json:"authorization"`
	Execution            string `json:"execution"`
	Reason               string `json:"reason"`
	DescriptionEscalated bool   `json:"description_escalated"`
	RepositoryWrites     int    `json:"repository_writes"`
	MutationAuthority    bool   `json:"mutation_authority"`
}

type ClaimTransition struct {
	ClaimID        string     `json:"claim_id"`
	Event          string     `json:"event"`
	Before         string     `json:"before"`
	After          string     `json:"after"`
	Coordinate     Coordinate `json:"coordinate"`
	EvidenceDigest string     `json:"evidence_digest,omitempty"`
}

type Receipt struct {
	Schema            string            `json:"schema"`
	CaseID            string            `json:"case_id"`
	Producer          string            `json:"producer"`
	Consumer          string            `json:"consumer"`
	MetaOperation     string            `json:"meta_operation"`
	ProofChoice       string            `json:"proof_choice"`
	Coordinate        Coordinate        `json:"coordinate"`
	SourceDigest      string            `json:"source_digest"`
	DescriptionDigest string            `json:"description_digest"`
	CapabilityDigest  string            `json:"capability_digest"`
	Decision          string            `json:"decision"`
	Authorization     string            `json:"authorization"`
	Execution         string            `json:"execution"`
	RepositoryWrites  int               `json:"repository_writes"`
	MutationAuthority bool              `json:"mutation_authority"`
	ClaimTransitions  []ClaimTransition `json:"claim_transitions"`
	ReceiptDigest     string            `json:"receipt_digest"`
}

type CaseResult struct {
	Definition  CaseDefinition  `json:"definition"`
	Attempt     Attempt         `json:"attempt"`
	Observation CaseObservation `json:"observation"`
	Receipt     Receipt         `json:"receipt"`
	Passed      bool            `json:"passed"`
}

type Indicator struct {
	MetricID      string `json:"metric_id"`
	Class         string `json:"class"`
	ProofChoice   string `json:"proof_choice"`
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	MetaOperation string `json:"meta_operation"`
	Unit          string `json:"unit"`
	Relation      string `json:"relation"`
	Value         int    `json:"value"`
	Target        int    `json:"target"`
	Satisfied     bool   `json:"satisfied"`
}

type Summary struct {
	CasesTotal                      int `json:"cases_total"`
	CasesPassed                     int `json:"cases_passed"`
	CaseCoverageBPS                 int `json:"case_coverage_bps"`
	DescriptionBound                int `json:"description_bound"`
	ExplicitAuthorizations          int `json:"explicit_authorizations"`
	AllowedExecutions               int `json:"allowed_executions"`
	DescriptionOnlyBlocked          int `json:"description_only_blocked"`
	ForgedAuthorizationsBlocked     int `json:"forged_authorizations_blocked"`
	OutOfScopeAuthorizationsBlocked int `json:"out_of_scope_authorizations_blocked"`
	DescriptionEscalationPaths      int `json:"description_escalation_paths"`
	ReplayMatches                   int `json:"replay_matches"`
	RepositoryWrites                int `json:"repository_writes"`
	MutationAuthority               int `json:"mutation_authority"`
}

type JudgeEvidence struct {
	Producer      string `json:"producer"`
	Consumer      string `json:"consumer"`
	ComparedCases int    `json:"compared_cases"`
	Mismatches    int    `json:"mismatches"`
	Decision      string `json:"decision"`
	Reason        string `json:"reason"`
}

type Report struct {
	Schema            string            `json:"schema"`
	Scope             string            `json:"scope"`
	HeadSHA           string            `json:"head_sha"`
	Source            SourceObservation `json:"source"`
	Decision          string            `json:"decision"`
	Resolution        string            `json:"resolution"`
	Reason            string            `json:"reason"`
	Coordinate        Coordinate        `json:"coordinate"`
	Cases             []CaseResult      `json:"cases"`
	Receipts          []Receipt         `json:"receipts"`
	Indicators        []Indicator       `json:"indicators"`
	MetaOperations    []MetaOperation   `json:"meta_operations"`
	Summary           Summary           `json:"summary"`
	IndependentJudge  JudgeEvidence     `json:"independent_judge"`
	RepositoryWrites  int               `json:"repository_writes"`
	MutationAuthority bool              `json:"mutation_authority"`
	MetaValue         string            `json:"meta_value"`
	NotClaimed        []string          `json:"not_claimed"`
	ReportDigest      string            `json:"report_digest"`
}

type MetaOperation struct {
	ID          string `json:"id"`
	Producer    string `json:"producer"`
	Consumer    string `json:"consumer"`
	ProofChoice string `json:"proof_choice"`
}

type Denominator struct {
	ID     string           `json:"id"`
	Cases  []CaseDefinition `json:"cases"`
	Digest string           `json:"digest"`
}

type Input struct {
	Path    string
	HeadSHA string
	Source  []byte
}

type CausalityCase struct {
	ID                       string `json:"id"`
	Kind                     string `json:"kind"`
	BaselineSourceDigest     string `json:"baseline_source_digest"`
	IntervenedSourceDigest   string `json:"intervened_source_digest"`
	BaselineSemanticDigest   string `json:"baseline_semantic_digest"`
	IntervenedSemanticDigest string `json:"intervened_semantic_digest"`
	SourceChanged            bool   `json:"source_changed"`
	SemanticChanged          bool   `json:"semantic_changed"`
	ExpectedSemanticChange   bool   `json:"expected_semantic_change"`
	ConsumerAccepted         bool   `json:"consumer_accepted"`
	Passed                   bool   `json:"passed"`
}

type CausalitySummary struct {
	Total       int `json:"total"`
	Passed      int `json:"passed"`
	CoverageBPS int `json:"coverage_bps"`
}

type CausalityReport struct {
	Schema  string           `json:"schema"`
	Cases   []CausalityCase  `json:"cases"`
	Summary CausalitySummary `json:"summary"`
}
