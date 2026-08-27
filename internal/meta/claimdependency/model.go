package claimdependency

const (
	ReceiptSchema = "gooo.meta.claim-dependency-receipt/v1"
	GraphSchema   = "gooo.meta.claim-dependency-graph/v1"
	Scope         = "CLAIM_STATE_PROPAGATION_ONLY"

	CaseDirectUnknown = "direct-unknown"
	CaseRefuted       = "refuted"
	CaseRecovered     = "recovered"

	ClaimTotal      = 6
	EdgeTotal       = 8
	TransitionTotal = ClaimTotal * 2
)

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type Claim struct {
	Ordinal       int        `json:"ordinal"`
	Axis          string     `json:"axis"`
	ClaimID       string     `json:"claim_id"`
	Statement     string     `json:"statement"`
	Producer      string     `json:"producer"`
	Consumer      string     `json:"consumer"`
	MetaOperation string     `json:"meta_operation"`
	ProofChoice   string     `json:"proof_choice"`
	Coordinate    Coordinate `json:"coordinate"`
}

type Edge struct {
	EdgeID      string `json:"edge_id"`
	FromClaimID string `json:"from_claim_id"`
	ToClaimID   string `json:"to_claim_id"`
	Kind        string `json:"kind"`
}

type Graph struct {
	Schema       string  `json:"schema"`
	Authority    string  `json:"authority"`
	Completeness string  `json:"completeness"`
	NodeTotal    int     `json:"node_total"`
	EdgeTotal    int     `json:"edge_total"`
	Nodes        []Claim `json:"nodes"`
	Edges        []Edge  `json:"edges"`
	Digest       string  `json:"digest"`
}

type Subject struct {
	Case             string `json:"case"`
	SourcePath       string `json:"source_path"`
	SourceDigest     string `json:"source_digest"`
	Producer         string `json:"producer"`
	Consumer         string `json:"consumer"`
	MetaOperation    string `json:"meta_operation"`
	ProofChoice      string `json:"proof_choice"`
	ReadOnly         bool   `json:"read_only"`
	RepositoryWrites int    `json:"repository_writes"`
	RecoveryFromCase string `json:"recovery_from_case,omitempty"`
}

type Transition struct {
	Sequence                 int        `json:"sequence"`
	ClaimID                  string     `json:"claim_id"`
	Event                    string     `json:"event"`
	Before                   string     `json:"before"`
	After                    string     `json:"after"`
	Coordinate               Coordinate `json:"coordinate"`
	EvidenceDigest           string     `json:"evidence_digest,omitempty"`
	PreviousTransitionDigest string     `json:"previous_transition_digest,omitempty"`
	TransitionDigest         string     `json:"transition_digest"`
}

type Resolution struct {
	ClaimID               string      `json:"claim_id"`
	Axis                  string      `json:"axis"`
	State                 string      `json:"state"`
	Kind                  string      `json:"kind"`
	ObservedEvent         string      `json:"observed_event"`
	Coordinate            Coordinate  `json:"coordinate"`
	FailureResponsibility string      `json:"failure_responsibility"`
	FailureOwnerClaimID   string      `json:"failure_owner_claim_id"`
	MissingEvidenceIDs    []string    `json:"missing_evidence_ids,omitempty"`
	BlockedByClaimIDs     []string    `json:"blocked_by_claim_ids,omitempty"`
	BlockedByEdgeIDs      []string    `json:"blocked_by_edge_ids,omitempty"`
	CausePath             []string    `json:"cause_path"`
	CauseEdgeIDs          []string    `json:"cause_edge_ids"`
	CauseTransitionDigest string      `json:"cause_transition_digest"`
	CauseCoordinate       *Coordinate `json:"cause_coordinate"`
}

type Metrics struct {
	FixedClaimTotal             int `json:"fixed_claim_total"`
	FixedEdgeTotal              int `json:"fixed_edge_total"`
	ClassifiedClaimTotal        int `json:"classified_claim_total"`
	OpenClaimTotal              int `json:"open_claim_total"`
	DischargedClaimTotal        int `json:"discharged_claim_total"`
	RefutedClaimTotal           int `json:"refuted_claim_total"`
	UnknownClaimTotal           int `json:"unknown_claim_total"`
	DirectUnknownClaimTotal     int `json:"direct_unknown_claim_total"`
	DependencyBlockedClaimTotal int `json:"dependency_blocked_claim_total"`
	DirectRefutedClaimTotal     int `json:"direct_refuted_claim_total"`
	DependencyRefutedClaimTotal int `json:"dependency_refuted_claim_total"`
	DependencyRecoveredTotal    int `json:"dependency_recovered_claim_total"`
	ObservedBlockingEdgeTotal   int `json:"observed_blocking_edge_total"`
	ObservedRefutingEdgeTotal   int `json:"observed_refuting_edge_total"`
	ObservedRecoveryEdgeTotal   int `json:"observed_recovery_edge_total"`
	MaximumCausePathDepth       int `json:"maximum_cause_path_depth"`
	TransitionTotal             int `json:"transition_total"`
	ClassificationBasisPoints   int `json:"classification_basis_points"`
}

type Decision struct {
	Value                       string `json:"value"`
	Resolution                  string `json:"resolution"`
	Reason                      string `json:"reason"`
	SemanticPromotionAuthorized bool   `json:"semantic_promotion_authorized"`
}

type Receipt struct {
	Schema      string       `json:"schema"`
	Scope       string       `json:"scope"`
	Subject     Subject      `json:"subject"`
	Graph       Graph        `json:"graph"`
	Metrics     Metrics      `json:"metrics"`
	Transitions []Transition `json:"transitions"`
	Resolutions []Resolution `json:"resolutions"`
	Decision    Decision     `json:"decision"`
	Digest      string       `json:"digest"`
}
