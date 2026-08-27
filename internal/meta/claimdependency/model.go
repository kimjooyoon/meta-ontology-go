package claimdependency

const (
	ReceiptSchema          = "gooo.meta.claim-dependency-receipt/v2"
	GraphSchema            = "gooo.meta.claim-dependency-graph/v2"
	ObservationSchema      = "gooo.meta.claim-dependency-observation/v1"
	Scope                  = "CLAIM_STATE_PROPAGATION_ONLY"
	ClaimTotal             = 6
	EdgeTotal              = 8
	InitialTransitionTotal = ClaimTotal * 2
	ProducerID             = "gooo://meta/claim-dependency/producer/v2"
	ConsumerID             = "gooo://meta/claim-dependency/independent-judge/v2"
	MetaOperationID        = "classify-claim-state-causality"
	ProofChoice            = "COHERENCE"
)

type EdgeKind string

const (
	Supports          EdgeKind = "SUPPORTS"
	Requires          EdgeKind = "REQUIRES"
	Contradicts       EdgeKind = "CONTRADICTS"
	FailureEntailment EdgeKind = "FAILURE_ENTAILMENT"
)

type ObservationPredicate string

const (
	ObservationUnknown       ObservationPredicate = "UNKNOWN"
	ObservationEvidence      ObservationPredicate = "EVIDENCE_ACCEPTED"
	ObservationContradiction ObservationPredicate = "EXPLICIT_CONTRADICTION"
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
	ActivityID    string     `json:"activity_id"`
	ActivityName  string     `json:"activity_name"`
	Statement     string     `json:"statement"`
	ValueProgram  string     `json:"value_program"`
	Producer      string     `json:"producer"`
	Consumer      string     `json:"consumer"`
	MetaOperation string     `json:"meta_operation"`
	ProofChoice   string     `json:"proof_choice"`
	Coordinate    Coordinate `json:"coordinate"`
}

type Edge struct {
	EdgeID        string   `json:"edge_id"`
	FromClaimID   string   `json:"from_claim_id"`
	ToClaimID     string   `json:"to_claim_id"`
	Kind          EdgeKind `json:"kind"`
	SemanticBasis string   `json:"semantic_basis"`
}

type Graph struct {
	Schema            string  `json:"schema"`
	Authority         string  `json:"authority"`
	Completeness      string  `json:"completeness"`
	CanonicalIRDigest string  `json:"canonical_ir_digest"`
	NodeTotal         int     `json:"node_total"`
	EdgeTotal         int     `json:"edge_total"`
	Nodes             []Claim `json:"nodes"`
	Edges             []Edge  `json:"edges"`
	Digest            string  `json:"digest"`
}

type Observation struct {
	Schema            string               `json:"schema"`
	Predicate         ObservationPredicate `json:"predicate"`
	SubjectClaimID    string               `json:"subject_claim_id"`
	EvidenceDigest    string               `json:"evidence_digest,omitempty"`
	ReadOnly          bool                 `json:"read_only"`
	RepositoryWrites  int                  `json:"repository_writes"`
	MutationAuthority bool                 `json:"mutation_authority"`
	Digest            string               `json:"digest"`
}

type Subject struct {
	SourcePath       string `json:"source_path"`
	SourceDigest     string `json:"source_digest"`
	SemanticDigest   string `json:"semantic_digest"`
	Producer         string `json:"producer"`
	Consumer         string `json:"consumer"`
	MetaOperation    string `json:"meta_operation"`
	ProofChoice      string `json:"proof_choice"`
	ReadOnly         bool   `json:"read_only"`
	RepositoryWrites int    `json:"repository_writes"`
}

type Transition struct {
	Sequence                 int        `json:"sequence"`
	ClaimID                  string     `json:"claim_id"`
	Event                    string     `json:"event"`
	Before                   string     `json:"before"`
	After                    string     `json:"after"`
	Coordinate               Coordinate `json:"coordinate"`
	EvidenceDigest           string     `json:"evidence_digest,omitempty"`
	Provenance               string     `json:"provenance"`
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
	EvidenceDigest        string      `json:"evidence_digest,omitempty"`
	Provenance            string      `json:"provenance"`
	FailureResponsibility string      `json:"failure_responsibility"`
	FailureOwnerClaimID   string      `json:"failure_owner_claim_id"`
	MissingEvidenceIDs    []string    `json:"missing_evidence_ids,omitempty"`
	BlockedByClaimIDs     []string    `json:"blocked_by_claim_ids,omitempty"`
	BlockedByEdgeIDs      []string    `json:"blocked_by_edge_ids,omitempty"`
	CausePath             []string    `json:"cause_path"`
	CauseEdgeIDs          []string    `json:"cause_edge_ids"`
	CauseEdgeKinds        []EdgeKind  `json:"cause_edge_kinds"`
	CauseTransitionDigest string      `json:"cause_transition_digest"`
	CauseCoordinate       *Coordinate `json:"cause_coordinate"`
}

type EdgeMetric struct {
	Kind     EdgeKind `json:"kind"`
	Total    int      `json:"total"`
	Blocking int      `json:"blocking"`
	Refuting int      `json:"refuting"`
	Recovery int      `json:"recovery"`
}

type Metrics struct {
	FixedClaimTotal             int          `json:"fixed_claim_total"`
	FixedEdgeTotal              int          `json:"fixed_edge_total"`
	ClassifiedClaimTotal        int          `json:"classified_claim_total"`
	OpenClaimTotal              int          `json:"open_claim_total"`
	DischargedClaimTotal        int          `json:"discharged_claim_total"`
	RefutedClaimTotal           int          `json:"refuted_claim_total"`
	UnknownClaimTotal           int          `json:"unknown_claim_total"`
	DirectUnknownClaimTotal     int          `json:"direct_unknown_claim_total"`
	DependencyBlockedClaimTotal int          `json:"dependency_blocked_claim_total"`
	DirectRefutedClaimTotal     int          `json:"direct_refuted_claim_total"`
	DependencyRefutedClaimTotal int          `json:"dependency_refuted_claim_total"`
	DirectDischargedClaimTotal  int          `json:"direct_discharged_claim_total"`
	DependencyDischargedTotal   int          `json:"dependency_discharged_claim_total"`
	ObservedBlockingEdgeTotal   int          `json:"observed_blocking_edge_total"`
	ObservedRefutingEdgeTotal   int          `json:"observed_refuting_edge_total"`
	ObservedRecoveryEdgeTotal   int          `json:"observed_recovery_edge_total"`
	MaximumCausePathDepth       int          `json:"maximum_cause_path_depth"`
	TransitionTotal             int          `json:"transition_total"`
	AppendOnlyTransitionTotal   int          `json:"append_only_transition_total"`
	ClassificationBasisPoints   int          `json:"classification_basis_points"`
	EdgeMetrics                 []EdgeMetric `json:"edge_metrics"`
}

type Decision struct {
	Value                       string `json:"value"`
	Resolution                  string `json:"resolution"`
	Reason                      string `json:"reason"`
	SemanticPromotionAuthorized bool   `json:"semantic_promotion_authorized"`
}

type Receipt struct {
	Schema                   string       `json:"schema"`
	Scope                    string       `json:"scope"`
	Subject                  Subject      `json:"subject"`
	Observation              Observation  `json:"observation"`
	Graph                    Graph        `json:"graph"`
	PriorReceiptDigest       string       `json:"prior_receipt_digest,omitempty"`
	PreviousTransitionDigest string       `json:"previous_transition_digest,omitempty"`
	PriorClaimStates         []string     `json:"prior_claim_states,omitempty"`
	ObservationDigest        string       `json:"observation_digest"`
	Transitions              []Transition `json:"transitions"`
	TransitionHeadDigest     string       `json:"transition_head_digest"`
	Resolutions              []Resolution `json:"resolutions"`
	Metrics                  Metrics      `json:"metrics"`
	Decision                 Decision     `json:"decision"`
	Digest                   string       `json:"digest"`
}
