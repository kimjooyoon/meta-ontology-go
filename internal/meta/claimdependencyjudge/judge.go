package claimdependencyjudge

// This package intentionally repeats the small raw-input reconstruction and
// state algebra. It must remain import-independent from the producer so a
// producer receipt is comparison evidence, not the judge's source of truth.
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const judgmentSchema = "gooo.meta.claim-dependency-judgment/v3"
const claimTotal, edgeTotal, initialTransitions = 6, 8, 12
const producerID = "gooo://meta/claim-dependency/producer/v3"
const consumerID = "gooo://meta/claim-dependency/independent-judge/v3"
const operation, proof = "classify-claim-state-causality", "COHERENCE"
const evidenceProcedure = "RAW_ARTIFACT_OBSERVATION_BINDING_V3"
const observationSchema = "gooo.meta.claim-dependency-observation/v3"
const observationBundleSchema = "gooo.meta.claim-dependency-observation-bundle/v2"
const validatorContractSchema = "gooo.meta.claim-dependency-validator-contract/v2"
const structuralManifestSchema = "gooo.meta.claim-dependency-structural-manifest/v1"
const failureReceiptSchema = "gooo.meta.claim-dependency-failure-receipt/v2"
const failureProcedure = "CI_EDGE_SPECIFIC_FAILURE_COMPARATOR_V2"

type edgeKind string

const (
	supports          edgeKind = "SUPPORTS"
	requires          edgeKind = "REQUIRES"
	contradicts       edgeKind = "CONTRADICTS"
	failureEntailment edgeKind = "FAILURE_ENTAILMENT"
)

type predicate string

const (
	unknown               predicate = "UNKNOWN"
	accepted              predicate = "EVIDENCE_ACCEPTED"
	explicitContradiction predicate = "EXPLICIT_CONTRADICTION"
	failureAntecedent     predicate = "FAILURE_ANTECEDENT_OBSERVED"
)

type coordinate struct {
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Reason string `json:"reason"`
}
type target struct {
	Inputs   []string `json:"inputs,omitempty"`
	Output   string   `json:"output,omitempty"`
	Artifact string   `json:"artifact"`
}
type observationReceipt struct {
	Schema              string           `json:"schema"`
	Provider            string           `json:"provider"`
	Binding             string           `json:"binding"`
	ClaimID             string           `json:"claim_id,omitempty"`
	PropositionDigest   string           `json:"proposition_digest,omitempty"`
	EdgeID              string           `json:"edge_id,omitempty"`
	FromClaimID         string           `json:"from_claim_id,omitempty"`
	ToClaimID           string           `json:"to_claim_id,omitempty"`
	EdgeKind            edgeKind         `json:"edge_kind,omitempty"`
	Target              target           `json:"target"`
	Occurrence          targetOccurrence `json:"target_occurrence"`
	TargetPath          string           `json:"target_path"`
	TargetBytesDigest   string           `json:"target_bytes_digest"`
	ExpectedPredicate   predicate        `json:"expected_predicate"`
	ExpectedValue       string           `json:"expected_value"`
	ObservedPredicate   predicate        `json:"observed_predicate"`
	ObservedValue       string           `json:"observed_value"`
	ComparisonResult    string           `json:"comparison_result"`
	Procedure           string           `json:"procedure"`
	ProcedureDigest     string           `json:"procedure_digest"`
	RawProvenanceDigest string           `json:"raw_provenance_digest"`
	Output              string           `json:"output"`
	OutputDigest        string           `json:"output_digest"`
	Coordinate          coordinate       `json:"coordinate"`
	Digest              string           `json:"digest"`
}
type observationBundle struct {
	Schema                          string                    `json:"schema"`
	Provider                        string                    `json:"provider"`
	SourcePath                      string                    `json:"source_path"`
	SourceDigest                    string                    `json:"source_digest"`
	ArtifactPath                    string                    `json:"artifact_path"`
	ArtifactBytesDigest             string                    `json:"artifact_bytes_digest"`
	ContractPath                    string                    `json:"contract_path"`
	ContractDigest                  string                    `json:"contract_digest"`
	ContractRaw                     []byte                    `json:"contract_raw"`
	StructuralManifestPath          string                    `json:"structural_manifest_path"`
	StructuralManifestDigest        string                    `json:"structural_manifest_digest"`
	StructuralManifestRaw           []byte                    `json:"structural_manifest_raw"`
	FailureReceiptPath              string                    `json:"failure_receipt_path,omitempty"`
	FailureReceiptDigest            string                    `json:"failure_receipt_digest,omitempty"`
	FailureReceiptRaw               []byte                    `json:"failure_receipt_raw,omitempty"`
	Profile                         string                    `json:"profile"`
	Observations                    []observationReceipt      `json:"observations"`
	StructuralContradictions        []structuralContradiction `json:"structural_contradictions,omitempty"`
	StructuralInventoryTotal        int                       `json:"structural_inventory_total"`
	SemanticOccurrenceNumerator     int                       `json:"semantic_occurrence_numerator"`
	SemanticOccurrenceDenominator   int                       `json:"semantic_occurrence_denominator"`
	RawProvenanceBindingNumerator   int                       `json:"raw_provenance_binding_numerator"`
	RawProvenanceBindingDenominator int                       `json:"raw_provenance_binding_denominator"`
	Digest                          string                    `json:"digest"`
}
type structuralContradiction struct {
	ClaimID           string           `json:"claim_id"`
	PropositionDigest string           `json:"proposition_digest"`
	ExpectedValue     string           `json:"expected_value"`
	DeclaredValue     string           `json:"declared_value"`
	ProcedureID       string           `json:"procedure_id"`
	Occurrence        targetOccurrence `json:"target_occurrence"`
	Digest            string           `json:"digest"`
}
type targetOccurrence struct {
	Address           string `json:"address"`
	ActivityName      string `json:"activity_name"`
	ClaimID           string `json:"claim_id"`
	PropositionDigest string `json:"proposition_digest"`
	Target            target `json:"target"`
	ValueProgram      string `json:"value_program"`
	RawSpanStart      int    `json:"raw_span_start"`
	RawSpanEnd        int    `json:"raw_span_end"`
	RawRowDigest      string `json:"raw_row_digest"`
	SemanticDigest    string `json:"semantic_digest"`
	ContextDigest     string `json:"context_digest"`
}
type validatorContract struct {
	Schema                 string           `json:"schema"`
	ContractID             string           `json:"contract_id"`
	ExpectedArtifactPath   string           `json:"expected_artifact_path"`
	ExpectedArtifactDigest string           `json:"expected_artifact_digest"`
	Claims                 []validatorClaim `json:"claims"`
}
type structuralInventoryManifest struct {
	Schema                        string   `json:"schema"`
	ManifestID                    string   `json:"manifest_id"`
	ContractID                    string   `json:"contract_id"`
	EligibleClaimIDs              []string `json:"eligible_claim_ids"`
	ExpectedContradictionClaimIDs []string `json:"expected_contradiction_claim_ids"`
}
type validatorClaim struct {
	ClaimID                string `json:"claim_id"`
	PropositionDigest      string `json:"proposition_digest"`
	ProcedureID            string `json:"procedure_id"`
	TargetRowDigest        string `json:"target_row_digest"`
	AlternateRowDigest     string `json:"alternate_row_digest,omitempty"`
	ExpectedMaterialDigest string `json:"expected_material_digest"`
	ActivityName           string `json:"activity_name"`
	ExpectedTarget         target `json:"expected_target"`
	ExpectedValueProgram   string `json:"expected_value_program"`
	AlternateValueProgram  string `json:"alternate_value_program,omitempty"`
}
type failureReceipt struct {
	Schema              string         `json:"schema"`
	Provider            string         `json:"provider"`
	SourcePath          string         `json:"source_path"`
	SourceDigest        string         `json:"source_digest"`
	ArtifactPath        string         `json:"artifact_path"`
	ArtifactBytesDigest string         `json:"artifact_bytes_digest"`
	EdgeID              string         `json:"edge_id"`
	FromClaimID         string         `json:"from_claim_id"`
	ToClaimID           string         `json:"to_claim_id"`
	EdgeKind            edgeKind       `json:"edge_kind"`
	Target              target         `json:"target"`
	Procedure           string         `json:"procedure"`
	ProcedureDigest     string         `json:"procedure_digest"`
	Executable          string         `json:"executable"`
	ExecutableDigest    string         `json:"executable_digest"`
	ExecutableRaw       []byte         `json:"executable_raw"`
	Argv                []string       `json:"argv"`
	InputTargets        []failureInput `json:"input_targets"`
	Stdout              []byte         `json:"stdout"`
	StdoutDigest        string         `json:"stdout_digest"`
	Stderr              []byte         `json:"stderr"`
	StderrDigest        string         `json:"stderr_digest"`
	ObservedExitCode    int            `json:"observed_exit_code"`
	Result              string         `json:"result"`
	Coordinate          coordinate     `json:"coordinate"`
	Digest              string         `json:"digest"`
}
type failureInput struct {
	ClaimID            string           `json:"claim_id"`
	PropositionDigest  string           `json:"proposition_digest"`
	Target             target           `json:"target"`
	Occurrence         targetOccurrence `json:"target_occurrence"`
	TargetOutputDigest string           `json:"target_output_digest"`
	ValueProgram       string           `json:"value_program"`
	ArtifactPath       string           `json:"artifact_path"`
	ArtifactDigest     string           `json:"artifact_digest"`
}
type claim struct {
	Ordinal           int        `json:"ordinal"`
	Axis              string     `json:"axis"`
	ClaimID           string     `json:"claim_id"`
	ActivityID        string     `json:"activity_id"`
	ActivityName      string     `json:"activity_name"`
	Proposition       string     `json:"proposition"`
	PropositionDigest string     `json:"proposition_digest"`
	Target            target     `json:"target"`
	ValueProgram      string     `json:"value_program"`
	Producer          string     `json:"producer"`
	Consumer          string     `json:"consumer"`
	MetaOperation     string     `json:"meta_operation"`
	ProofChoice       string     `json:"proof_choice"`
	Coordinate        coordinate `json:"coordinate"`
}
type edge struct {
	EdgeID              string   `json:"edge_id"`
	FromClaimID         string   `json:"from_claim_id"`
	ToClaimID           string   `json:"to_claim_id"`
	Kind                edgeKind `json:"kind"`
	ActivationPredicate string   `json:"activation_predicate"`
	SemanticBasis       string   `json:"semantic_basis"`
}
type graph struct {
	Schema            string  `json:"schema"`
	Authority         string  `json:"authority"`
	Completeness      string  `json:"completeness"`
	CanonicalIRDigest string  `json:"canonical_ir_digest"`
	NodeTotal         int     `json:"node_total"`
	EdgeTotal         int     `json:"edge_total"`
	Nodes             []claim `json:"nodes"`
	Edges             []edge  `json:"edges"`
	Digest            string  `json:"digest"`
}
type evidenceClaim struct {
	ClaimID           string     `json:"claim_id"`
	PropositionDigest string     `json:"proposition_digest"`
	ObservedPredicate predicate  `json:"observed_predicate"`
	ObservedValue     string     `json:"observed_value"`
	Status            string     `json:"status"`
	Coordinate        coordinate `json:"coordinate"`
	Digest            string     `json:"digest"`
}
type capability struct {
	Provider   string     `json:"provider"`
	Permission string     `json:"permission"`
	Status     string     `json:"status"`
	Toolchain  toolchain  `json:"toolchain"`
	Coordinate coordinate `json:"coordinate"`
	Digest     string     `json:"digest"`
}
type toolchain struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}
type authorityCase struct {
	CaseID             string `json:"case_id"`
	NetworkState       string `json:"network_state"`
	CapabilityStatus   string `json:"capability_status"`
	ExpectedResolution string `json:"expected_resolution"`
	ObservedResolution string `json:"observed_resolution"`
}
type snapshot struct {
	RepositoryRoot   string     `json:"repository_root"`
	TrackedDigest    string     `json:"tracked_digest"`
	UntrackedDigest  string     `json:"untracked_digest"`
	BeforeDigest     string     `json:"before_digest"`
	AfterDigest      string     `json:"after_digest"`
	OutputPath       string     `json:"output_path"`
	OutputDigest     string     `json:"output_digest"`
	RepositoryWrites int        `json:"repository_writes"`
	Coordinate       coordinate `json:"coordinate"`
}
type evidenceReceipt struct {
	Schema                          string                    `json:"schema"`
	Provider                        string                    `json:"provider"`
	SourcePath                      string                    `json:"source_path"`
	SourceBytesDigest               string                    `json:"source_bytes_digest"`
	SourceGraphDigest               string                    `json:"source_graph_digest"`
	ArtifactPath                    string                    `json:"artifact_path"`
	ArtifactBytesDigest             string                    `json:"artifact_bytes_digest"`
	Operation                       string                    `json:"operation"`
	RequestStatus                   string                    `json:"request_status"`
	Procedure                       string                    `json:"procedure"`
	ObservationPath                 string                    `json:"observation_path,omitempty"`
	ObservationBundleDigest         string                    `json:"observation_bundle_digest,omitempty"`
	ObservationBundleRawDigest      string                    `json:"observation_bundle_raw_digest,omitempty"`
	ObservationBundleRaw            []byte                    `json:"observation_bundle_raw,omitempty"`
	ValidatorContractPath           string                    `json:"validator_contract_path"`
	ValidatorContractDigest         string                    `json:"validator_contract_digest"`
	ValidatorContractRaw            []byte                    `json:"validator_contract_raw"`
	StructuralManifestPath          string                    `json:"structural_manifest_path"`
	StructuralManifestDigest        string                    `json:"structural_manifest_digest"`
	StructuralManifestRaw           []byte                    `json:"structural_manifest_raw"`
	Observations                    []observationReceipt      `json:"observations"`
	StructuralContradictions        []structuralContradiction `json:"structural_contradictions,omitempty"`
	StructuralInventoryTotal        int                       `json:"structural_inventory_total"`
	SemanticOccurrenceNumerator     int                       `json:"semantic_occurrence_numerator"`
	SemanticOccurrenceDenominator   int                       `json:"semantic_occurrence_denominator"`
	RawProvenanceBindingNumerator   int                       `json:"raw_provenance_binding_numerator"`
	RawProvenanceBindingDenominator int                       `json:"raw_provenance_binding_denominator"`
	ObservedPredicate               predicate                 `json:"observed_predicate"`
	ObservedValue                   string                    `json:"observed_value"`
	SemanticEvidenceDigest          string                    `json:"semantic_evidence_digest"`
	Status                          string                    `json:"status"`
	Coordinate                      coordinate                `json:"coordinate"`
	Claims                          []evidenceClaim           `json:"claims"`
	Capability                      capability                `json:"capability"`
	Snapshot                        snapshot                  `json:"snapshot"`
	Digest                          string                    `json:"digest"`
}
type subject struct {
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
	AuthorityCoordinate coordinate `json:"authority_coordinate"`
}
type transition struct {
	Sequence                  int        `json:"sequence"`
	ClaimID                   string     `json:"claim_id"`
	Event                     string     `json:"event"`
	Before                    string     `json:"before"`
	After                     string     `json:"after"`
	Coordinate                coordinate `json:"coordinate"`
	EvidenceDigest            string     `json:"evidence_digest,omitempty"`
	UpstreamEdgeIDs           []string   `json:"upstream_edge_ids,omitempty"`
	UpstreamTransitionDigests []string   `json:"upstream_transition_digests,omitempty"`
	Provenance                string     `json:"provenance"`
	PreviousTransitionDigest  string     `json:"previous_transition_digest,omitempty"`
	TransitionDigest          string     `json:"transition_digest"`
}
type resolution struct {
	ClaimID                string      `json:"claim_id"`
	Axis                   string      `json:"axis"`
	PropositionDigest      string      `json:"proposition_digest"`
	State                  string      `json:"state"`
	Kind                   string      `json:"kind"`
	ObservedEvent          string      `json:"observed_event"`
	Coordinate             coordinate  `json:"coordinate"`
	EvidenceDigest         string      `json:"evidence_digest,omitempty"`
	Provenance             string      `json:"provenance"`
	FailureResponsibility  string      `json:"failure_responsibility"`
	FailureOwnerClaimID    string      `json:"failure_owner_claim_id"`
	MissingEvidenceIDs     []string    `json:"missing_evidence_ids,omitempty"`
	BlockedByClaimIDs      []string    `json:"blocked_by_claim_ids,omitempty"`
	BlockedByEdgeIDs       []string    `json:"blocked_by_edge_ids,omitempty"`
	CausePath              []string    `json:"cause_path"`
	CauseEdgeIDs           []string    `json:"cause_edge_ids"`
	CauseEdgeKinds         []edgeKind  `json:"cause_edge_kinds"`
	CauseTransitionDigests []string    `json:"cause_transition_digests"`
	CauseCoordinate        *coordinate `json:"cause_coordinate"`
}
type truthCase struct {
	Schema                    string   `json:"schema"`
	CaseID                    string   `json:"case_id"`
	Kind                      edgeKind `json:"kind"`
	Direction                 string   `json:"direction"`
	UpstreamState             string   `json:"upstream_state"`
	LocalPredicate            string   `json:"local_predicate"`
	ExpectedState             string   `json:"expected_state"`
	Positive                  bool     `json:"positive"`
	ContradictionObserved     bool     `json:"contradiction_observed"`
	FailureAntecedentObserved bool     `json:"failure_antecedent_observed"`
	SemanticBasis             string   `json:"semantic_basis"`
}
type edgeMetric struct {
	Kind           edgeKind `json:"kind"`
	Eligible       int      `json:"eligible"`
	ObservedCausal int      `json:"observed_causal"`
	Blocking       int      `json:"blocking"`
	Refuting       int      `json:"refuting"`
	Discharge      int      `json:"discharge"`
}
type metrics struct {
	FixedClaimTotal                              int          `json:"fixed_claim_total"`
	DistinctPropositionTotal                     int          `json:"distinct_proposition_total"`
	StructuralContradictionNumerator             int          `json:"structural_contradiction_numerator"`
	StructuralContradictionDenominator           int          `json:"structural_contradiction_denominator"`
	SemanticOccurrenceNumerator                  int          `json:"semantic_occurrence_numerator"`
	SemanticOccurrenceDenominator                int          `json:"semantic_occurrence_denominator"`
	RawProvenanceBindingNumerator                int          `json:"raw_provenance_binding_numerator"`
	RawProvenanceBindingDenominator              int          `json:"raw_provenance_binding_denominator"`
	CommentOnlySemanticPreservationNumerator     int          `json:"comment_only_semantic_preservation_numerator"`
	CommentOnlySemanticPreservationDenominator   int          `json:"comment_only_semantic_preservation_denominator"`
	StructuralContradictionDenominatorCoordinate coordinate   `json:"structural_contradiction_denominator_coordinate"`
	FixedEdgeTotal                               int          `json:"fixed_edge_total"`
	EligibleEdgeTotal                            int          `json:"eligible_edge_total"`
	ObservedCausalEdgeTotal                      int          `json:"observed_causal_edge_total"`
	ShortestPathEdgeUnionTotal                   int          `json:"shortest_path_edge_union_total"`
	ClassifiedClaimTotal                         int          `json:"classified_claim_total"`
	OpenClaimTotal                               int          `json:"open_claim_total"`
	DischargedClaimTotal                         int          `json:"discharged_claim_total"`
	RefutedClaimTotal                            int          `json:"refuted_claim_total"`
	CurrentEvidenceTotal                         int          `json:"current_evidence_total"`
	HistoricalEvidenceTotal                      int          `json:"historical_evidence_total"`
	UnknownEvidenceTotal                         int          `json:"unknown_evidence_total"`
	DirectUnknownClaimTotal                      int          `json:"direct_unknown_claim_total"`
	DependencyBlockedClaimTotal                  int          `json:"dependency_blocked_claim_total"`
	DirectRefutedClaimTotal                      int          `json:"direct_refuted_claim_total"`
	DependencyRefutedClaimTotal                  int          `json:"dependency_refuted_claim_total"`
	DirectDischargedClaimTotal                   int          `json:"direct_discharged_claim_total"`
	DependencyDischargedTotal                    int          `json:"dependency_discharged_claim_total"`
	ObservedBlockingEdgeTotal                    int          `json:"observed_blocking_edge_total"`
	ObservedRefutingEdgeTotal                    int          `json:"observed_refuting_edge_total"`
	ObservedRecoveryEdgeTotal                    int          `json:"observed_recovery_edge_total"`
	MaximumCausePathDepth                        int          `json:"maximum_cause_path_depth"`
	TransitionTotal                              int          `json:"transition_total"`
	AppendOnlyTransitionTotal                    int          `json:"append_only_transition_total"`
	ClassificationBasisPoints                    int          `json:"classification_basis_points"`
	TruthTableCaseTotal                          int          `json:"truth_table_case_total"`
	AuthorityCaseTotal                           int          `json:"authority_case_total"`
	EdgeMetrics                                  []edgeMetric `json:"edge_metrics"`
}
type decision struct {
	Value                       string `json:"value"`
	Resolution                  string `json:"resolution"`
	Reason                      string `json:"reason"`
	SemanticPromotionAuthorized bool   `json:"semantic_promotion_authorized"`
}
type receipt struct {
	Schema                   string          `json:"schema"`
	Scope                    string          `json:"scope"`
	Subject                  subject         `json:"subject"`
	Evidence                 evidenceReceipt `json:"evidence"`
	Graph                    graph           `json:"graph"`
	TruthTable               []truthCase     `json:"truth_table"`
	AuthorityCases           []authorityCase `json:"authority_cases"`
	PriorReceiptDigest       string          `json:"prior_receipt_digest,omitempty"`
	PreviousTransitionDigest string          `json:"previous_transition_digest,omitempty"`
	PriorClaimStates         []string        `json:"prior_claim_states,omitempty"`
	EvidenceDigest           string          `json:"evidence_digest"`
	Transitions              []transition    `json:"transitions"`
	TransitionHeadDigest     string          `json:"transition_head_digest"`
	Resolutions              []resolution    `json:"resolutions"`
	Metrics                  metrics         `json:"metrics"`
	Decision                 decision        `json:"decision"`
	Digest                   string          `json:"digest"`
}

type Judgment struct {
	Schema                           string  `json:"schema"`
	ReceiptDigest                    string  `json:"receipt_digest"`
	Predicate                        string  `json:"predicate"`
	Decision                         string  `json:"decision"`
	Resolution                       string  `json:"resolution"`
	Reason                           string  `json:"reason"`
	Accepted                         bool    `json:"accepted"`
	IndependentReplay                string  `json:"independent_replay"`
	Metrics                          metrics `json:"metrics"`
	ReadOnly                         bool    `json:"read_only"`
	SemanticPromotionAuthorized      bool    `json:"semantic_promotion_authorized"`
	AuthorityResolution              string  `json:"authority_resolution"`
	SourceReconstruction             string  `json:"source_reconstruction"`
	SourceReconstructionNumerator    int     `json:"source_reconstruction_numerator"`
	SourceReconstructionDenominator  int     `json:"source_reconstruction_denominator"`
	ProducerPackageImportNumerator   int     `json:"producer_package_import_numerator"`
	ProducerPackageImportDenominator int     `json:"producer_package_import_denominator"`
	AppendOnlyRecoveryChainTotal     int     `json:"append_only_recovery_chain_total"`
	Digest                           string  `json:"digest"`
}
type reconstructed struct {
	Graph graph
}

func Judge(source []byte, sourcePath string, priorBytes, evidenceBytes, receiptBytes []byte) (Judgment, error) {
	return judgeWithExternalMaterials(source, sourcePath, priorBytes, evidenceBytes, receiptBytes, "", "")
}

// JudgeWithExternalMaterials is the CI entry point. The judge reads the
// contract and structural manifest independently of the producer bundle and
// requires every embedded copy to match those exact bytes.
func JudgeWithExternalMaterials(source []byte, sourcePath string, priorBytes, evidenceBytes, receiptBytes []byte, contractPath, manifestPath string) (Judgment, error) {
	if contractPath == "" || manifestPath == "" {
		return Judgment{}, fmt.Errorf("independent judge requires external validator contract and structural manifest")
	}
	return judgeWithExternalMaterials(source, sourcePath, priorBytes, evidenceBytes, receiptBytes, contractPath, manifestPath)
}

func judgeWithExternalMaterials(source []byte, sourcePath string, priorBytes, evidenceBytes, receiptBytes []byte, contractPath, manifestPath string) (Judgment, error) {
	current, err := reconstruct(source, sourcePath)
	if err != nil {
		return Judgment{}, err
	}
	externalContract, externalContractBytes, err := readValidatorContract(contractPath)
	if err != nil {
		return Judgment{}, err
	}
	externalManifest, externalManifestBytes, err := readStructuralInventoryManifest(manifestPath)
	if err != nil {
		return Judgment{}, err
	}
	if err := validateStructuralInventoryManifest(externalManifest, externalContract, current.Graph); err != nil {
		return Judgment{}, err
	}
	for _, c := range current.Graph.Nodes {
		material, ok := contractClaim(externalContract, c.ActivityName)
		if !ok || !claimIdentityMatchesContract(c, material) {
			return Judgment{}, fmt.Errorf("external validator contract claim inventory does not match source claim %q", c.ActivityName)
		}
	}
	var evidence evidenceReceipt
	if err := json.Unmarshal(evidenceBytes, &evidence); err != nil {
		return Judgment{}, fmt.Errorf("decode raw evidence: %w", err)
	}
	if evidence.ValidatorContractPath != contractPath || evidence.ValidatorContractDigest != digestBytes(externalContractBytes) || !bytes.Equal(evidence.ValidatorContractRaw, externalContractBytes) {
		return Judgment{}, fmt.Errorf("EXTERNAL_VALIDATOR_CONTRACT_MISMATCH: evidence is not bound to judge contract bytes")
	}
	if evidence.StructuralManifestPath != manifestPath || evidence.StructuralManifestDigest != digestBytes(externalManifestBytes) || !bytes.Equal(evidence.StructuralManifestRaw, externalManifestBytes) {
		return Judgment{}, fmt.Errorf("EXTERNAL_STRUCTURAL_MANIFEST_MISMATCH: evidence is not bound to judge manifest bytes")
	}
	if err := validateEvidence(evidence); err != nil {
		return Judgment{}, err
	}
	evidenceSource, err := os.ReadFile(evidence.SourcePath)
	if err != nil {
		return Judgment{}, fmt.Errorf("judge cannot re-observe evidence source: %w", err)
	}
	evidenceGraph, err := reconstruct(evidenceSource, evidence.SourcePath)
	if err != nil {
		return Judgment{}, err
	}
	if evidence.SourcePath != sourcePath || evidence.SourceBytesDigest != digestBytes(evidenceSource) || evidence.SourceGraphDigest != evidenceGraph.Graph.Digest || !reflect.DeepEqual(evidenceGraph.Graph, current.Graph) {
		return Judgment{}, fmt.Errorf("evidence source is not the judge's raw source")
	}
	artifact, err := os.ReadFile(evidence.ArtifactPath)
	if err != nil {
		return Judgment{}, fmt.Errorf("judge cannot re-observe artifact: %w", err)
	}
	if digestBytes(artifact) != evidence.ArtifactBytesDigest {
		return Judgment{}, fmt.Errorf("judge observed artifact bytes digest mismatch")
	}
	if err := validateEvidenceClaims(evidence, current.Graph, artifact); err != nil {
		return Judgment{}, err
	}
	var got receipt
	if err := json.Unmarshal(receiptBytes, &got); err != nil {
		return Judgment{}, fmt.Errorf("decode receipt: %w", err)
	}
	if got.Graph.Digest != current.Graph.Digest || !reflect.DeepEqual(got.Graph, current.Graph) {
		return Judgment{}, fmt.Errorf("receipt graph is not reconstructed from raw source")
	}
	if got.EvidenceDigest != evidence.Digest || !reflect.DeepEqual(got.Evidence, evidence) {
		return Judgment{}, fmt.Errorf("receipt is not bound to raw evidence receipt")
	}
	sourceDigest := digestBytes(source)
	if evidence.ObservationPath != "" {
		if len(evidence.ObservationBundleRaw) == 0 || evidence.ObservationBundleRawDigest == "" || digestBytes(evidence.ObservationBundleRaw) != evidence.ObservationBundleRawDigest {
			return Judgment{}, fmt.Errorf("judge lacks durable target observation bytes")
		}
		var bundle observationBundle
		if err := strictJSON(evidence.ObservationBundleRaw, &bundle); err != nil {
			return Judgment{}, fmt.Errorf("source-bound target bundle decode: %w", err)
		}
		if bundle.SourcePath != sourcePath || bundle.SourceDigest != sourceDigest {
			return Judgment{}, fmt.Errorf("target observation bundle is not bound to judge raw source")
		}
	}
	if got.Subject.SourceDigest != sourceDigest || got.Subject.SourcePath != sourcePath || got.Subject.SemanticDigest != current.Graph.CanonicalIRDigest || got.Subject.Producer != producerID || got.Subject.Consumer != consumerID || got.Subject.MetaOperation != operation || got.Subject.ProofChoice != proof {
		return Judgment{}, fmt.Errorf("receipt subject provenance is invalid")
	}
	expectedAuthority := authorityResolution(evidence)
	if got.Subject.AuthorityResolution != expectedAuthority || got.Subject.ReadOnly != (expectedAuthority == "NET_REPOSITORY_STATE_UNCHANGED") || got.Subject.RepositoryWrites != evidence.Snapshot.RepositoryWrites || !reflect.DeepEqual(got.Subject.AuthorityCoordinate, evidence.Capability.Coordinate) {
		return Judgment{}, fmt.Errorf("receipt subject authority is not independently reproduced")
	}
	var prior *receipt
	if len(priorBytes) > 0 {
		var value receipt
		if err := strictJSON(priorBytes, &value); err != nil {
			return Judgment{}, err
		}
		if err := validatePrior(current, value); err != nil {
			return Judgment{}, err
		}
		prior = &value
		d := receiptDigest(value)
		if got.PriorReceiptDigest != d || got.PreviousTransitionDigest != value.TransitionHeadDigest || !sameStrings(got.PriorClaimStates, statesOf(value.Resolutions)) {
			return Judgment{}, fmt.Errorf("recovery does not bind prior receipt head and states")
		}
		if len(got.Transitions) < len(value.Transitions) || !reflect.DeepEqual(got.Transitions[:len(value.Transitions)], value.Transitions) {
			return Judgment{}, fmt.Errorf("recovery is not append-only")
		}
	}
	states, templates := classify(current.Graph, evidence)
	provenance := fmt.Sprintf("source-semantic:%s|claim-evidence:%s|producer:%s|consumer:%s", current.Graph.CanonicalIRDigest, semanticEvidenceDigest(evidence), producerID, consumerID)
	expectedTransitions := transitionsFor(current.Graph, templates, provenance, prior)
	if !reflect.DeepEqual(got.Transitions, expectedTransitions) {
		return Judgment{}, fmt.Errorf("transition chain cannot be independently reproduced")
	}
	currentOutcomes := templates
	if prior != nil {
		currentOutcomes = expectedTransitions[len(expectedTransitions)-claimTotal:]
	}
	expectedResolutions := buildResolutions(current.Graph, states, currentOutcomes, provenance)
	if !reflect.DeepEqual(got.Resolutions, expectedResolutions) {
		return Judgment{}, fmt.Errorf("resolutions cannot be independently reproduced")
	}
	expectedMetrics := deriveMetrics(current.Graph, states, expectedResolutions, currentOutcomes, evidence, prior != nil)
	if prior != nil {
		expectedMetrics.AppendOnlyTransitionTotal = claimTotal
	}
	if !reflect.DeepEqual(got.Metrics, expectedMetrics) {
		return Judgment{}, fmt.Errorf("metrics cannot be independently reproduced")
	}
	expectedDecision := decisionFor(states, evidence, prior != nil)
	if !reflect.DeepEqual(got.Decision, expectedDecision) {
		return Judgment{}, fmt.Errorf("decision cannot be independently reproduced")
	}
	if got.TruthTable == nil || !reflect.DeepEqual(got.TruthTable, truthTable()) {
		return Judgment{}, fmt.Errorf("truth table is not independently reproduced")
	}
	if got.AuthorityCases == nil || !reflect.DeepEqual(got.AuthorityCases, authorityCases()) {
		return Judgment{}, fmt.Errorf("authority resolution cases are not independently reproduced")
	}
	if err := validateAuthorityCases(got.AuthorityCases); err != nil {
		return Judgment{}, err
	}
	if err := validateTruthTable(got.TruthTable); err != nil {
		return Judgment{}, err
	}
	if got.TransitionHeadDigest != got.Transitions[len(got.Transitions)-1].TransitionDigest || receiptDigest(got) != got.Digest {
		return Judgment{}, fmt.Errorf("receipt digest or transition head is invalid")
	}
	judgment := Judgment{Schema: judgmentSchema, ReceiptDigest: got.Digest, Predicate: string(evidence.ObservedPredicate), Decision: expectedDecision.Value, Resolution: expectedDecision.Resolution, Reason: expectedDecision.Reason, Accepted: true, IndependentReplay: "RAW_GOOO_PARSE_LOWER_SOURCE_TARGET_REOBSERVE_AND_TRANSITION_REDERIVED", Metrics: expectedMetrics, ReadOnly: got.Subject.ReadOnly && got.Subject.RepositoryWrites == 0, SemanticPromotionAuthorized: false, AuthorityResolution: got.Subject.AuthorityResolution, SourceReconstruction: "syntax.ParseFile->bidir.Lower->semantic.IR", SourceReconstructionNumerator: 1, SourceReconstructionDenominator: 1, ProducerPackageImportNumerator: 0, ProducerPackageImportDenominator: 1, AppendOnlyRecoveryChainTotal: boolInt(prior != nil)}
	judgment.Digest = digestJSON(judgment)
	return judgment, nil
}

func reconstruct(source []byte, sourcePath string) (reconstructed, error) {
	file, diagnostics := syntax.ParseFile(sourcePath, string(source))
	if file == nil || diagnostics.HasErrors() {
		return reconstructed{}, fmt.Errorf("judge parse failed: %s", diagnostics.Error())
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		return reconstructed{}, err
	}
	if err := ir.Validate(); err != nil {
		return reconstructed{}, err
	}
	activities := map[string]semantic.Node{}
	for _, node := range ir.Graph.Nodes() {
		if node.Kind == semantic.Activity {
			activities[node.Name] = node
		}
	}
	generatedBy, usedBy := map[string]string{}, map[string][]string{}
	for _, fact := range ir.Graph.AllFacts() {
		switch fact.Predicate {
		case semantic.WasGeneratedBy:
			generatedBy[fact.Subject.String()] = fact.Object.String()
		case semantic.Used:
			usedBy[fact.Subject.String()] = append(usedBy[fact.Subject.String()], fact.Object.String())
		}
	}
	claims, activityIndex := []claim{}, map[string]int{}
	activityNames := map[string]bool{}
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if !ok {
			continue
		}
		if activityNames[activity.Name] {
			return reconstructed{}, fmt.Errorf("duplicate activity %q is not a unique semantic occurrence", activity.Name)
		}
		activityNames[activity.Name] = true
		node, ok := activities[activity.Name]
		if !ok || node.ValueProgram == "" {
			return reconstructed{}, fmt.Errorf("judge cannot bind activity")
		}
		inputs := append([]string(nil), usedBy[node.ID.String()]...)
		sort.Strings(inputs)
		output := ""
		for entityID, activityID := range generatedBy {
			if activityID == node.ID.String() {
				output = entityID
				break
			}
		}
		artifact := "gooo://claim-dependency/artifact/" + strings.ToLower(activity.Name)
		proposition := fmt.Sprintf("execute(activity=%s,inputs=[%s],output=%s,artifact=%s,value=%s)", node.ID.String(), strings.Join(inputs, ","), output, artifact, node.ValueProgram)
		activityIndex[node.ID.String()] = len(claims)
		claims = append(claims, claim{len(claims) + 1, strings.ToLower(activity.Name), node.ID.String(), node.ID.String(), activity.Name, proposition, digestBytes([]byte(proposition)), target{Inputs: inputs, Output: output, Artifact: artifact}, node.ValueProgram, producerID, consumerID, operation, proof, coordinate{"CLAIM", activity.Name, "NORMALIZED_EXECUTION_PROPOSITION"}})
	}
	if len(claims) != claimTotal {
		return reconstructed{}, fmt.Errorf("judge reconstructed %d claims", len(claims))
	}
	seen := map[string]bool{}
	claimIDs := map[string]bool{}
	for _, c := range claims {
		seen[c.PropositionDigest] = true
	}
	if len(seen) != claimTotal {
		return reconstructed{}, fmt.Errorf("judge found non-distinct propositions")
	}
	type candidate struct {
		from, to int
		kind     edgeKind
	}
	candidates := []candidate{}
	for downstreamID, entities := range usedBy {
		to, ok := activityIndex[downstreamID]
		if !ok {
			continue
		}
		for _, entityID := range entities {
			upstreamID, ok := generatedBy[entityID]
			if !ok {
				continue
			}
			from, ok := activityIndex[upstreamID]
			if !ok || from == to {
				continue
			}
			kind, ok := edgeKind(claims[to].ValueProgram)
			if !ok {
				return reconstructed{}, fmt.Errorf("judge found untyped edge")
			}
			candidates = append(candidates, candidate{from, to, kind})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].from != candidates[j].from {
			return candidates[i].from < candidates[j].from
		}
		if candidates[i].to != candidates[j].to {
			return candidates[i].to < candidates[j].to
		}
		return candidates[i].kind < candidates[j].kind
	})
	if len(candidates) != edgeTotal {
		return reconstructed{}, fmt.Errorf("judge reconstructed %d edges", len(candidates))
	}
	edges := make([]edge, len(candidates))
	for i, c := range candidates {
		edges[i] = edge{fmt.Sprintf("E%02d", i+1), claims[c.from].ClaimID, claims[c.to].ClaimID, c.kind, activationPredicate(c.kind), "prov:wasGeneratedBy + prov:used + source-derived value-program edge predicate"}
	}
	result := graph{"gooo.meta.claim-dependency-graph/v3", "CANONICAL_IR_FROM_SYNTAX_PARSE_AND_BIDIR_LOWER", "CLOSED_WORLD_SOURCE_RECONSTRUCTED", prefixedDigest(ir.StableHash()), claimTotal, edgeTotal, claims, edges, ""}
	result.Digest = graphDigest(result)
	return reconstructed{result}, nil
}

func activationPredicate(kind edgeKind) string {
	switch kind {
	case requires:
		return "UPSTREAM_DISCHARGED_AND_LOCAL_EVIDENCE"
	case contradicts:
		return "UPSTREAM_DISCHARGED_PROPOSITION"
	case failureEntailment:
		return "OBSERVED_FAILURE_ANTECEDENT_AND_UPSTREAM_REFUTED"
	default:
		return "UPSTREAM_UNKNOWN_BLOCKS_ONLY"
	}
}

func edgeKind(program string) (edgeKind, bool) {
	if !strings.HasPrefix(program, "claim.edge:") {
		return "", false
	}
	value := strings.TrimPrefix(program, "claim.edge:")
	if i := strings.IndexByte(value, '|'); i >= 0 {
		value = value[:i]
	}
	switch value {
	case "supports":
		return supports, true
	case "requires":
		return requires, true
	case "contradicts":
		return contradicts, true
	case "failure-entailment":
		return failureEntailment, true
	}
	return "", false
}
func truthTable() []truthCase {
	return []truthCase{
		{Schema: "gooo.meta.claim-dependency-truth-table/v1", CaseID: "SUPPORTS-POSITIVE", Kind: supports, Direction: "direction-matching", UpstreamState: "DISCHARGED", LocalPredicate: "EVIDENCE_ACCEPTED", ExpectedState: "OPEN", Positive: true, SemanticBasis: "support never discharges or refutes by itself"},
		{Schema: "gooo.meta.claim-dependency-truth-table/v1", CaseID: "SUPPORTS-REVERSED", Kind: supports, Direction: "reversed-direction", UpstreamState: "REFUTED", LocalPredicate: "UNKNOWN", ExpectedState: "OPEN", Positive: false, SemanticBasis: "support does not refute"},
		{Schema: "gooo.meta.claim-dependency-truth-table/v1", CaseID: "REQUIRES-POSITIVE", Kind: requires, Direction: "direction-matching", UpstreamState: "DISCHARGED", LocalPredicate: "EVIDENCE_ACCEPTED", ExpectedState: "DISCHARGED", Positive: true, SemanticBasis: "upstream and local requirement hold"},
		{Schema: "gooo.meta.claim-dependency-truth-table/v1", CaseID: "REQUIRES-UNKNOWN", Kind: requires, Direction: "direction-matching", UpstreamState: "DISCHARGED", LocalPredicate: "UNKNOWN", ExpectedState: "OPEN", Positive: true, SemanticBasis: "local requirement evidence is unknown"},
		{Schema: "gooo.meta.claim-dependency-truth-table/v1", CaseID: "CONTRADICTS-POSITIVE", Kind: contradicts, Direction: "direction-matching", UpstreamState: "DISCHARGED", LocalPredicate: "UNKNOWN", ExpectedState: "REFUTED", Positive: true, ContradictionObserved: true, SemanticBasis: "established proposition and observed contradiction agree in the declared direction"},
		{Schema: "gooo.meta.claim-dependency-truth-table/v1", CaseID: "CONTRADICTS-REVERSED", Kind: contradicts, Direction: "reversed-direction", UpstreamState: "DISCHARGED", LocalPredicate: "UNKNOWN", ExpectedState: "OPEN", Positive: false, ContradictionObserved: true, SemanticBasis: "reversed contradiction direction cannot refute"},
		{Schema: "gooo.meta.claim-dependency-truth-table/v1", CaseID: "FAILURE_ENTAILMENT-POSITIVE", Kind: failureEntailment, Direction: "direction-matching", UpstreamState: "REFUTED", LocalPredicate: "UNKNOWN", ExpectedState: "REFUTED", Positive: true, FailureAntecedentObserved: true, SemanticBasis: "declared failure antecedent is observed and entails target failure"},
		{Schema: "gooo.meta.claim-dependency-truth-table/v1", CaseID: "FAILURE_ENTAILMENT-UNKNOWN", Kind: failureEntailment, Direction: "direction-matching", UpstreamState: "REFUTED", LocalPredicate: "UNKNOWN", ExpectedState: "OPEN", Positive: true, FailureAntecedentObserved: false, SemanticBasis: "upstream refutation alone does not activate failure entailment"},
	}
}

type relationOutcome string

const (
	relationOpen       relationOutcome = "OPEN"
	relationDischarged relationOutcome = "DISCHARGED"
	relationRefuted    relationOutcome = "REFUTED"
)

func edgeRelation(kind edgeKind, upstreamState string, local predicate, directionMatches, contradictionObserved, failureAntecedentObserved bool) relationOutcome {
	if !directionMatches {
		return relationOpen
	}
	switch kind {
	case requires:
		if upstreamState == "DISCHARGED" && local == accepted {
			return relationDischarged
		}
	case contradicts:
		if upstreamState == "DISCHARGED" && contradictionObserved {
			return relationRefuted
		}
	case failureEntailment:
		if upstreamState == "REFUTED" && failureAntecedentObserved {
			return relationRefuted
		}
	case supports:
		// SUPPORTS blocks unresolved claims but never discharges or refutes.
	}
	return relationOpen
}

func validateTruthTable(cases []truthCase) error {
	if len(cases) != 2*4 {
		return fmt.Errorf("truth table has %d cases, want 8", len(cases))
	}
	seen := map[edgeKind]int{}
	for _, test := range cases {
		actual := edgeRelation(test.Kind, test.UpstreamState, predicate(test.LocalPredicate), test.Positive, test.ContradictionObserved, test.FailureAntecedentObserved)
		if string(actual) != test.ExpectedState {
			return fmt.Errorf("truth table case %q computed %s, expected %s", test.CaseID, actual, test.ExpectedState)
		}
		seen[test.Kind]++
	}
	for _, kind := range []edgeKind{supports, requires, contradicts, failureEntailment} {
		if seen[kind] != 2 {
			return fmt.Errorf("truth table edge kind %s has %d cases", kind, seen[kind])
		}
	}
	return nil
}
func validateEvidence(value evidenceReceipt) error {
	if value.Schema != "gooo.meta.claim-dependency-evidence/v2" || (value.Status != "CURRENT_EVIDENCE" && value.Status != "UNKNOWN") || value.Provider == "" || value.SourcePath == "" || value.SourceBytesDigest == "" || value.SourceGraphDigest == "" || value.ArtifactPath == "" || value.ArtifactBytesDigest == "" || value.Digest == "" || value.RequestStatus != "CLAIMED_INPUT" || value.Procedure != evidenceProcedure {
		return fmt.Errorf("raw evidence identity invalid")
	}
	if evidenceDigest(value) != value.Digest {
		return fmt.Errorf("raw evidence digest invalid")
	}
	if value.SemanticEvidenceDigest != semanticEvidenceDigest(value) {
		return fmt.Errorf("semantic evidence projection digest is invalid")
	}
	if value.Snapshot.BeforeDigest == "" || value.Snapshot.AfterDigest == "" || value.Snapshot.OutputPath == "" || value.Snapshot.RepositoryWrites != 0 || value.Snapshot.BeforeDigest != value.Snapshot.AfterDigest || value.Capability.Status != "CURRENT_EVIDENCE" || value.Capability.Toolchain.Name != "go" || value.Capability.Toolchain.Version != "go1.27.0" {
		return fmt.Errorf("raw evidence effects or capability invalid")
	}
	if capabilityDigest(value.Capability) != value.Capability.Digest {
		return fmt.Errorf("capability digest invalid")
	}
	sourceBytes, err := os.ReadFile(value.SourcePath)
	if err != nil {
		return fmt.Errorf("judge cannot re-observe evidence source: %w", err)
	}
	sourceGraph, err := reconstruct(sourceBytes, value.SourcePath)
	if err != nil {
		return err
	}
	if digestBytes(sourceBytes) != value.SourceBytesDigest || sourceGraph.Graph.Digest != value.SourceGraphDigest {
		return fmt.Errorf("raw evidence source bytes or graph digest changed")
	}
	var contract validatorContract
	var manifest structuralInventoryManifest
	contractComplete := value.ValidatorContractPath != "" && value.ValidatorContractDigest != "" && len(value.ValidatorContractRaw) != 0
	manifestComplete := value.StructuralManifestPath != "" && value.StructuralManifestDigest != "" && len(value.StructuralManifestRaw) != 0
	if !contractComplete || !manifestComplete {
		if contractComplete || manifestComplete || value.ObservationPath != "" {
			return fmt.Errorf("current observation lacks durable external validator materials")
		}
	} else {
		if digestBytes(value.ValidatorContractRaw) != value.ValidatorContractDigest || strictJSON(value.ValidatorContractRaw, &contract) != nil {
			return fmt.Errorf("external validator contract bytes are invalid")
		}
		if _, err := readValidatorContract(value.ValidatorContractPath); err != nil {
			return fmt.Errorf("external validator contract path is not readable: %w", err)
		}
		if err := validateContractForGraph(contract, sourceGraph.Graph); err != nil {
			return err
		}
		if digestBytes(value.StructuralManifestRaw) != value.StructuralManifestDigest || strictJSON(value.StructuralManifestRaw, &manifest) != nil || validateStructuralInventoryManifest(manifest, contract, sourceGraph.Graph) != nil || value.StructuralInventoryTotal != len(manifest.EligibleClaimIDs) {
			return fmt.Errorf("external structural manifest binding is invalid")
		}
	}
	artifactBytes, err := os.ReadFile(value.ArtifactPath)
	if err != nil {
		return fmt.Errorf("judge cannot re-observe artifact: %w", err)
	}
	if digestBytes(artifactBytes) != value.ArtifactBytesDigest {
		return fmt.Errorf("raw evidence artifact bytes digest changed")
	}
	observations := []observationReceipt(nil)
	var observedBundle observationBundle
	if value.ObservationPath != "" {
		if len(value.ObservationBundleRaw) == 0 || value.ObservationBundleRawDigest == "" || digestBytes(value.ObservationBundleRaw) != value.ObservationBundleRawDigest {
			return fmt.Errorf("raw evidence lacks durable target bundle")
		}
		if err := strictJSON(value.ObservationBundleRaw, &observedBundle); err != nil {
			return fmt.Errorf("target observation bundle decode: %w", err)
		}
		if err := validateObservationBundle(observedBundle, value.SourcePath, sourceBytes, value.ArtifactPath, artifactBytes, sourceGraph.Graph); err != nil {
			return err
		}
		if observedBundle.Digest != value.ObservationBundleDigest || !reflect.DeepEqual(observedBundle.Observations, value.Observations) || !reflect.DeepEqual(observedBundle.StructuralContradictions, value.StructuralContradictions) || observedBundle.StructuralInventoryTotal != value.StructuralInventoryTotal || observedBundle.SemanticOccurrenceNumerator != value.SemanticOccurrenceNumerator || observedBundle.SemanticOccurrenceDenominator != value.SemanticOccurrenceDenominator || observedBundle.RawProvenanceBindingNumerator != value.RawProvenanceBindingNumerator || observedBundle.RawProvenanceBindingDenominator != value.RawProvenanceBindingDenominator {
			return fmt.Errorf("embedded target observation bundle differs from raw bundle")
		}
		if observedBundle.ContractPath != value.ValidatorContractPath || observedBundle.ContractDigest != value.ValidatorContractDigest || !bytes.Equal(observedBundle.ContractRaw, value.ValidatorContractRaw) || observedBundle.StructuralManifestPath != value.StructuralManifestPath || observedBundle.StructuralManifestDigest != value.StructuralManifestDigest || !bytes.Equal(observedBundle.StructuralManifestRaw, value.StructuralManifestRaw) {
			return fmt.Errorf("embedded validator materials differ from evidence external materials")
		}
		observations = observedBundle.Observations
	} else if len(value.Observations) != 0 || len(value.StructuralContradictions) != 0 || value.ObservationBundleDigest != "" || value.ObservationBundleRawDigest != "" || len(value.ObservationBundleRaw) != 0 || value.StructuralInventoryTotal != len(manifest.EligibleClaimIDs) || value.SemanticOccurrenceNumerator != 0 || value.RawProvenanceBindingNumerator != 0 {
		return fmt.Errorf("raw evidence has observations without a raw bundle")
	}
	expectedStatus := "UNKNOWN"
	if hasClaimOrEdgeObservation(observations) {
		expectedStatus = "CURRENT_EVIDENCE"
	}
	if value.Status != expectedStatus || value.ObservedPredicate != summaryPredicate(observations) {
		return fmt.Errorf("raw evidence status or predicate is not derived from claim-scoped observations")
	}
	expectedValue := fmt.Sprintf("observation:ABSENT|stage:OBSERVE|step:current-evidence-provider|reason:EXTERNAL_TARGET_OBSERVATION_MISSING|artifact_path_digest:%s|artifact_bytes_digest:%s", digestBytes([]byte(value.ArtifactPath)), digestBytes(artifactBytes))
	if value.ObservationPath != "" {
		expectedValue = observationBundleValue(observedBundle)
	}
	if value.ObservedValue != expectedValue {
		return fmt.Errorf("raw evidence value is not derived from target observations")
	}
	return nil
}
func validateEvidenceClaims(value evidenceReceipt, g graph, artifactBytes []byte) error {
	if len(value.Claims) != g.NodeTotal {
		return fmt.Errorf("raw evidence claim denominator mismatch")
	}
	for i, c := range value.Claims {
		if c.ClaimID != g.Nodes[i].ClaimID || c.PropositionDigest != g.Nodes[i].PropositionDigest || c.Digest == "" || evidenceClaimDigest(c) != c.Digest {
			return fmt.Errorf("raw evidence claim %d is invalid", i+1)
		}
		expected := evidenceClaim{ClaimID: g.Nodes[i].ClaimID, PropositionDigest: g.Nodes[i].PropositionDigest, ObservedPredicate: unknown, ObservedValue: absentClaimValue(g.Nodes[i]), Status: "UNKNOWN", Coordinate: coordinate{"OBSERVE", g.Nodes[i].ActivityName, "CLAIM_OBSERVATION_MISSING"}}
		if observed, ok := claimObservationFor(g.Nodes[i], value.Observations); ok {
			expected.ObservedPredicate = observed.ObservedPredicate
			expected.ObservedValue = observed.ObservedValue
			expected.Status = "CURRENT_EVIDENCE"
			expected.Coordinate = observed.Coordinate
		}
		expected.Digest = evidenceClaimDigest(expected)
		if !reflect.DeepEqual(c, expected) {
			return fmt.Errorf("raw evidence claim %d is not derived from its exact observation binding", i+1)
		}
	}
	_ = artifactBytes
	return nil
}

func validateObservationBundle(bundle observationBundle, sourcePath string, source []byte, artifactPath string, artifact []byte, g graph) error {
	if bundle.Schema != observationBundleSchema || bundle.Provider == "" || bundle.SourcePath != sourcePath || bundle.SourceDigest != digestBytes(source) || bundle.ArtifactPath != artifactPath || bundle.ArtifactBytesDigest != digestBytes(artifact) || bundle.ContractPath == "" || bundle.ContractDigest == "" || len(bundle.ContractRaw) == 0 || bundle.StructuralManifestPath == "" || bundle.StructuralManifestDigest == "" || len(bundle.StructuralManifestRaw) == 0 || bundle.Profile == "" || bundle.Digest == "" || len(bundle.Observations) == 0 {
		return fmt.Errorf("target observation bundle identity or target binding is invalid")
	}
	if digestBytes(bundle.ContractRaw) != bundle.ContractDigest {
		return fmt.Errorf("validator contract bytes changed")
	}
	var contract validatorContract
	if err := strictJSON(bundle.ContractRaw, &contract); err != nil {
		return fmt.Errorf("embedded validator contract decode: %w", err)
	}
	if err := validateValidatorContract(contract); err != nil {
		return err
	}
	if digestBytes(bundle.StructuralManifestRaw) != bundle.StructuralManifestDigest {
		return fmt.Errorf("structural inventory manifest bytes changed")
	}
	var manifest structuralInventoryManifest
	if err := strictJSON(bundle.StructuralManifestRaw, &manifest); err != nil {
		return fmt.Errorf("embedded structural inventory manifest decode: %w", err)
	}
	if err := validateStructuralInventoryManifest(manifest, contract, g); err != nil {
		return err
	}
	for _, c := range g.Nodes {
		material, ok := contractClaim(contract, c.ActivityName)
		if !ok || !claimIdentityMatchesContract(c, material) {
			return fmt.Errorf("embedded validator contract claim inventory does not match source graph")
		}
	}
	failure := failureReceipt{}
	if bundle.FailureReceiptPath != "" {
		if len(bundle.FailureReceiptRaw) == 0 || digestBytes(bundle.FailureReceiptRaw) != bundle.FailureReceiptDigest {
			return fmt.Errorf("failure receipt bytes changed")
		}
		if err := strictJSON(bundle.FailureReceiptRaw, &failure); err != nil {
			return fmt.Errorf("failure receipt decode: %w", err)
		}
		if err := validateFailureReceipt(failure, sourcePath, source, artifactPath, artifact, g); err != nil {
			return err
		}
	}
	if observationBundleDigest(bundle) != bundle.Digest {
		return fmt.Errorf("target observation bundle digest is invalid")
	}
	seen := map[string]bool{}
	for _, observation := range bundle.Observations {
		if err := validateObservation(observation, artifactPath, artifact, g, contract, failure, bundle.FailureReceiptPath != ""); err != nil {
			return err
		}
		key := observation.Binding + "|" + observation.ClaimID + "|" + observation.EdgeID
		if seen[key] {
			return fmt.Errorf("target observation bundle has duplicate binding %q", key)
		}
		seen[key] = true
	}
	expectedStructural, err := deriveStructuralInventory(g, contract, artifact, artifactPath)
	if err != nil {
		if !strings.HasPrefix(err.Error(), "TARGET_SYNTAX_OR_LOWER_INVALID") && !strings.HasPrefix(err.Error(), "TARGET_ACTIVITY_OCCURRENCE") {
			return err
		}
		expectedStructural = []structuralContradiction{}
	}
	semanticOccurrences, rawProvenanceBindings := 0, 0
	semanticOccurrenceDigests := map[string]bool{}
	for _, c := range g.Nodes {
		if occurrence, _, occurrenceErr := canonicalTargetOccurrence(artifact, artifactPath, c); occurrenceErr == nil {
			semanticOccurrences++
			rawProvenanceBindings++
			if occurrence.SemanticDigest == "" || semanticOccurrenceDigests[occurrence.SemanticDigest] {
				return fmt.Errorf("target occurrence semantic digest is not unique")
			}
			semanticOccurrenceDigests[occurrence.SemanticDigest] = true
		}
	}
	if bundle.StructuralInventoryTotal != len(manifest.EligibleClaimIDs) {
		return fmt.Errorf("structural inventory denominator is not external: got=%d want=%d", bundle.StructuralInventoryTotal, len(manifest.EligibleClaimIDs))
	}
	if bundle.SemanticOccurrenceNumerator != semanticOccurrences || bundle.SemanticOccurrenceDenominator != g.NodeTotal || bundle.RawProvenanceBindingNumerator != rawProvenanceBindings || bundle.RawProvenanceBindingDenominator != g.NodeTotal {
		return fmt.Errorf("target occurrence metrics are not re-derived from canonical target occurrences")
	}
	if err := validateStructuralInventory(bundle.StructuralContradictions, expectedStructural); err != nil {
		return err
	}
	if err := validateStructuralInventoryManifestRows(bundle.StructuralContradictions, manifest); err != nil {
		return err
	}
	return nil
}

func validateObservation(value observationReceipt, artifactPath string, artifact []byte, g graph, contract validatorContract, failure failureReceipt, hasFailure bool) error {
	if value.Schema != observationSchema || value.Provider == "" || value.TargetPath != artifactPath || value.Procedure == "" || value.ProcedureDigest != observationProcedureDigest(value.Procedure, value.ClaimID, value.PropositionDigest, value.EdgeID, value.Occurrence) || value.RawProvenanceDigest != rawProvenanceDigest(value.Occurrence, value.TargetBytesDigest) || value.OutputDigest != digestBytes([]byte(value.Output)) || value.Coordinate.Stage == "" || value.Digest == "" {
		return fmt.Errorf("target observation identity or target binding is invalid")
	}
	if value.TargetBytesDigest != digestBytes(artifact) || (value.ComparisonResult != "MATCH" && value.ComparisonResult != "MISMATCH") {
		return fmt.Errorf("target observation comparison is invalid")
	}
	if value.ComparisonResult == "MATCH" && value.ExpectedValue != value.ObservedValue || value.ComparisonResult == "MISMATCH" && value.ExpectedValue == value.ObservedValue {
		return fmt.Errorf("target observation comparison result does not match values")
	}
	switch value.Binding {
	case "CLAIM":
		i := indexOf(value.ClaimID, g)
		if i < 0 || value.PropositionDigest != g.Nodes[i].PropositionDigest || !reflect.DeepEqual(value.Target, g.Nodes[i].Target) || value.ExpectedPredicate != accepted || value.ObservedPredicate != accepted || value.ComparisonResult != "MATCH" {
			return fmt.Errorf("claim-scoped target observation is not bound to its claim")
		}
		material, ok := contractClaim(contract, g.Nodes[i].ActivityName)
		occurrence, row, occurrenceErr := canonicalTargetOccurrence(artifact, artifactPath, g.Nodes[i])
		expected := claimObservationMaterial(g.Nodes[i], material, artifactPath, digestBytes(artifact), occurrence.RawRowDigest)
		if occurrenceErr != nil || !ok || !claimIdentityMatchesContract(g.Nodes[i], material) || g.Nodes[i].ValueProgram != material.ExpectedValueProgram || artifactPath != contract.ExpectedArtifactPath || digestBytes(artifact) != contract.ExpectedArtifactDigest || occurrence.RawRowDigest != material.TargetRowDigest || !reflect.DeepEqual(value.Occurrence, occurrence) || value.ExpectedValue != expected || value.ObservedValue != expected || value.OutputDigest != digestBytes(row) {
			return fmt.Errorf("claim observation does not match external validator material")
		}
	case "EDGE":
		i := indexOfEdge(value.EdgeID, g)
		if i < 0 {
			return fmt.Errorf("edge-scoped target observation references an unknown edge")
		}
		ed := g.Edges[i]
		to := indexOf(ed.ToClaimID, g)
		if to < 0 || value.FromClaimID != ed.FromClaimID || value.ToClaimID != ed.ToClaimID || value.EdgeKind != ed.Kind || value.ClaimID != ed.ToClaimID || value.PropositionDigest != g.Nodes[to].PropositionDigest || !reflect.DeepEqual(value.Target, g.Nodes[to].Target) || value.ExpectedPredicate != value.ObservedPredicate || value.ComparisonResult != "MISMATCH" || (ed.Kind == contradicts && value.ObservedPredicate != explicitContradiction) || (ed.Kind == failureEntailment && value.ObservedPredicate != failureAntecedent) {
			return fmt.Errorf("edge-scoped target observation is not bound to its edge")
		}
		from := indexOf(ed.FromClaimID, g)
		if from < 0 {
			return fmt.Errorf("edge observation references an unknown upstream claim")
		}
		fromContract, fromOK := contractClaim(contract, g.Nodes[from].ActivityName)
		toContract, toOK := contractClaim(contract, g.Nodes[to].ActivityName)
		if !fromOK || !toOK || artifactPath != contract.ExpectedArtifactPath || digestBytes(artifact) != contract.ExpectedArtifactDigest {
			return fmt.Errorf("edge observation does not match external validator material")
		}
		fromOccurrence, _, fromErr := canonicalTargetOccurrence(artifact, artifactPath, g.Nodes[from])
		toOccurrence, _, toErr := canonicalTargetOccurrence(artifact, artifactPath, g.Nodes[to])
		fromRowDigest, toRowDigest := fromOccurrence.RawRowDigest, toOccurrence.RawRowDigest
		if artifactPath != contract.ExpectedArtifactPath || digestBytes(artifact) != contract.ExpectedArtifactDigest || fromErr != nil || toErr != nil {
			return fmt.Errorf("edge observation target rows are not externally bound")
		}
		if !reflect.DeepEqual(value.Occurrence, toOccurrence) {
			return fmt.Errorf("edge observation occurrence is not source-bound")
		}
		if ed.Kind == contradicts && (fromRowDigest != fromContract.TargetRowDigest || toContract.AlternateRowDigest == "" || toRowDigest != toContract.AlternateRowDigest) {
			return fmt.Errorf("contradiction edge observation has the wrong direction or value binding")
		}
		if ed.Kind == failureEntailment && (fromContract.AlternateRowDigest == "" || toContract.AlternateRowDigest == "" || fromRowDigest != fromContract.AlternateRowDigest || toRowDigest != toContract.AlternateRowDigest) {
			return fmt.Errorf("failure edge observation has the wrong failure binding")
		}
		expected := edgeMaterial("contract", ed, g, contract, digestBytes(artifact), artifactPath, false)
		observed := edgeTargetMaterial("observed", ed, g, fromRowDigest, toRowDigest, artifactPath, digestBytes(artifact))
		if ed.Kind == contradicts && (value.ExpectedValue != expected || value.ObservedValue != observed || value.Output != "CONTRADICTS_TARGET_VALUE_OPPOSITE") {
			return fmt.Errorf("contradiction edge observation is not a structured opposite-value comparison")
		}
		if ed.Kind == failureEntailment && (!hasFailure || failure.EdgeID != ed.EdgeID || failure.ObservedExitCode != 1 || failure.Result != "NONZERO_EXIT" || value.ExpectedValue != expected || value.ObservedValue != observed+"|exit="+strconv.Itoa(failure.ObservedExitCode)+"|result="+failure.Result || value.Output != digestBytes(append(failure.Stdout, failure.Stderr...))) {
			return fmt.Errorf("failure edge observation lacks an exact non-zero failure receipt")
		}
	case "UNRELATED_ARTIFACT":
		if value.ClaimID != "" || value.EdgeID != "" || value.PropositionDigest != "" || value.ObservedPredicate != unknown || value.ExpectedPredicate != unknown {
			return fmt.Errorf("unrelated observation carries a claim or edge binding")
		}
	default:
		return fmt.Errorf("unknown target observation binding %q", value.Binding)
	}
	if observationDigest(value) != value.Digest {
		return fmt.Errorf("target observation digest is invalid")
	}
	return nil
}

func readValidatorContract(path string) (validatorContract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return validatorContract{}, fmt.Errorf("validator contract: %w", err)
	}
	var value validatorContract
	if err := strictJSON(data, &value); err != nil {
		return validatorContract{}, fmt.Errorf("validator contract decode: %w", err)
	}
	if value.Schema != validatorContractSchema || value.ContractID == "" || value.ExpectedArtifactPath == "" || value.ExpectedArtifactDigest == "" || len(value.Claims) != claimTotal {
		return validatorContract{}, fmt.Errorf("validator contract identity or denominator is invalid")
	}
	seen := map[string]bool{}
	claimIDs := map[string]bool{}
	targets := map[string]bool{}
	procedures := map[string]bool{}
	rowDigests := map[string]bool{}
	for _, c := range value.Claims {
		if c.ClaimID == "" || c.PropositionDigest == "" || c.ProcedureID == "" || c.TargetRowDigest == "" || c.ExpectedMaterialDigest == "" || c.ActivityName == "" || c.ExpectedValueProgram == "" || c.ExpectedTarget.Artifact == "" || c.ExpectedTarget.Output == "" || seen[c.ActivityName] || claimIDs[c.ClaimID] || procedures[c.ProcedureID] {
			return validatorContract{}, fmt.Errorf("validator contract claim material is invalid")
		}
		seen[c.ActivityName] = true
		claimIDs[c.ClaimID] = true
		procedures[c.ProcedureID] = true
		if c.ProcedureID != validatorProcedureID(c.ActivityName) {
			return validatorContract{}, fmt.Errorf("VALIDATOR_CONTRACT_PROCEDURE_RELABEL: activity=%s", c.ActivityName)
		}
		if c.ExpectedMaterialDigest != validatorExpectedMaterialDigest(c) || rowDigests[c.TargetRowDigest] {
			return validatorContract{}, fmt.Errorf("validator contract expected material digest is invalid")
		}
		rowDigests[c.TargetRowDigest] = true
		targetKey := strings.Join(c.ExpectedTarget.Inputs, ",") + "|" + c.ExpectedTarget.Output + "|" + c.ExpectedTarget.Artifact
		if targets[targetKey] {
			return validatorContract{}, fmt.Errorf("validator contract targets are not distinct")
		}
		targets[targetKey] = true
	}
	return value, nil
}

func readStructuralInventoryManifest(path string) (structuralInventoryManifest, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return structuralInventoryManifest{}, nil, fmt.Errorf("structural inventory manifest: %w", err)
	}
	var value structuralInventoryManifest
	if err := strictJSON(data, &value); err != nil {
		return structuralInventoryManifest{}, nil, fmt.Errorf("structural inventory manifest decode: %w", err)
	}
	return value, data, nil
}

func validateStructuralInventoryManifest(value structuralInventoryManifest, contract validatorContract, g graph) error {
	if value.Schema != structuralManifestSchema || value.ManifestID == "" || value.ContractID != contract.ContractID || len(value.EligibleClaimIDs) == 0 {
		return fmt.Errorf("STRUCTURAL_MANIFEST_IDENTITY_INVALID")
	}
	known := map[string]bool{}
	for _, c := range g.Nodes {
		known[c.ClaimID] = true
	}
	eligible := map[string]bool{}
	for _, claimID := range value.EligibleClaimIDs {
		if claimID == "" || !known[claimID] || eligible[claimID] {
			return fmt.Errorf("STRUCTURAL_MANIFEST_ELIGIBLE_CLAIMS_INVALID: claim=%s", claimID)
		}
		eligible[claimID] = true
	}
	expected := map[string]bool{}
	for _, claimID := range value.ExpectedContradictionClaimIDs {
		if claimID == "" || !eligible[claimID] || expected[claimID] {
			return fmt.Errorf("STRUCTURAL_MANIFEST_EXPECTED_CLAIMS_INVALID: claim=%s", claimID)
		}
		expected[claimID] = true
	}
	return nil
}

func validateStructuralInventoryManifestRows(observed []structuralContradiction, manifest structuralInventoryManifest) error {
	expected := map[string]bool{}
	for _, claimID := range manifest.ExpectedContradictionClaimIDs {
		expected[claimID] = true
	}
	actual := map[string]bool{}
	for _, finding := range observed {
		if actual[finding.ClaimID] {
			return fmt.Errorf("STRUCTURAL_INVENTORY_DUPLICATE: claim=%s", finding.ClaimID)
		}
		actual[finding.ClaimID] = true
	}
	if len(actual) < len(expected) {
		return fmt.Errorf("STRUCTURAL_INVENTORY_MISSING: observed=%d expected=%d", len(actual), len(expected))
	}
	if len(actual) > len(expected) {
		return fmt.Errorf("STRUCTURAL_INVENTORY_ADDITIONAL: observed=%d expected=%d", len(actual), len(expected))
	}
	for claimID := range expected {
		if !actual[claimID] {
			return fmt.Errorf("STRUCTURAL_INVENTORY_REPLACEMENT: claim=%s", claimID)
		}
	}
	for claimID := range actual {
		if !expected[claimID] {
			return fmt.Errorf("STRUCTURAL_INVENTORY_REPLACEMENT: claim=%s", claimID)
		}
	}
	return nil
}

func validatorProcedureID(activity string) string {
	return map[string]string{"Root": "CI_CLAIM_TARGET_ROW_ROOT_V1", "Derived": "CI_CLAIM_TARGET_ROW_DERIVED_V1", "SupportCheck": "CI_CLAIM_TARGET_ROW_SUPPORT_V1", "RequirementCheck": "CI_CLAIM_TARGET_ROW_REQUIREMENT_V1", "ContradictionCheck": "CI_CLAIM_TARGET_ROW_CONTRADICTION_V1", "FailureEntailmentCheck": "CI_CLAIM_TARGET_ROW_FAILURE_V1"}[activity]
}

func strictJSON(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return fmt.Errorf("trailing JSON data: %w", err)
	}
	return nil
}

func validatorExpectedMaterialDigest(c validatorClaim) string {
	return digestBytes([]byte(fmt.Sprintf("claim-contract|claim_id=%s|proposition_digest=%s|procedure_id=%s|target_row_digest=%s|target_inputs=%s|target_output=%s|target_artifact=%s|expected_value_program=%s", c.ClaimID, c.PropositionDigest, c.ProcedureID, c.TargetRowDigest, strings.Join(c.ExpectedTarget.Inputs, ","), c.ExpectedTarget.Output, c.ExpectedTarget.Artifact, c.ExpectedValueProgram)))
}

func claimIdentityMatchesContract(c claim, expected validatorClaim) bool {
	return c.ClaimID == expected.ClaimID && c.PropositionDigest == expected.PropositionDigest && c.ActivityName == expected.ActivityName && reflect.DeepEqual(c.Target, expected.ExpectedTarget)
}

func validateContractForGraph(contract validatorContract, g graph) error {
	for _, c := range g.Nodes {
		material, ok := contractClaim(contract, c.ActivityName)
		if !ok || !claimIdentityMatchesContract(c, material) {
			return fmt.Errorf("external validator contract claim inventory does not match source claim %q", c.ActivityName)
		}
	}
	return nil
}

func claimObservationMaterial(c claim, expected validatorClaim, artifactPath, artifactDigest, rowDigest string) string {
	return fmt.Sprintf("claim-observation|claim_id=%s|proposition_digest=%s|procedure_id=%s|target_inputs=%s|target_output=%s|target_artifact=%s|expected_value_program=%s", c.ClaimID, c.PropositionDigest, expected.ProcedureID, strings.Join(c.Target.Inputs, ","), c.Target.Output, c.Target.Artifact, expected.ExpectedValueProgram)
}

func edgeTargetMaterial(prefix string, ed edge, g graph, fromRowDigest, toRowDigest, artifactPath, artifactDigest string) string {
	return fmt.Sprintf("%s|edge=%s|kind=%s|from=%s|to=%s|from_target_row_digest=%s|to_target_row_digest=%s|artifact_path=%s|artifact_bytes_digest=%s", prefix, ed.EdgeID, ed.Kind, ed.FromClaimID, ed.ToClaimID, fromRowDigest, toRowDigest, artifactPath, artifactDigest)
}

// canonicalTargetOccurrence independently identifies one target activity via
// syntax.ParseFile and the lowered graph. Its AST span is used only for the
// raw row digest, never as an execution or identity predicate.
func canonicalTargetOccurrence(artifact []byte, artifactPath string, sourceClaim claim) (targetOccurrence, []byte, error) {
	file, diagnostics := syntax.ParseFile(artifactPath, string(artifact))
	if file == nil || diagnostics.HasErrors() {
		return targetOccurrence{}, nil, fmt.Errorf("TARGET_SYNTAX_OR_LOWER_INVALID: %s", diagnostics.Error())
	}
	var activities []*syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		activity, ok := declaration.(*syntax.ActivityDecl)
		if ok && activity.Name == sourceClaim.ActivityName {
			activities = append(activities, activity)
		}
	}
	if len(activities) != 1 {
		return targetOccurrence{}, nil, fmt.Errorf("TARGET_ACTIVITY_OCCURRENCE_CARDINALITY: activity=%s count=%d", sourceClaim.ActivityName, len(activities))
	}
	target, err := reconstruct(artifact, artifactPath)
	if err != nil {
		return targetOccurrence{}, nil, fmt.Errorf("TARGET_SYNTAX_OR_LOWER_INVALID: %w", err)
	}
	var targetClaim *claim
	for i := range target.Graph.Nodes {
		if target.Graph.Nodes[i].ActivityName == sourceClaim.ActivityName {
			if targetClaim != nil {
				return targetOccurrence{}, nil, fmt.Errorf("TARGET_ACTIVITY_OCCURRENCE_CARDINALITY: activity=%s semantic_count>1", sourceClaim.ActivityName)
			}
			targetClaim = &target.Graph.Nodes[i]
		}
	}
	if targetClaim == nil || targetClaim.ClaimID != sourceClaim.ClaimID || targetClaim.PropositionDigest != sourceClaim.PropositionDigest || !reflect.DeepEqual(targetClaim.Target, sourceClaim.Target) {
		return targetOccurrence{}, nil, fmt.Errorf("TARGET_OCCURRENCE_SEMANTIC_ADDRESS_MISMATCH: activity=%s", sourceClaim.ActivityName)
	}
	span := activities[0].Span
	if span.Start.Offset < 0 || span.End.Offset < span.Start.Offset || span.End.Offset > len(artifact) {
		return targetOccurrence{}, nil, fmt.Errorf("TARGET_ACTIVITY_OCCURRENCE_SPAN_INVALID: activity=%s", sourceClaim.ActivityName)
	}
	raw := append([]byte(nil), artifact[span.Start.Offset:span.End.Offset]...)
	occurrence := targetOccurrence{Address: "activity:" + targetClaim.ClaimID, ActivityName: targetClaim.ActivityName, ClaimID: targetClaim.ClaimID, PropositionDigest: targetClaim.PropositionDigest, Target: targetClaim.Target, ValueProgram: targetClaim.ValueProgram, RawSpanStart: span.Start.Offset, RawSpanEnd: span.End.Offset, RawRowDigest: digestBytes(raw), SemanticDigest: targetOccurrenceSemanticDigest(*targetClaim), ContextDigest: target.Graph.CanonicalIRDigest}
	return occurrence, raw, nil
}

func targetOccurrenceMaterial(value targetOccurrence) string {
	return fmt.Sprintf("address=%s|activity=%s|claim_id=%s|proposition_digest=%s|target_inputs=%s|target_output=%s|target_artifact=%s|value_program=%s|semantic_digest=%s", value.Address, value.ActivityName, value.ClaimID, value.PropositionDigest, strings.Join(value.Target.Inputs, ","), value.Target.Output, value.Target.Artifact, value.ValueProgram, value.SemanticDigest)
}

func targetOccurrenceSemanticDigest(c claim) string {
	payload := fmt.Sprintf("target-occurrence|activity=%s|claim_id=%s|proposition=%s|target_inputs=%s|target_output=%s|target_artifact=%s|value_program=%s", c.ActivityName, c.ClaimID, c.Proposition, strings.Join(c.Target.Inputs, ","), c.Target.Output, c.Target.Artifact, c.ValueProgram)
	return digestBytes([]byte(payload))
}

func observationProcedureDigest(procedure, claimID, propositionDigest, edgeID string, occurrence targetOccurrence) string {
	payload := fmt.Sprintf("procedure|procedure_id=%s|claim_id=%s|proposition_digest=%s|edge_id=%s|%s", procedure, claimID, propositionDigest, edgeID, targetOccurrenceMaterial(occurrence))
	return digestBytes([]byte(payload))
}

func rawProvenanceDigest(occurrence targetOccurrence, artifactDigest string) string {
	payload := fmt.Sprintf("raw-provenance|artifact_digest=%s|address=%s|raw_span_start=%d|raw_span_end=%d|raw_row_digest=%s", artifactDigest, occurrence.Address, occurrence.RawSpanStart, occurrence.RawSpanEnd, occurrence.RawRowDigest)
	return digestBytes([]byte(payload))
}

func structuralContradictionDigest(value structuralContradiction) string {
	value.Digest = ""
	return digestJSON(value)
}

func deriveStructuralInventory(g graph, contract validatorContract, artifact []byte, artifactPath string) ([]structuralContradiction, error) {
	result := []structuralContradiction{}
	for _, c := range g.Nodes {
		material, ok := contractClaim(contract, c.ActivityName)
		if !ok {
			return nil, fmt.Errorf("STRUCTURAL_INVENTORY_CONTRACT_CLAIM_MISSING: %s", c.ActivityName)
		}
		if c.ValueProgram == material.ExpectedValueProgram {
			continue
		}
		occurrence, _, err := canonicalTargetOccurrence(artifact, artifactPath, c)
		if err != nil {
			return nil, err
		}
		finding := structuralContradiction{ClaimID: c.ClaimID, PropositionDigest: c.PropositionDigest, ExpectedValue: material.ExpectedValueProgram, DeclaredValue: c.ValueProgram, ProcedureID: material.ProcedureID, Occurrence: occurrence}
		finding.Digest = structuralContradictionDigest(finding)
		result = append(result, finding)
	}
	return result, nil
}

func validateStructuralInventory(observed, expected []structuralContradiction) error {
	expectedByClaim := map[string]structuralContradiction{}
	for _, finding := range expected {
		expectedByClaim[finding.ClaimID] = finding
	}
	seen := map[string]bool{}
	for _, finding := range observed {
		if seen[finding.ClaimID] {
			return fmt.Errorf("STRUCTURAL_INVENTORY_DUPLICATE: claim=%s", finding.ClaimID)
		}
		seen[finding.ClaimID] = true
	}
	if len(observed) < len(expected) {
		return fmt.Errorf("STRUCTURAL_INVENTORY_MISSING: observed=%d expected=%d", len(observed), len(expected))
	}
	if len(observed) > len(expected) {
		return fmt.Errorf("STRUCTURAL_INVENTORY_ADDITIONAL: observed=%d expected=%d", len(observed), len(expected))
	}
	for _, finding := range observed {
		want, ok := expectedByClaim[finding.ClaimID]
		if !ok || !reflect.DeepEqual(finding, want) {
			return fmt.Errorf("STRUCTURAL_INVENTORY_REPLACEMENT: claim=%s", finding.ClaimID)
		}
	}
	return nil
}
func contractClaim(value validatorContract, activity string) (validatorClaim, bool) {
	for _, c := range value.Claims {
		if c.ActivityName == activity {
			return c, true
		}
	}
	return validatorClaim{}, false
}
func claimMaterial(prefix string, c claim, t target, artifactPath, artifactDigest, program string) string {
	return fmt.Sprintf("%s|activity=%s|target_inputs=%s|target_output=%s|target_artifact=%s|target_path=%s|artifact_bytes_digest=%s|value_program=%s", prefix, c.ActivityName, strings.Join(t.Inputs, ","), t.Output, t.Artifact, artifactPath, artifactDigest, program)
}
func claimMatchesExpected(c claim, expected validatorClaim, contract validatorContract, artifactPath, artifactDigest string) bool {
	return reflect.DeepEqual(c.Target, expected.ExpectedTarget) && artifactPath == contract.ExpectedArtifactPath && artifactDigest == contract.ExpectedArtifactDigest && c.ValueProgram == expected.ExpectedValueProgram
}
func claimMatchesAlternate(c claim, expected validatorClaim, contract validatorContract, artifactPath, artifactDigest string) bool {
	return expected.AlternateValueProgram != "" && reflect.DeepEqual(c.Target, expected.ExpectedTarget) && artifactPath == contract.ExpectedArtifactPath && artifactDigest == contract.ExpectedArtifactDigest && c.ValueProgram == expected.AlternateValueProgram
}
func edgeMaterial(prefix string, ed edge, g graph, contract validatorContract, artifactDigest, artifactPath string, observed bool) string {
	from := indexOf(ed.FromClaimID, g)
	to := indexOf(ed.ToClaimID, g)
	fromContract, _ := contractClaim(contract, g.Nodes[from].ActivityName)
	toContract, _ := contractClaim(contract, g.Nodes[to].ActivityName)
	fromProgram, toProgram := fromContract.ExpectedValueProgram, toContract.ExpectedValueProgram
	if observed {
		fromProgram, toProgram = g.Nodes[from].ValueProgram, g.Nodes[to].ValueProgram
	}
	return fmt.Sprintf("%s|edge=%s|kind=%s|from=%s|to=%s|from_value_program=%s|to_value_program=%s|artifact_path=%s|artifact_bytes_digest=%s", prefix, ed.EdgeID, ed.Kind, ed.FromClaimID, ed.ToClaimID, fromProgram, toProgram, artifactPath, artifactDigest)
}
func validateFailureReceipt(value failureReceipt, sourcePath string, source []byte, artifactPath string, artifact []byte, g graph) error {
	if value.Schema != failureReceiptSchema || value.Provider == "" || value.SourcePath != sourcePath || value.SourceDigest != digestBytes(source) || value.ArtifactPath != artifactPath || value.ArtifactBytesDigest != digestBytes(artifact) || value.EdgeKind != failureEntailment || value.ObservedExitCode != 1 || value.Result != "NONZERO_EXIT" || value.Procedure != failureProcedure || value.Executable != "CI_EDGE_SPECIFIC_FAILURE_COMPARATOR" || value.ExecutableDigest == "" || digestBytes(value.ExecutableRaw) != value.ExecutableDigest || len(value.Argv) != 5 || value.Argv[0] != "-comparator" || value.Argv[1] != "-input" || value.Argv[3] != "-edge-id" || value.Argv[4] != value.EdgeID || value.StdoutDigest != digestBytes(value.Stdout) || value.StderrDigest != digestBytes(value.Stderr) || value.ProcedureDigest != failureProcedureDigest(value) || value.Coordinate.Stage == "" || value.Digest == "" {
		return fmt.Errorf("failure receipt is not an observed non-zero process result")
	}
	if !bytes.Contains(value.ExecutableRaw, []byte("FAILURE_ANTECEDENT")) || !bytes.Contains(value.ExecutableRaw, []byte("EDGE_SPECIFIC")) || !bytes.Contains(value.Stdout, []byte("FAILURE_ANTECEDENT_OBSERVED")) || !bytes.Contains(value.Stdout, []byte("EDGE_SPECIFIC")) {
		return fmt.Errorf("failure receipt executable is not the fixed edge comparator")
	}
	i := indexOfEdge(value.EdgeID, g)
	if i < 0 {
		return fmt.Errorf("failure receipt references an unknown edge")
	}
	ed := g.Edges[i]
	to := indexOf(ed.ToClaimID, g)
	if to < 0 || ed.Kind != failureEntailment || value.FromClaimID != ed.FromClaimID || value.ToClaimID != ed.ToClaimID || !reflect.DeepEqual(value.Target, g.Nodes[to].Target) || len(value.InputTargets) != 2 {
		return fmt.Errorf("failure receipt edge binding is invalid")
	}
	from := indexOf(ed.FromClaimID, g)
	if from < 0 {
		return fmt.Errorf("failure receipt edge binding is invalid")
	}
	fromOccurrence, fromRow, fromErr := canonicalTargetOccurrence(artifact, artifactPath, g.Nodes[from])
	toOccurrence, toRow, toErr := canonicalTargetOccurrence(artifact, artifactPath, g.Nodes[to])
	if fromErr != nil || toErr != nil || !failureInputMatches(value.InputTargets[0], g.Nodes[from], artifactPath, digestBytes(artifact), fromOccurrence) || !failureInputMatches(value.InputTargets[1], g.Nodes[to], artifactPath, digestBytes(artifact), toOccurrence) {
		return fmt.Errorf("failure receipt input targets are not bound to edge claims")
	}
	if !bytes.Contains(fromRow, []byte(g.Nodes[from].ValueProgram)) || !bytes.Contains(toRow, []byte(g.Nodes[to].ValueProgram)) {
		return fmt.Errorf("failure receipt process did not consume edge target outputs")
	}
	if failureDigest(value) != value.Digest {
		return fmt.Errorf("failure receipt digest is invalid")
	}
	return nil
}

func failureInputMatches(input failureInput, c claim, artifactPath, artifactDigest string, occurrence targetOccurrence) bool {
	return input.ClaimID == c.ClaimID && input.PropositionDigest == c.PropositionDigest && reflect.DeepEqual(input.Target, c.Target) && reflect.DeepEqual(input.Occurrence, occurrence) && input.TargetOutputDigest == occurrence.RawRowDigest && input.ValueProgram == c.ValueProgram && input.ArtifactPath == artifactPath && input.ArtifactDigest == artifactDigest
}

func failureProcedureDigest(value failureReceipt) string {
	parts := []string{value.Procedure, value.ExecutableDigest}
	parts = append(parts, value.Argv...)
	for _, input := range value.InputTargets {
		parts = append(parts, input.ClaimID, input.PropositionDigest, input.ValueProgram, input.ArtifactPath, input.ArtifactDigest, input.TargetOutputDigest, rawProvenanceDigest(input.Occurrence, input.ArtifactDigest), targetOccurrenceMaterial(input.Occurrence))
	}
	return digestBytes([]byte(strings.Join(parts, "|")))
}

func claimObservationFor(c claim, observations []observationReceipt) (observationReceipt, bool) {
	for _, observation := range observations {
		if observation.Binding == "CLAIM" && observation.ClaimID == c.ClaimID && observation.PropositionDigest == c.PropositionDigest && reflect.DeepEqual(observation.Target, c.Target) {
			return observation, true
		}
	}
	return observationReceipt{}, false
}

func hasClaimOrEdgeObservation(observations []observationReceipt) bool {
	for _, observation := range observations {
		if observation.Binding == "CLAIM" || observation.Binding == "EDGE" {
			return true
		}
	}
	return false
}

func summaryPredicate(observations []observationReceipt) predicate {
	for _, observation := range observations {
		if observation.Binding == "EDGE" && observation.ObservedPredicate == explicitContradiction {
			return explicitContradiction
		}
	}
	for _, observation := range observations {
		if observation.Binding == "EDGE" && observation.ObservedPredicate == failureAntecedent {
			return failureAntecedent
		}
	}
	for _, observation := range observations {
		if observation.Binding == "CLAIM" && observation.ObservedPredicate == explicitContradiction {
			return explicitContradiction
		}
	}
	for _, observation := range observations {
		if observation.Binding == "CLAIM" && observation.ObservedPredicate == accepted {
			return accepted
		}
	}
	return unknown
}

func observationBundleValue(bundle observationBundle) string {
	return fmt.Sprintf("bundle:%s|source_digest:%s|artifact_path:%s|artifact_bytes_digest:%s|contract_digest:%s|observation_total:%d", bundle.Digest, bundle.SourceDigest, bundle.ArtifactPath, bundle.ArtifactBytesDigest, bundle.ContractDigest, len(bundle.Observations))
}

func absentClaimValue(c claim) string {
	return fmt.Sprintf("observation:ABSENT|claim_id:%s|proposition_digest:%s|target_artifact:%s|stage:OBSERVE|step:claim-observation|reason:CLAIM_OBSERVATION_MISSING", c.ClaimID, c.PropositionDigest, c.Target.Artifact)
}
func validatePrior(current reconstructed, value receipt) error {
	if value.Schema != "gooo.meta.claim-dependency-receipt/v3" || value.Scope != "CLAIM_STATE_PROPAGATION_ONLY" || value.Evidence.ObservedPredicate != unknown || value.Graph.Digest != current.Graph.Digest || len(value.Resolutions) != claimTotal || len(value.PriorClaimStates) != 0 || receiptDigest(value) != value.Digest {
		return fmt.Errorf("prior UNKNOWN ledger invalid")
	}
	replayed, err := replayReceipt(value)
	if err != nil {
		return fmt.Errorf("prior receipt replay failed: %w", err)
	}
	if !reflect.DeepEqual(replayed, value) {
		return fmt.Errorf("prior receipt is not a complete replay of raw source and evidence")
	}
	if err := validateChain(value.Transitions, value.TransitionHeadDigest); err != nil {
		return err
	}
	for i, r := range value.Resolutions {
		if r.State != "OPEN" || r.ClaimID != current.Graph.Nodes[i].ClaimID {
			return fmt.Errorf("prior state %d not OPEN", i+1)
		}
	}
	return nil
}

func replayReceipt(value receipt) (receipt, error) {
	source, err := os.ReadFile(value.Subject.SourcePath)
	if err != nil {
		return receipt{}, fmt.Errorf("prior source cannot be re-observed: %w", err)
	}
	parsed, err := reconstruct(source, value.Subject.SourcePath)
	if err != nil {
		return receipt{}, err
	}
	artifact, err := os.ReadFile(value.Evidence.ArtifactPath)
	if err != nil {
		return receipt{}, fmt.Errorf("prior artifact cannot be re-observed: %w", err)
	}
	if err := validateEvidence(value.Evidence); err != nil {
		return receipt{}, err
	}
	if digestBytes(artifact) != value.Evidence.ArtifactBytesDigest {
		return receipt{}, fmt.Errorf("prior artifact bytes digest mismatch")
	}
	if err := validateEvidenceClaims(value.Evidence, parsed.Graph, artifact); err != nil {
		return receipt{}, err
	}
	states, outcomes := classify(parsed.Graph, value.Evidence)
	sourceDigest := digestBytes(source)
	provenance := fmt.Sprintf("source-semantic:%s|claim-evidence:%s|producer:%s|consumer:%s", parsed.Graph.CanonicalIRDigest, semanticEvidenceDigest(value.Evidence), producerID, consumerID)
	transitions := transitionsFor(parsed.Graph, outcomes, provenance, nil)
	resolutions := buildResolutions(parsed.Graph, states, outcomes, provenance)
	metrics := deriveMetrics(parsed.Graph, states, resolutions, outcomes, value.Evidence, false)
	decision := decisionFor(states, value.Evidence, false)
	authority := authorityResolution(value.Evidence)
	result := receipt{Schema: "gooo.meta.claim-dependency-receipt/v3", Scope: "CLAIM_STATE_PROPAGATION_ONLY", Subject: subject{SourcePath: value.Subject.SourcePath, SourceDigest: sourceDigest, SemanticDigest: parsed.Graph.CanonicalIRDigest, Producer: producerID, Consumer: consumerID, MetaOperation: operation, ProofChoice: proof, ReadOnly: authority == "NET_REPOSITORY_STATE_UNCHANGED", RepositoryWrites: value.Evidence.Snapshot.RepositoryWrites, AuthorityResolution: authority, AuthorityCoordinate: value.Evidence.Capability.Coordinate}, Evidence: value.Evidence, Graph: parsed.Graph, TruthTable: truthTable(), AuthorityCases: authorityCases(), EvidenceDigest: value.Evidence.Digest, Transitions: transitions, TransitionHeadDigest: transitions[len(transitions)-1].TransitionDigest, Resolutions: resolutions, Metrics: metrics, Decision: decision}
	result.Digest = receiptDigest(result)
	return result, nil
}

type local struct {
	predicate predicate
	digest    string
	available bool
}

func classify(g graph, e evidenceReceipt) ([]string, []transition) {
	states := make([]string, len(g.Nodes))
	outcomes := make([]transition, len(g.Nodes))
	locals := make([]local, len(g.Nodes))
	for i, c := range g.Nodes {
		for _, ec := range e.Claims {
			if ec.ClaimID == c.ClaimID && ec.PropositionDigest == c.PropositionDigest && ec.Status == "CURRENT_EVIDENCE" {
				locals[i] = local{ec.ObservedPredicate, ec.Digest, true}
				break
			}
		}
		state, event, reason := "OPEN", "DEPENDENCY_BLOCKED", "UPSTREAM_UNKNOWN_OR_NON_REFUTING"
		incoming := incomingEdges(i, g)
		refuting := []string{}
		hasRequires, allRequires := false, true
		for _, ed := range incoming {
			from := indexOf(ed.FromClaimID, g)
			if from < 0 {
				continue
			}
			relation := edgeRelation(ed.Kind, states[from], locals[i].predicate, true, edgeObservationActive(ed, e.Observations, explicitContradiction, g), edgeObservationActive(ed, e.Observations, failureAntecedent, g))
			if relation == relationRefuted {
				refuting = append(refuting, ed.EdgeID)
			}
			if ed.Kind == requires {
				hasRequires = true
				if relation != relationDischarged {
					allRequires = false
				}
			}
		}
		if locals[i].available && locals[i].predicate == explicitContradiction {
			state, event, reason = "REFUTED", "EXPLICIT_CONTRADICTION", "CURRENT_EVIDENCE_EXPLICIT_CONTRADICTION"
		} else if len(refuting) > 0 {
			state, event, reason = "REFUTED", "DEPENDENCY_REFUTED", "EXPLICIT_TYPED_REFUTING_EDGE"
		} else if locals[i].available && locals[i].predicate == accepted {
			if !hasRequires || allRequires {
				state, event, reason = "DISCHARGED", "EVIDENCE_ACCEPTED", "LOCAL_CLAIM_EVIDENCE_PREDICATE"
				if hasRequires {
					event, reason = "DEPENDENCY_DISCHARGED", "ALL_REQUIRES_UPSTREAM_AND_LOCAL_EVIDENCE"
				}
			}
		}
		states[i] = state
		outcomes[i] = transition{0, g.Nodes[i].ClaimID, event, "OPEN", state, coordinate{stage(i), g.Nodes[i].ActivityName, reason}, locals[i].digest, transitionEdges(i, g, states, state, refuting, locals[i].predicate, e.Observations), nil, "pending", "", ""}
	}
	return states, outcomes
}

func edgeObservationActive(ed edge, observations []observationReceipt, wanted predicate, g graph) bool {
	to := indexOf(ed.ToClaimID, g)
	if to < 0 {
		return false
	}
	for _, observation := range observations {
		if observation.Binding == "EDGE" && observation.EdgeID == ed.EdgeID && observation.FromClaimID == ed.FromClaimID && observation.ToClaimID == ed.ToClaimID && observation.EdgeKind == ed.Kind && observation.ClaimID == ed.ToClaimID && observation.PropositionDigest == g.Nodes[to].PropositionDigest && reflect.DeepEqual(observation.Target, g.Nodes[to].Target) && observation.ObservedPredicate == wanted {
			return true
		}
	}
	return false
}
func stage(i int) string {
	if i == 0 {
		return "OBSERVE"
	}
	return "PROPAGATE"
}
func incomingEdges(i int, g graph) []edge {
	result := []edge{}
	for _, e := range g.Edges {
		if e.ToClaimID == g.Nodes[i].ClaimID {
			result = append(result, e)
		}
	}
	return result
}
func transitionEdges(i int, g graph, states []string, state string, refuting []string, local predicate, observations []observationReceipt) []string {
	if len(refuting) > 0 {
		return refuting
	}
	if state == "DISCHARGED" {
		var result []string
		for _, e := range incomingEdges(i, g) {
			from := indexOf(e.FromClaimID, g)
			if e.Kind == requires && from >= 0 && edgeRelation(e.Kind, states[from], local, true, edgeObservationActive(e, observations, explicitContradiction, g), edgeObservationActive(e, observations, failureAntecedent, g)) == relationDischarged {
				result = append(result, e.EdgeID)
			}
		}
		return result
	}
	if state != "OPEN" {
		return nil
	}
	var result []string
	for _, e := range incomingEdges(i, g) {
		from := indexOf(e.FromClaimID, g)
		if from >= 0 && (e.Kind == supports || e.Kind == requires) && edgeRelation(e.Kind, states[from], local, true, edgeObservationActive(e, observations, explicitContradiction, g), edgeObservationActive(e, observations, failureAntecedent, g)) == relationOpen && (states[from] == "OPEN" || states[from] == "REFUTED") {
			result = append(result, e.EdgeID)
		}
	}
	return result
}
func transitionsFor(g graph, outcomes []transition, provenance string, prior *receipt) []transition {
	result := []transition{}
	previous := ""
	if prior != nil {
		result = append(result, prior.Transitions...)
		previous = prior.TransitionHeadDigest
	}
	if prior == nil {
		for _, c := range g.Nodes {
			value := transition{len(result) + 1, c.ClaimID, "CLAIM_REGISTERED", "UNRECORDED", "OPEN", coordinate{"DECLARE", c.ActivityName, "CLAIM_REGISTERED"}, "", nil, nil, provenance, previous, ""}
			value.TransitionDigest = transitionDigest(value)
			result = append(result, value)
			previous = value.TransitionDigest
		}
	}
	for _, value := range outcomes {
		value.Sequence = len(result) + 1
		value.Provenance = provenance
		value.PreviousTransitionDigest = previous
		value.UpstreamTransitionDigests = upstreamDigests(value.UpstreamEdgeIDs, g, result)
		value.TransitionDigest = transitionDigest(value)
		result = append(result, value)
		previous = value.TransitionDigest
	}
	return result
}
func upstreamDigests(ids []string, g graph, transitions []transition) []string {
	var result []string
	for _, id := range ids {
		for _, e := range g.Edges {
			if e.EdgeID != id {
				continue
			}
			for i := len(transitions) - 1; i >= 0; i-- {
				if transitions[i].ClaimID == e.FromClaimID && transitions[i].Event != "CLAIM_REGISTERED" {
					result = append(result, transitions[i].TransitionDigest)
					break
				}
			}
		}
	}
	return result
}
func buildResolutions(g graph, states []string, outcomes []transition, provenance string) []resolution {
	result := make([]resolution, len(g.Nodes))
	for i, c := range g.Nodes {
		path, ids, kinds := shortestPath(i, g, states[i])
		digests := []string{}
		for _, n := range path {
			for _, o := range outcomes {
				if o.ClaimID == g.Nodes[n].ClaimID {
					digests = append(digests, o.TransitionDigest)
				}
			}
		}
		causePath := idsForPath(path, g)
		responsibility, owner := failureAttribution(i, states[i], causePath, g)
		value := resolution{ClaimID: c.ClaimID, Axis: c.Axis, PropositionDigest: c.PropositionDigest, State: states[i], Kind: resolutionKind(i, states[i], outcomes[i]), ObservedEvent: outcomes[i].Event, Coordinate: outcomes[i].Coordinate, EvidenceDigest: outcomes[i].EvidenceDigest, Provenance: provenance, FailureResponsibility: responsibility, FailureOwnerClaimID: owner, CausePath: causePath, CauseEdgeIDs: ids, CauseEdgeKinds: kinds, CauseTransitionDigests: digests, CauseCoordinate: &outcomes[i].Coordinate}
		if states[i] == "OPEN" {
			value.MissingEvidenceIDs = []string{"evidence:" + c.ClaimID}
			value.BlockedByClaimIDs, value.BlockedByEdgeIDs = blockedFrontier(i, g, states)
		}
		result[i] = value
	}
	return result
}
func failureAttribution(index int, state string, path []string, g graph) (string, string) {
	if state == "DISCHARGED" {
		return "N/A", ""
	}
	if index == 0 {
		return "DIRECT_CLAIM", g.Nodes[index].ClaimID
	}
	if len(path) > 1 {
		return "UPSTREAM_CLAIM", path[0]
	}
	for _, e := range incomingEdges(index, g) {
		from := indexOf(e.FromClaimID, g)
		if from >= 0 && from != index {
			return "UPSTREAM_CLAIM", g.Nodes[from].ClaimID
		}
	}
	return "DIRECT_CLAIM", g.Nodes[index].ClaimID
}
func resolutionKind(i int, state string, outcome transition) string {
	if i == 0 {
		if state == "REFUTED" {
			return "DIRECT_REFUTED"
		}
		if state == "DISCHARGED" {
			return "DIRECT_DISCHARGED"
		}
		return "DIRECT_UNKNOWN"
	}
	if state == "REFUTED" {
		return "DEPENDENCY_REFUTED"
	}
	if state == "DISCHARGED" {
		if len(outcome.UpstreamEdgeIDs) == 0 {
			return "DIRECT_DISCHARGED"
		}
		return "DEPENDENCY_DISCHARGED"
	}
	return "DEPENDENCY_BLOCKED"
}
func shortestPath(index int, g graph, state string) ([]int, []string, []edgeKind) {
	if index == 0 {
		return []int{0}, nil, nil
	}
	allowed := map[edgeKind]bool{supports: state == "OPEN", requires: state == "OPEN" || state == "DISCHARGED", contradicts: state == "REFUTED", failureEntailment: state == "REFUTED"}
	best := []int(nil)
	bestEdges := []edge(nil)
	var walk func(int, []int, []edge)
	walk = func(current int, path []int, edges []edge) {
		if current == index {
			if best == nil || len(path) < len(best) || (len(path) == len(best) && pathKey(path, g) < pathKey(best, g)) {
				best = append([]int(nil), path...)
				bestEdges = append([]edge(nil), edges...)
			}
			return
		}
		for _, e := range g.Edges {
			if e.FromClaimID != g.Nodes[current].ClaimID || !allowed[e.Kind] {
				continue
			}
			next := indexOf(e.ToClaimID, g)
			seen := false
			for _, n := range path {
				if n == next {
					seen = true
				}
			}
			if !seen {
				walk(next, append(path, next), append(edges, e))
			}
		}
	}
	walk(0, []int{0}, nil)
	if best == nil {
		return []int{index}, nil, nil
	}
	ids, kinds := []string{}, []edgeKind{}
	for _, e := range bestEdges {
		ids, kinds = append(ids, e.EdgeID), append(kinds, e.Kind)
	}
	return best, ids, kinds
}
func pathKey(path []int, g graph) string {
	result := make([]string, len(path))
	for i, n := range path {
		result[i] = g.Nodes[n].ClaimID
	}
	return strings.Join(result, "\x00")
}
func idsForPath(path []int, g graph) []string {
	result := make([]string, len(path))
	for i, n := range path {
		result[i] = g.Nodes[n].ClaimID
	}
	return result
}
func indexOf(id string, g graph) int {
	for i, c := range g.Nodes {
		if c.ClaimID == id {
			return i
		}
	}
	return -1
}
func indexOfEdge(id string, g graph) int {
	for i, e := range g.Edges {
		if e.EdgeID == id {
			return i
		}
	}
	return -1
}
func blockedFrontier(i int, g graph, states []string) ([]string, []string) {
	var claims, edges []string
	for _, e := range incomingEdges(i, g) {
		from := indexOf(e.FromClaimID, g)
		if from >= 0 && (e.Kind == supports || e.Kind == requires) && (states[from] == "OPEN" || states[from] == "REFUTED") {
			claims, edges = append(claims, e.FromClaimID), append(edges, e.EdgeID)
		}
	}
	return claims, edges
}

func deriveMetrics(g graph, states []string, resolutions []resolution, outcomes []transition, e evidenceReceipt, recovered bool) metrics {
	structuralDenominator := e.StructuralInventoryTotal
	denominatorCoordinate := coordinate{Stage: "OBSERVE", Step: "structural-inventory", Reason: "EXTERNAL_STRUCTURAL_MANIFEST_INVENTORY"}
	if structuralDenominator == 0 {
		denominatorCoordinate.Reason = "NO_CURRENT_TARGET_OBSERVATION_EXPECTED_INVENTORY_ZERO"
	}
	result := metrics{FixedClaimTotal: claimTotal, DistinctPropositionTotal: distinct(g), FixedEdgeTotal: edgeTotal, EligibleEdgeTotal: len(g.Edges), ClassifiedClaimTotal: len(states), ClassificationBasisPoints: 10000, TransitionTotal: initialTransitions, TruthTableCaseTotal: len(truthTable()), AuthorityCaseTotal: len(authorityCases()), StructuralContradictionNumerator: len(e.StructuralContradictions), StructuralContradictionDenominator: structuralDenominator, SemanticOccurrenceNumerator: e.SemanticOccurrenceNumerator, SemanticOccurrenceDenominator: e.SemanticOccurrenceDenominator, RawProvenanceBindingNumerator: e.RawProvenanceBindingNumerator, RawProvenanceBindingDenominator: e.RawProvenanceBindingDenominator, StructuralContradictionDenominatorCoordinate: denominatorCoordinate}
	if recovered {
		result.TransitionTotal += claimTotal
	}
	observed, shortestUnion := map[string]bool{}, map[string]bool{}
	for _, ec := range e.Claims {
		if ec.Status == "CURRENT_EVIDENCE" {
			result.CurrentEvidenceTotal++
		}
		if ec.Status == "HISTORICAL_FIXTURE" {
			result.HistoricalEvidenceTotal++
		}
		if ec.ObservedPredicate == unknown {
			result.UnknownEvidenceTotal++
		}
	}
	for _, state := range states {
		switch state {
		case "OPEN":
			result.OpenClaimTotal++
		case "DISCHARGED":
			result.DischargedClaimTotal++
		case "REFUTED":
			result.RefutedClaimTotal++
		}
	}
	for i, r := range resolutions {
		switch r.Kind {
		case "DIRECT_UNKNOWN":
			result.DirectUnknownClaimTotal++
		case "DEPENDENCY_BLOCKED":
			result.DependencyBlockedClaimTotal++
		case "DIRECT_REFUTED":
			result.DirectRefutedClaimTotal++
		case "DEPENDENCY_REFUTED":
			result.DependencyRefutedClaimTotal++
		case "DIRECT_DISCHARGED":
			result.DirectDischargedClaimTotal++
		case "DEPENDENCY_DISCHARGED":
			result.DependencyDischargedTotal++
		}
		for _, id := range outcomes[i].UpstreamEdgeIDs {
			observed[id] = true
		}
		for _, id := range r.CauseEdgeIDs {
			shortestUnion[id] = true
		}
		if len(r.CauseEdgeIDs) > result.MaximumCausePathDepth {
			result.MaximumCausePathDepth = len(r.CauseEdgeIDs)
		}
	}
	result.ObservedCausalEdgeTotal, result.ShortestPathEdgeUnionTotal = len(observed), len(shortestUnion)
	for _, kind := range []edgeKind{supports, requires, contradicts, failureEntailment} {
		m := edgeMetric{Kind: kind}
		for _, e := range g.Edges {
			if e.Kind != kind {
				continue
			}
			m.Eligible++
			if observed[e.EdgeID] {
				m.ObservedCausal++
			}
			for _, o := range outcomes {
				if contains(o.UpstreamEdgeIDs, e.EdgeID) {
					if o.After == "OPEN" {
						m.Blocking++
					}
					if o.After == "REFUTED" {
						m.Refuting++
					}
					if recovered && o.After == "DISCHARGED" {
						m.Discharge++
					}
				}
			}
		}
		result.ObservedBlockingEdgeTotal += m.Blocking
		result.ObservedRefutingEdgeTotal += m.Refuting
		result.ObservedRecoveryEdgeTotal += m.Discharge
		result.EdgeMetrics = append(result.EdgeMetrics, m)
	}
	return result
}
func distinct(g graph) int {
	seen := map[string]bool{}
	for _, c := range g.Nodes {
		seen[c.PropositionDigest] = true
	}
	return len(seen)
}

func semanticEvidenceDigest(evidence evidenceReceipt) string {
	claims := make([]semanticEvidenceClaim, len(evidence.Claims))
	for i, claim := range evidence.Claims {
		claims[i] = semanticEvidenceClaim{ClaimID: claim.ClaimID, PropositionDigest: claim.PropositionDigest, ObservedPredicate: claim.ObservedPredicate, ObservedValue: claim.ObservedValue, Status: claim.Status, Coordinate: claim.Coordinate}
	}
	observations := make([]semanticEvidenceObservation, 0, len(evidence.Observations))
	for _, observation := range evidence.Observations {
		observations = append(observations, semanticObservation(observation))
	}
	structural := make([]semanticStructuralContradiction, len(evidence.StructuralContradictions))
	for i, finding := range evidence.StructuralContradictions {
		structural[i] = semanticStructuralContradiction{ClaimID: finding.ClaimID, PropositionDigest: finding.PropositionDigest, ExpectedValue: finding.ExpectedValue, DeclaredValue: finding.DeclaredValue, ProcedureID: finding.ProcedureID, Occurrence: semanticOccurrenceProjection(finding.Occurrence)}
	}
	manifest := semanticStructuralManifest{}
	if evidence.StructuralManifestRaw != nil {
		var value structuralInventoryManifest
		if strictJSON(evidence.StructuralManifestRaw, &value) == nil {
			manifest = semanticStructuralManifest{EligibleClaimIDs: append([]string(nil), value.EligibleClaimIDs...), ExpectedContradictionClaimIDs: append([]string(nil), value.ExpectedContradictionClaimIDs...)}
		}
	}
	projection := semanticEvidenceProjection{Status: evidence.Status, ObservedPredicate: evidence.ObservedPredicate, Claims: claims, Observations: observations, StructuralContradictions: structural, StructuralManifest: manifest, StructuralInventoryTotal: evidence.StructuralInventoryTotal, SemanticOccurrenceNumerator: evidence.SemanticOccurrenceNumerator, SemanticOccurrenceDenominator: evidence.SemanticOccurrenceDenominator}
	return digestJSON(projection)
}

type semanticEvidenceProjection struct {
	Status                        string                            `json:"status"`
	ObservedPredicate             predicate                         `json:"observed_predicate"`
	Claims                        []semanticEvidenceClaim           `json:"claims"`
	Observations                  []semanticEvidenceObservation     `json:"observations"`
	StructuralContradictions      []semanticStructuralContradiction `json:"structural_contradictions"`
	StructuralManifest            semanticStructuralManifest        `json:"structural_manifest"`
	StructuralInventoryTotal      int                               `json:"structural_inventory_total"`
	SemanticOccurrenceNumerator   int                               `json:"semantic_occurrence_numerator"`
	SemanticOccurrenceDenominator int                               `json:"semantic_occurrence_denominator"`
}

type semanticStructuralManifest struct {
	EligibleClaimIDs              []string `json:"eligible_claim_ids"`
	ExpectedContradictionClaimIDs []string `json:"expected_contradiction_claim_ids"`
}

type semanticEvidenceClaim struct {
	ClaimID           string     `json:"claim_id"`
	PropositionDigest string     `json:"proposition_digest"`
	ObservedPredicate predicate  `json:"observed_predicate"`
	ObservedValue     string     `json:"observed_value"`
	Status            string     `json:"status"`
	Coordinate        coordinate `json:"coordinate"`
}

type semanticEvidenceObservation struct {
	Binding           string             `json:"binding"`
	ClaimID           string             `json:"claim_id,omitempty"`
	PropositionDigest string             `json:"proposition_digest,omitempty"`
	EdgeID            string             `json:"edge_id,omitempty"`
	FromClaimID       string             `json:"from_claim_id,omitempty"`
	ToClaimID         string             `json:"to_claim_id,omitempty"`
	EdgeKind          edgeKind           `json:"edge_kind,omitempty"`
	Target            target             `json:"target"`
	Occurrence        semanticOccurrence `json:"occurrence"`
	ExpectedPredicate predicate          `json:"expected_predicate"`
	ObservedPredicate predicate          `json:"observed_predicate"`
	ComparisonResult  string             `json:"comparison_result"`
	SemanticExpected  string             `json:"semantic_expected"`
	SemanticObserved  string             `json:"semantic_observed"`
	Procedure         string             `json:"procedure"`
	ProcedureDigest   string             `json:"procedure_digest"`
	Coordinate        coordinate         `json:"coordinate"`
}

type semanticOccurrence struct {
	Address           string `json:"address"`
	ActivityName      string `json:"activity_name"`
	ClaimID           string `json:"claim_id"`
	PropositionDigest string `json:"proposition_digest"`
	Target            target `json:"target"`
	ValueProgram      string `json:"value_program"`
	SemanticDigest    string `json:"semantic_digest"`
	ContextDigest     string `json:"context_digest"`
}

type semanticStructuralContradiction struct {
	ClaimID           string             `json:"claim_id"`
	PropositionDigest string             `json:"proposition_digest"`
	ExpectedValue     string             `json:"expected_value"`
	DeclaredValue     string             `json:"declared_value"`
	ProcedureID       string             `json:"procedure_id"`
	Occurrence        semanticOccurrence `json:"occurrence"`
}

func semanticObservation(value observationReceipt) semanticEvidenceObservation {
	return semanticEvidenceObservation{Binding: value.Binding, ClaimID: value.ClaimID, PropositionDigest: value.PropositionDigest, EdgeID: value.EdgeID, FromClaimID: value.FromClaimID, ToClaimID: value.ToClaimID, EdgeKind: value.EdgeKind, Target: value.Target, Occurrence: semanticOccurrenceProjection(value.Occurrence), ExpectedPredicate: value.ExpectedPredicate, ObservedPredicate: value.ObservedPredicate, ComparisonResult: value.ComparisonResult, SemanticExpected: semanticObservationValue(value, true), SemanticObserved: semanticObservationValue(value, false), Procedure: value.Procedure, ProcedureDigest: value.ProcedureDigest, Coordinate: value.Coordinate}
}

func semanticObservationValue(value observationReceipt, expected bool) string {
	if value.Binding == "CLAIM" {
		if expected {
			return value.ExpectedValue
		}
		return value.ObservedValue
	}
	if value.Binding == "EDGE" {
		if value.EdgeKind == contradicts {
			return "CONTRADICTS_TARGET_VALUE_OPPOSITE"
		}
		if value.EdgeKind == failureEntailment {
			return "FAILURE_ANTECEDENT_PROCESS"
		}
	}
	return string(value.ObservedPredicate)
}

func semanticOccurrenceProjection(value targetOccurrence) semanticOccurrence {
	return semanticOccurrence{Address: value.Address, ActivityName: value.ActivityName, ClaimID: value.ClaimID, PropositionDigest: value.PropositionDigest, Target: value.Target, ValueProgram: value.ValueProgram, SemanticDigest: value.SemanticDigest, ContextDigest: value.ContextDigest}
}
func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
func decisionFor(states []string, e evidenceReceipt, recovered bool) decision {
	switch authorityResolution(e) {
	case "NET_REPOSITORY_STATE_CHANGED":
		return decision{"FAIL_CLOSED", "AUTHORITY_CHANGED", "AUTHORITY/REPOSITORY_SNAPSHOT/NET_REPOSITORY_STATE_CHANGED", false}
	case "TRANSIENT_WRITE_AUTHORITY_UNKNOWN":
		return decision{"FAIL_CLOSED", "AUTHORITY_UNKNOWN", "AUTHORITY/WORKFLOW_PERMISSIONS/TRANSIENT_WRITE_AUTHORITY_UNKNOWN", false}
	}
	if all(states, "DISCHARGED") {
		if recovered {
			return decision{"PASS", "CAUSAL_RECOVERY_DISCHARGED", "APPEND_ONLY_EVIDENCE_RECOVERY", false}
		}
		return decision{"PASS", "DIRECT_EVIDENCE_DISCHARGED", "CURRENT_EVIDENCE_PREDICATES_SATISFIED", false}
	}
	if any(states, "REFUTED") {
		if count(states, "REFUTED") == 1 {
			return decision{"FAIL_CLOSED", "DIRECT_REFUTATION", "ONLY_DIRECT_EXPLICIT_CONTRADICTION", false}
		}
		return decision{"FAIL_CLOSED", "CAUSAL_REFUTATION", "EXPLICIT_CONTRADICTION_OR_FAILURE_ENTAILMENT", false}
	}
	return decision{"FAIL_CLOSED", "UNRESOLVED_CLAIM", "UNKNOWN_REMAINS_OPEN", false}
}
func authorityResolution(e evidenceReceipt) string {
	if e.Snapshot.RepositoryWrites != 0 || e.Snapshot.BeforeDigest != e.Snapshot.AfterDigest {
		return "NET_REPOSITORY_STATE_CHANGED"
	}
	if e.Capability.Status != "CURRENT_EVIDENCE" || e.Capability.Provider == "" {
		return "TRANSIENT_WRITE_AUTHORITY_UNKNOWN"
	}
	return "NET_REPOSITORY_STATE_UNCHANGED"
}
func authorityCases() []authorityCase {
	result := []authorityCase{{"NET-SAME-CURRENT", "NET_SAME", "CURRENT_EVIDENCE", "NET_REPOSITORY_STATE_UNCHANGED", ""}, {"NET-CHANGED-CURRENT", "NET_CHANGED", "CURRENT_EVIDENCE", "NET_REPOSITORY_STATE_CHANGED", ""}, {"TRANSIENT-UNKNOWN", "TRANSIENT_UNKNOWN", "UNKNOWN", "TRANSIENT_WRITE_AUTHORITY_UNKNOWN", ""}}
	for i := range result {
		if result[i].NetworkState == "NET_CHANGED" {
			result[i].ObservedResolution = "NET_REPOSITORY_STATE_CHANGED"
		} else if result[i].NetworkState == "TRANSIENT_UNKNOWN" || result[i].CapabilityStatus != "CURRENT_EVIDENCE" {
			result[i].ObservedResolution = "TRANSIENT_WRITE_AUTHORITY_UNKNOWN"
		} else {
			result[i].ObservedResolution = "NET_REPOSITORY_STATE_UNCHANGED"
		}
	}
	return result
}
func validateAuthorityCases(values []authorityCase) error {
	if len(values) != 3 {
		return fmt.Errorf("authority cases have %d cases, want 3", len(values))
	}
	for _, value := range values {
		if value.ExpectedResolution == "" || value.ExpectedResolution != value.ObservedResolution {
			return fmt.Errorf("authority case %q did not execute its expected resolution", value.CaseID)
		}
	}
	return nil
}
func all(values []string, target string) bool {
	for _, v := range values {
		if v != target {
			return false
		}
	}
	return true
}
func any(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
func count(values []string, target string) int {
	total := 0
	for _, value := range values {
		if value == target {
			total++
		}
	}
	return total
}
func statesOf(values []resolution) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = v.State
	}
	return result
}
func validateChain(values []transition, head string) error {
	if len(values) == 0 || values[len(values)-1].TransitionDigest != head {
		return fmt.Errorf("transition head mismatch")
	}
	previous := ""
	for i, v := range values {
		if v.Sequence != i+1 || v.PreviousTransitionDigest != previous || transitionDigest(v) != v.TransitionDigest {
			return fmt.Errorf("transition %d chain mismatch", i+1)
		}
		previous = v.TransitionDigest
	}
	return nil
}
func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func digestJSON(v any) string {
	data, _ := json.Marshal(v)
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func digestBytes(v []byte) string {
	sum := sha256.Sum256(v)
	return "sha256:" + hex.EncodeToString(sum[:])
}
func prefixedDigest(v string) string {
	if strings.HasPrefix(v, "sha256:") {
		return v
	}
	return "sha256:" + v
}
func graphDigest(v graph) string                    { v.Digest = ""; return digestJSON(v) }
func receiptDigest(v receipt) string                { v.Digest = ""; return digestJSON(v) }
func evidenceDigest(v evidenceReceipt) string       { v.Digest = ""; return digestJSON(v) }
func observationDigest(v observationReceipt) string { v.Digest = ""; return digestJSON(v) }
func observationBundleDigest(v observationBundle) string {
	v.Digest = ""
	v.Profile = ""
	return digestJSON(v)
}
func failureDigest(v failureReceipt) string      { v.Digest = ""; return digestJSON(v) }
func evidenceClaimDigest(v evidenceClaim) string { v.Digest = ""; return digestJSON(v) }
func capabilityDigest(v capability) string       { v.Digest = ""; return digestJSON(v) }
func transitionDigest(v transition) string       { v.TransitionDigest = ""; return digestJSON(v) }
