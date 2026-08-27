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
	ObservationFailure       ObservationPredicate = "FAILURE_ANTECEDENT_OBSERVED"
)

type EvidenceStatus string

const (
	CurrentEvidence   EvidenceStatus = "CURRENT_EVIDENCE"
	HistoricalFixture EvidenceStatus = "HISTORICAL_FIXTURE"
	UnknownEvidence   EvidenceStatus = "UNKNOWN"
)

type AuthorityCase struct {
	CaseID             string         `json:"case_id"`
	NetworkState       string         `json:"network_state"`
	CapabilityStatus   EvidenceStatus `json:"capability_status"`
	ExpectedResolution string         `json:"expected_resolution"`
	ObservedResolution string         `json:"observed_resolution"`
}

type ObservationReceipt struct {
	Schema            string               `json:"schema"`
	Provider          string               `json:"provider"`
	Binding           string               `json:"binding"`
	ClaimID           string               `json:"claim_id,omitempty"`
	PropositionDigest string               `json:"proposition_digest,omitempty"`
	EdgeID            string               `json:"edge_id,omitempty"`
	FromClaimID       string               `json:"from_claim_id,omitempty"`
	ToClaimID         string               `json:"to_claim_id,omitempty"`
	EdgeKind          EdgeKind             `json:"edge_kind,omitempty"`
	Target            TargetAddress        `json:"target"`
	Occurrence        TargetOccurrence     `json:"target_occurrence"`
	TargetPath        string               `json:"target_path"`
	TargetBytesDigest string               `json:"target_bytes_digest"`
	ExpectedPredicate ObservationPredicate `json:"expected_predicate"`
	ExpectedValue     string               `json:"expected_value"`
	ObservedPredicate ObservationPredicate `json:"observed_predicate"`
	ObservedValue     string               `json:"observed_value"`
	ComparisonResult  string               `json:"comparison_result"`
	Procedure         string               `json:"procedure"`
	ProcedureDigest   string               `json:"procedure_digest"`
	Output            string               `json:"output"`
	OutputDigest      string               `json:"output_digest"`
	Coordinate        Coordinate           `json:"coordinate"`
	Digest            string               `json:"digest"`
}

type ObservationBundle struct {
	Schema                   string                    `json:"schema"`
	Provider                 string                    `json:"provider"`
	SourcePath               string                    `json:"source_path"`
	SourceDigest             string                    `json:"source_digest"`
	ArtifactPath             string                    `json:"artifact_path"`
	ArtifactBytesDigest      string                    `json:"artifact_bytes_digest"`
	ContractPath             string                    `json:"contract_path"`
	ContractDigest           string                    `json:"contract_digest"`
	ContractRaw              []byte                    `json:"contract_raw"`
	FailureReceiptPath       string                    `json:"failure_receipt_path,omitempty"`
	FailureReceiptDigest     string                    `json:"failure_receipt_digest,omitempty"`
	FailureReceiptRaw        []byte                    `json:"failure_receipt_raw,omitempty"`
	Profile                  string                    `json:"profile"`
	Observations             []ObservationReceipt      `json:"observations"`
	StructuralContradictions []StructuralContradiction `json:"structural_contradictions,omitempty"`
	Digest                   string                    `json:"digest"`
}

// StructuralContradiction records a source/contract disagreement without
// pretending that the target process was observed to fail.
type StructuralContradiction struct {
	ClaimID           string           `json:"claim_id"`
	PropositionDigest string           `json:"proposition_digest"`
	ExpectedValue     string           `json:"expected_value"`
	DeclaredValue     string           `json:"declared_value"`
	ProcedureID       string           `json:"procedure_id"`
	Occurrence        TargetOccurrence `json:"target_occurrence"`
	Digest            string           `json:"digest"`
}

// TargetOccurrence is the canonical parse/lower identity of one activity in
// the independently observed artifact.  RowDigest is the raw declaration
// bytes; SemanticDigest is the lowered target IR digest.  Neither is inferred
// from a profile label or a line-prefix scan.
type TargetOccurrence struct {
	Address           string        `json:"address"`
	ActivityName      string        `json:"activity_name"`
	ClaimID           string        `json:"claim_id"`
	PropositionDigest string        `json:"proposition_digest"`
	Target            TargetAddress `json:"target"`
	ValueProgram      string        `json:"value_program"`
	RowDigest         string        `json:"row_digest"`
	SemanticDigest    string        `json:"semantic_digest"`
}

// ValidatorContract is external expected material. It contains no outcome,
// status, or edge activation decision; those are calculated from raw target
// bytes and, for failure entailment, an independently captured process exit.
type ValidatorContract struct {
	Schema                 string           `json:"schema"`
	ContractID             string           `json:"contract_id"`
	ExpectedArtifactPath   string           `json:"expected_artifact_path"`
	ExpectedArtifactDigest string           `json:"expected_artifact_digest"`
	Claims                 []ValidatorClaim `json:"claims"`
}

type ValidatorClaim struct {
	ClaimID                string        `json:"claim_id"`
	PropositionDigest      string        `json:"proposition_digest"`
	ProcedureID            string        `json:"procedure_id"`
	TargetRowDigest        string        `json:"target_row_digest"`
	AlternateRowDigest     string        `json:"alternate_row_digest,omitempty"`
	ExpectedMaterialDigest string        `json:"expected_material_digest"`
	ActivityName           string        `json:"activity_name"`
	ExpectedTarget         TargetAddress `json:"expected_target"`
	ExpectedValueProgram   string        `json:"expected_value_program"`
	AlternateValueProgram  string        `json:"alternate_value_program,omitempty"`
}

// FailureReceipt is made only after a CI process has actually returned a
// non-zero exit. A caller-supplied label is never sufficient to construct it.
type FailureReceipt struct {
	Schema              string         `json:"schema"`
	Provider            string         `json:"provider"`
	SourcePath          string         `json:"source_path"`
	SourceDigest        string         `json:"source_digest"`
	ArtifactPath        string         `json:"artifact_path"`
	ArtifactBytesDigest string         `json:"artifact_bytes_digest"`
	EdgeID              string         `json:"edge_id"`
	FromClaimID         string         `json:"from_claim_id"`
	ToClaimID           string         `json:"to_claim_id"`
	EdgeKind            EdgeKind       `json:"edge_kind"`
	Target              TargetAddress  `json:"target"`
	Procedure           string         `json:"procedure"`
	ProcedureDigest     string         `json:"procedure_digest"`
	Executable          string         `json:"executable"`
	ExecutableDigest    string         `json:"executable_digest"`
	ExecutableRaw       []byte         `json:"executable_raw"`
	Argv                []string       `json:"argv"`
	InputTargets        []FailureInput `json:"input_targets"`
	Stdout              []byte         `json:"stdout"`
	StdoutDigest        string         `json:"stdout_digest"`
	Stderr              []byte         `json:"stderr"`
	StderrDigest        string         `json:"stderr_digest"`
	ObservedExitCode    int            `json:"observed_exit_code"`
	Result              string         `json:"result"`
	Coordinate          Coordinate     `json:"coordinate"`
	Digest              string         `json:"digest"`
}

type FailureInput struct {
	ClaimID            string           `json:"claim_id"`
	PropositionDigest  string           `json:"proposition_digest"`
	Target             TargetAddress    `json:"target"`
	Occurrence         TargetOccurrence `json:"target_occurrence"`
	TargetOutputDigest string           `json:"target_output_digest"`
	ValueProgram       string           `json:"value_program"`
	ArtifactPath       string           `json:"artifact_path"`
	ArtifactDigest     string           `json:"artifact_digest"`
}

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
	EdgeID              string   `json:"edge_id"`
	FromClaimID         string   `json:"from_claim_id"`
	ToClaimID           string   `json:"to_claim_id"`
	Kind                EdgeKind `json:"kind"`
	ActivationPredicate string   `json:"activation_predicate"`
	SemanticBasis       string   `json:"semantic_basis"`
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
	Provider   string            `json:"provider"`
	Permission string            `json:"permission"`
	Status     EvidenceStatus    `json:"status"`
	Toolchain  ToolchainEvidence `json:"toolchain"`
	Coordinate Coordinate        `json:"coordinate"`
	Digest     string            `json:"digest"`
}

type ToolchainEvidence struct {
	Name    string `json:"name"`
	Version string `json:"version"`
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
	Schema                     string                    `json:"schema"`
	Provider                   string                    `json:"provider"`
	SourcePath                 string                    `json:"source_path"`
	SourceBytesDigest          string                    `json:"source_bytes_digest"`
	SourceGraphDigest          string                    `json:"source_graph_digest"`
	ArtifactPath               string                    `json:"artifact_path"`
	ArtifactBytesDigest        string                    `json:"artifact_bytes_digest"`
	Operation                  string                    `json:"operation"`
	RequestStatus              string                    `json:"request_status"`
	Procedure                  string                    `json:"procedure"`
	ObservationPath            string                    `json:"observation_path,omitempty"`
	ObservationBundleDigest    string                    `json:"observation_bundle_digest,omitempty"`
	ObservationBundleRawDigest string                    `json:"observation_bundle_raw_digest,omitempty"`
	ObservationBundleRaw       []byte                    `json:"observation_bundle_raw,omitempty"`
	Observations               []ObservationReceipt      `json:"observations"`
	StructuralContradictions   []StructuralContradiction `json:"structural_contradictions,omitempty"`
	ObservedPredicate          ObservationPredicate      `json:"observed_predicate"`
	ObservedValue              string                    `json:"observed_value"`
	Status                     EvidenceStatus            `json:"status"`
	Coordinate                 Coordinate                `json:"coordinate"`
	Claims                     []EvidenceClaim           `json:"claims"`
	Capability                 CapabilityEvidence        `json:"capability"`
	Snapshot                   RepositorySnapshot        `json:"snapshot"`
	Digest                     string                    `json:"digest"`
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
	Schema                    string   `json:"schema"`
	CaseID                    string   `json:"case_id"`
	Kind                      EdgeKind `json:"kind"`
	Direction                 string   `json:"direction"`
	UpstreamState             string   `json:"upstream_state"`
	LocalPredicate            string   `json:"local_predicate"`
	ExpectedState             string   `json:"expected_state"`
	Positive                  bool     `json:"positive"`
	ContradictionObserved     bool     `json:"contradiction_observed"`
	FailureAntecedentObserved bool     `json:"failure_antecedent_observed"`
	SemanticBasis             string   `json:"semantic_basis"`
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
	FixedClaimTotal                    int          `json:"fixed_claim_total"`
	DistinctPropositionTotal           int          `json:"distinct_proposition_total"`
	StructuralContradictionNumerator   int          `json:"structural_contradiction_numerator"`
	StructuralContradictionDenominator int          `json:"structural_contradiction_denominator"`
	FixedEdgeTotal                     int          `json:"fixed_edge_total"`
	EligibleEdgeTotal                  int          `json:"eligible_edge_total"`
	ObservedCausalEdgeTotal            int          `json:"observed_causal_edge_total"`
	ShortestPathEdgeUnionTotal         int          `json:"shortest_path_edge_union_total"`
	ClassifiedClaimTotal               int          `json:"classified_claim_total"`
	OpenClaimTotal                     int          `json:"open_claim_total"`
	DischargedClaimTotal               int          `json:"discharged_claim_total"`
	RefutedClaimTotal                  int          `json:"refuted_claim_total"`
	CurrentEvidenceTotal               int          `json:"current_evidence_total"`
	HistoricalEvidenceTotal            int          `json:"historical_evidence_total"`
	UnknownEvidenceTotal               int          `json:"unknown_evidence_total"`
	DirectUnknownClaimTotal            int          `json:"direct_unknown_claim_total"`
	DependencyBlockedClaimTotal        int          `json:"dependency_blocked_claim_total"`
	DirectRefutedClaimTotal            int          `json:"direct_refuted_claim_total"`
	DependencyRefutedClaimTotal        int          `json:"dependency_refuted_claim_total"`
	DirectDischargedClaimTotal         int          `json:"direct_discharged_claim_total"`
	DependencyDischargedTotal          int          `json:"dependency_discharged_claim_total"`
	ObservedBlockingEdgeTotal          int          `json:"observed_blocking_edge_total"`
	ObservedRefutingEdgeTotal          int          `json:"observed_refuting_edge_total"`
	ObservedRecoveryEdgeTotal          int          `json:"observed_recovery_edge_total"`
	MaximumCausePathDepth              int          `json:"maximum_cause_path_depth"`
	TransitionTotal                    int          `json:"transition_total"`
	AppendOnlyTransitionTotal          int          `json:"append_only_transition_total"`
	ClassificationBasisPoints          int          `json:"classification_basis_points"`
	TruthTableCaseTotal                int          `json:"truth_table_case_total"`
	AuthorityCaseTotal                 int          `json:"authority_case_total"`
	EdgeMetrics                        []EdgeMetric `json:"edge_metrics"`
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
	AuthorityCases           []AuthorityCase  `json:"authority_cases"`
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
