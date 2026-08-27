package causalci

// Input is the read-only observation supplied by a CI producer. It describes
// a source authority, a fixed check catalog, a persistent claim ledger, and a
// small scenario corpus. It contains no command strings and cannot authorize
// repository mutation.
type Input struct {
	Schema           string            `json:"schema"`
	SourcePath       string            `json:"source_path"`
	Operation        Operation         `json:"operation"`
	Policy           Policy            `json:"policy"`
	ClaimTransitions []ClaimTransition `json:"claim_transitions"`
	Cases            []Case            `json:"cases"`
}

type Operation struct {
	Producer          string `json:"producer"`
	Consumer          string `json:"consumer"`
	MetaOperation     string `json:"meta_operation"`
	ProofChoice       string `json:"proof_choice"`
	MutationAuthority bool   `json:"mutation_authority"`
	ReadOnly          bool   `json:"read_only"`
}

type Policy struct {
	Schema      string  `json:"schema"`
	Checks      []Check `json:"checks"`
	FullSuiteID string  `json:"full_suite_id"`
}

type Check struct {
	ID          string `json:"id"`
	Ordinal     int    `json:"ordinal"`
	Scope       string `json:"scope"`
	Description string `json:"description"`
}

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type ClaimTransition struct {
	Sequence       int        `json:"sequence"`
	ClaimID        string     `json:"claim_id"`
	Before         string     `json:"before"`
	After          string     `json:"after"`
	Event          string     `json:"event"`
	Coordinate     Coordinate `json:"coordinate"`
	EvidenceDigest string     `json:"evidence_digest"`
}

type Case struct {
	ID           string       `json:"id"`
	ChangedFiles []string     `json:"changed_files"`
	Claims       []Claim      `json:"claims"`
	ImpactEdges  []ImpactEdge `json:"impact_edges"`
}

type Claim struct {
	ID       string `json:"id"`
	Question string `json:"question"`
	State    string `json:"state"`
}

type ImpactEdge struct {
	ID         string     `json:"id"`
	From       string     `json:"from"`
	To         string     `json:"to"`
	Kind       string     `json:"kind"`
	Known      bool       `json:"known"`
	Reason     string     `json:"reason"`
	Coordinate Coordinate `json:"coordinate"`
}

type SourceEvidence struct {
	Path      string `json:"path"`
	Digest    string `json:"digest"`
	Authority string `json:"authority"`
}

type TransitionEvidence struct {
	Sequence       int        `json:"sequence"`
	ClaimID        string     `json:"claim_id"`
	Before         string     `json:"before"`
	After          string     `json:"after"`
	Event          string     `json:"event"`
	Coordinate     Coordinate `json:"coordinate"`
	EvidenceDigest string     `json:"evidence_digest"`
	PreviousDigest string     `json:"previous_digest"`
	Digest         string     `json:"digest"`
}

type PathEvidence struct {
	ChangedFile string   `json:"changed_file"`
	ClaimIDs    []string `json:"claim_ids"`
	CheckID     string   `json:"check_id"`
	EdgeIDs     []string `json:"edge_ids"`
	Explanation string   `json:"explanation"`
	ProofChoice string   `json:"proof_choice"`
}

type UnknownCause struct {
	ChangedFile string     `json:"changed_file"`
	EdgeID      string     `json:"edge_id"`
	Coordinate  Coordinate `json:"coordinate"`
	Reason      string     `json:"reason"`
}

type CheckChoice struct {
	CheckID     string   `json:"check_id"`
	ClaimIDs    []string `json:"claim_ids,omitempty"`
	PathIDs     []string `json:"path_ids,omitempty"`
	ProofChoice string   `json:"proof_choice"`
	Reason      string   `json:"reason"`
}

type CaseReceipt struct {
	ID             string         `json:"id"`
	ChangedFiles   []string       `json:"changed_files"`
	Decision       string         `json:"decision"`
	Resolution     string         `json:"resolution"`
	Reason         string         `json:"reason"`
	Coordinate     Coordinate     `json:"coordinate"`
	Paths          []PathEvidence `json:"paths"`
	UnknownCauses  []UnknownCause `json:"unknown_causes,omitempty"`
	SelectedChecks []CheckChoice  `json:"selected_checks"`
}

type Indicator struct {
	ID          string `json:"id"`
	Observed    int    `json:"observed"`
	Denominator int    `json:"denominator"`
	Satisfied   bool   `json:"satisfied"`
}

type Metrics struct {
	CaseTotal                 int `json:"case_total"`
	ChangedFileTotal          int `json:"changed_file_total"`
	ImpactEdgeTotal           int `json:"impact_edge_total"`
	KnownImpactEdgeTotal      int `json:"known_impact_edge_total"`
	SelectedCheckTotal        int `json:"selected_check_total"`
	FullFallbackCaseTotal     int `json:"full_fallback_case_total"`
	RejectedCaseTotal         int `json:"rejected_case_total"`
	ClaimTransitionTotal      int `json:"claim_transition_total"`
	FixedCheckDenominator     int `json:"fixed_check_denominator"`
	FixedIndicatorSatisfied   int `json:"fixed_indicator_satisfied"`
	FixedIndicatorDenominator int `json:"fixed_indicator_denominator"`
}

type IndependentVerifier struct {
	ID       string `json:"id"`
	Mode     string `json:"mode"`
	Required bool   `json:"required"`
	ReadOnly bool   `json:"read_only"`
}

type Receipt struct {
	Schema              string               `json:"schema"`
	Scope               string               `json:"scope"`
	Source              SourceEvidence       `json:"source"`
	InputDigest         string               `json:"input_digest"`
	Operation           Operation            `json:"operation"`
	PolicySchema        string               `json:"policy_schema"`
	FullSuiteID         string               `json:"full_suite_id"`
	ClaimTransitionHead string               `json:"claim_transition_head"`
	ClaimTransitions    []TransitionEvidence `json:"claim_transitions"`
	Cases               []CaseReceipt        `json:"cases"`
	Metrics             Metrics              `json:"metrics"`
	Indicators          []Indicator          `json:"indicators"`
	IndependentVerifier IndependentVerifier  `json:"independent_verifier"`
	Decision            string               `json:"decision"`
	Reason              string               `json:"reason"`
	Digest              string               `json:"digest"`
}
