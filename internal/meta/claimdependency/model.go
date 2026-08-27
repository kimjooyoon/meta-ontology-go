package claimdependency

// The model is deliberately about executable propositions and evidence, not
// declaration existence.  JSON is the interchange format used by the CI
// producer and the raw-input consumer.
const (
	ReceiptSchema          = "gooo.meta.claim-dependency-receipt/v3"
	GraphSchema            = "gooo.meta.claim-dependency-graph/v3"
	EvidenceSchema         = "gooo.meta.claim-dependency-evidence/v2"
	TruthTableSchema       = "gooo.meta.claim-dependency-truth-table/v1"
	Scope                  = "CLAIM_STATE_PROPAGATION_ONLY"
	ClaimTotal             = 6
	EdgeTotal              = 8
	InitialTransitionTotal = 12
	ProducerID             = "gooo://meta/claim-dependency/producer/v3"
	ConsumerID             = "gooo://meta/claim-dependency/independent-judge/v3"
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

type EvidenceStatus string

const (
	CurrentEvidence   EvidenceStatus = "CURRENT_EVIDENCE"
	HistoricalFixture EvidenceStatus = "HISTORICAL_FIXTURE"
	UnknownEvidence   EvidenceStatus = "UNKNOWN"
)

type Coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}

type TargetAddress struct {
	Inputs   []string `json:"inputs,omitempty"`
	Output   string   `json:"output,omitempty"`
	Artifact string   `json:"artifact"`
}

type Claim struct {
	Ordinal           int           `json:"ordinal"`
	Axis              string        `json:"axis"`
	ClaimID           string        `json:"claim_id"`
	ActivityID        string        `json:"activity_id"`
	ActivityName      string        `json:"activity_name"`
	Proposition       string        `json:"proposition"`
	PropositionDigest string        `json:"proposition_digest"`
	Target            TargetAddress `json:"target"`
	ValueProgram      string        `json:"value_program"`
	Producer          string        `json:"producer"`
	Consumer          string        `json:"consumer"`
	MetaOperation     string        `json:"meta_operation"`
	ProofChoice       string        `json:"proof_choice"`
	Coordinate        Coordinate    `json:"coordinate"`
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

type EvidenceClaim struct {
	ClaimID           string               `json:"claim_id"`
	PropositionDigest string               `json:"proposition_digest"`
	ObservedPredicate ObservationPredicate `json:"observed_predicate"`
	ObservedValue     string               `json:"observed_value"`
	Status            EvidenceStatus       `json:"status"`
	Coordinate        Coordinate           `json:"coordinate"`
	Digest            string               `json:"digest"`
}

type CapabilityEvidence struct {
	Provider   string         `json:"provider"`
	Permission string         `json:"permission"`
	Status     EvidenceStatus `json:"status"`
	Coordinate Coordinate     `json:"coordinate"`
	Digest     string         `json:"digest"`
}

type RepositorySnapshot struct {
	RepositoryRoot   string     `json:"repository_root"`
	TrackedDigest    string     `json:"tracked_digest"`
	UntrackedDigest  string     `json:"untracked_digest"`
	BeforeDigest     string     `json:"before_digest"`
	AfterDigest      string     `json:"after_digest"`
	OutputPath       string     `json:"output_path"`
	OutputDigest     string     `json:"output_digest"`
	RepositoryWrites int        `json:"repository_writes"`
	Coordinate       Coordinate `json:"coordinate"`
}

type EvidenceReceipt struct {
	Schema              string               `json:"schema"`
	Provider            string               `json:"provider"`
	ArtifactPath        string               `json:"artifact_path"`
	ArtifactBytesDigest string               `json:"artifact_bytes_digest"`
	Operation           string               `json:"operation"`
	ObservedPredicate   ObservationPredicate `json:"observed_predicate"`
	ObservedValue       string               `json:"observed_value"`
	Status              EvidenceStatus       `json:"status"`
	Coordinate          Coordinate           `json:"coordinate"`
	Claims              []EvidenceClaim      `json:"claims"`
	Capability          CapabilityEvidence   `json:"capability"`
	Snapshot            RepositorySnapshot   `json:"snapshot"`
	Digest              string               `json:"digest"`
}

type Subject struct {
	SourcePath          string     `json:"source_path"`
	SourceDigest        string     `json:"source_digest"`
	SemanticDigest      string     `json:"semantic_digest"`
	Producer            string     `json:"producer"`
	Consumer            string     `json:"consumer"`
	MetaOperation       string     `json:"meta_operation"`
	ProofChoice         string     `json:"proof_choice"`
	ReadOnly            bool       `json:"read_only"`
	RepositoryWrites    int        `json:"repository_writes"`
	AuthorityResolution string     `json:"authority_resolution"`
	AuthorityCoordinate Coordinate `json:"authority_coordinate"`
}

type Transition struct {
	Sequence                  int        `json:"sequence"`
	ClaimID                   string     `json:"claim_id"`
	Event                     string     `json:"event"`
	Before                    string     `json:"before"`
	After                     string     `json:"after"`
	Coordinate                Coordinate `json:"coordinate"`
	EvidenceDigest            string     `json:"evidence_digest,omitempty"`
	UpstreamEdgeIDs           []string   `json:"upstream_edge_ids,omitempty"`
	UpstreamTransitionDigests []string   `json:"upstream_transition_digests,omitempty"`
	Provenance                string     `json:"provenance"`
	PreviousTransitionDigest  string     `json:"previous_transition_digest,omitempty"`
	TransitionDigest          string     `json:"transition_digest"`
}

type Resolution struct {
	ClaimID                string      `json:"claim_id"`
	Axis                   string      `json:"axis"`
	PropositionDigest      string      `json:"proposition_digest"`
	State                  string      `json:"state"`
	Kind                   string      `json:"kind"`
	ObservedEvent          string      `json:"observed_event"`
	Coordinate             Coordinate  `json:"coordinate"`
	EvidenceDigest         string      `json:"evidence_digest,omitempty"`
	Provenance             string      `json:"provenance"`
	FailureResponsibility  string      `json:"failure_responsibility"`
	FailureOwnerClaimID    string      `json:"failure_owner_claim_id"`
	MissingEvidenceIDs     []string    `json:"missing_evidence_ids,omitempty"`
	BlockedByClaimIDs      []string    `json:"blocked_by_claim_ids,omitempty"`
	BlockedByEdgeIDs       []string    `json:"blocked_by_edge_ids,omitempty"`
	CausePath              []string    `json:"cause_path"`
	CauseEdgeIDs           []string    `json:"cause_edge_ids"`
	CauseEdgeKinds         []EdgeKind  `json:"cause_edge_kinds"`
	CauseTransitionDigests []string    `json:"cause_transition_digests"`
	CauseCoordinate        *Coordinate `json:"cause_coordinate"`
}

type TruthTableCase struct {
	Schema         string   `json:"schema"`
	CaseID         string   `json:"case_id"`
	Kind           EdgeKind `json:"kind"`
	Direction      string   `json:"direction"`
	UpstreamState  string   `json:"upstream_state"`
	LocalPredicate string   `json:"local_predicate"`
	ExpectedState  string   `json:"expected_state"`
	Positive       bool     `json:"positive"`
	SemanticBasis  string   `json:"semantic_basis"`
}

type EdgeMetric struct {
	Kind           EdgeKind `json:"kind"`
	Eligible       int      `json:"eligible"`
	ObservedCausal int      `json:"observed_causal"`
	Blocking       int      `json:"blocking"`
	Refuting       int      `json:"refuting"`
	Discharge      int      `json:"discharge"`
}

type Metrics struct {
	FixedClaimTotal             int          `json:"fixed_claim_total"`
	DistinctPropositionTotal    int          `json:"distinct_proposition_total"`
	FixedEdgeTotal              int          `json:"fixed_edge_total"`
	EligibleEdgeTotal           int          `json:"eligible_edge_total"`
	ObservedCausalEdgeTotal     int          `json:"observed_causal_edge_total"`
	MinimumCausalEdgeTotal      int          `json:"minimum_causal_edge_total"`
	ClassifiedClaimTotal        int          `json:"classified_claim_total"`
	OpenClaimTotal              int          `json:"open_claim_total"`
	DischargedClaimTotal        int          `json:"discharged_claim_total"`
	RefutedClaimTotal           int          `json:"refuted_claim_total"`
	CurrentEvidenceTotal        int          `json:"current_evidence_total"`
	HistoricalEvidenceTotal     int          `json:"historical_evidence_total"`
	UnknownEvidenceTotal        int          `json:"unknown_evidence_total"`
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
	TruthTableCaseTotal         int          `json:"truth_table_case_total"`
	EdgeMetrics                 []EdgeMetric `json:"edge_metrics"`
}

type Decision struct {
	Value                       string `json:"value"`
	Resolution                  string `json:"resolution"`
	Reason                      string `json:"reason"`
	SemanticPromotionAuthorized bool   `json:"semantic_promotion_authorized"`
}

type Receipt struct {
	Schema                   string           `json:"schema"`
	Scope                    string           `json:"scope"`
	Subject                  Subject          `json:"subject"`
	Evidence                 EvidenceReceipt  `json:"evidence"`
	Graph                    Graph            `json:"graph"`
	TruthTable               []TruthTableCase `json:"truth_table"`
	PriorReceiptDigest       string           `json:"prior_receipt_digest,omitempty"`
	PreviousTransitionDigest string           `json:"previous_transition_digest,omitempty"`
	PriorClaimStates         []string         `json:"prior_claim_states,omitempty"`
	EvidenceDigest           string           `json:"evidence_digest"`
	Transitions              []Transition     `json:"transitions"`
	TransitionHeadDigest     string           `json:"transition_head_digest"`
	Resolutions              []Resolution     `json:"resolutions"`
	Metrics                  Metrics          `json:"metrics"`
	Decision                 Decision         `json:"decision"`
	Digest                   string           `json:"digest"`
}
