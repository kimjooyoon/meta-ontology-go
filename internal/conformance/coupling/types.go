package coupling

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

const (
	SchemaV1       = "gooo/coupling/v1"
	CorpusSchemaV1 = "gooo/coupling-corpus/v1"
)

type Decision string

const (
	DecisionPass       Decision = "PASS"
	DecisionFailClosed Decision = "FAIL_CLOSED"
	DecisionUnknown    Decision = "UNKNOWN"
)

type Reason string

const (
	ReasonNone                   Reason = "none"
	ReasonRequiredInputMissing   Reason = "required-input-missing"
	ReasonInputAmbiguous         Reason = "ambiguous-input"
	ReasonRegistryBinding        Reason = "registry-binding-mismatch"
	ReasonSurfaceUnregistered    Reason = "surface-unregistered"
	ReasonChangedSurface         Reason = "changed-surface-mismatch"
	ReasonMissingReceipt         Reason = "missing-receipt"
	ReasonOrphanReceipt          Reason = "orphan-receipt"
	ReasonDuplicateReceipt       Reason = "duplicate-receipt"
	ReasonStaleReceipt           Reason = "stale-receipt"
	ReasonDigestMismatch         Reason = "digest-mismatch"
	ReasonProfileMismatch        Reason = "profile-mismatch"
	ReasonSourceUnbound          Reason = "source-unbound"
	ReasonDeltaWithoutSource     Reason = "delta-without-source"
	ReasonNoDeltaWithoutEquality Reason = "no-delta-without-equality"
	ReasonInvalidDelta           Reason = "invalid-delta"
	ReasonPathMissing            Reason = "path-missing"
	ReasonPathMalformed          Reason = "path-malformed"
	ReasonPathClosure            Reason = "path-closure-mismatch"
	ReasonCandidateAuthority     Reason = "candidate-not-authority"
	ReasonResourceUnbound        Reason = "resource-unbound"
)

type ChangeClaim string

const (
	ClaimDelta   ChangeClaim = "DELTA"
	ClaimNoDelta ChangeClaim = "NO_DELTA"
)

type ReceiptKind string

const (
	ReceiptSemanticDelta   ReceiptKind = "SEMANTIC_DELTA"
	ReceiptNoSemanticDelta ReceiptKind = "NO_SEMANTIC_DELTA"
)

type EvaluationConfig struct {
	ToolchainDigest string                `json:"toolchain_digest"`
	Profile         ProfileConfig         `json:"profile"`
	ResourceBinding ResourceBindingConfig `json:"resource_binding"`
}

type ProfileConfig struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

type ResourceBindingConfig struct {
	ProviderID     string `json:"provider_id"`
	ObserverID     string `json:"observer_id"`
	ProviderDigest string `json:"provider_digest"`
	ObserverDigest string `json:"observer_digest"`
	SnapshotDigest string `json:"snapshot_digest"`
	SourceDigest   string `json:"source_digest"`
}

// SourceManifest is an explicit completeness claim. ZeroChange is only
// authoritative when Complete is true and all snapshot bindings agree.
type SourceManifest struct {
	Complete             bool   `json:"complete"`
	ZeroChange           bool   `json:"zero_change"`
	BeforeSnapshotDigest string `json:"before_snapshot_digest"`
	AfterSnapshotDigest  string `json:"after_snapshot_digest"`
	ToolchainDigest      string `json:"toolchain_digest"`
	ProfileDigest        string `json:"profile_digest"`
	RegistryDigest       string `json:"registry_digest"`
}

// SemanticNode keeps labels available for the rename partition. Canonical
// semantic identity deliberately excludes Name and Aliases.
type SemanticNode struct {
	ID        string   `json:"id"`
	Kind      string   `json:"kind"`
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	Aliases   []string `json:"aliases,omitempty"`
}

type SemanticRelation struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

type SemanticIR struct {
	Nodes     []SemanticNode     `json:"nodes"`
	Relations []SemanticRelation `json:"relations"`
}

type CodeBinding struct {
	RegisteredSurfaceID string `json:"registered_surface_id"`
	CodeSymbolID        string `json:"code_symbol_id"`
	SemanticOwnerID     string `json:"semantic_owner_id"`
	SourceMapID         string `json:"source_map_id"`
	BindingDigest       string `json:"binding_digest"`
	PackageLabel        string `json:"package_label,omitempty"`
	FileLabel           string `json:"file_label,omitempty"`
	SourceSpan          string `json:"source_span,omitempty"`
}

type CodeChange struct {
	CodeSymbolID string `json:"code_symbol_id"`
	BeforeDigest string `json:"before_digest"`
	AfterDigest  string `json:"after_digest"`
}

// ExternalResourceReceipt is the only admissible source for resource
// observations. The oracle never derives CPU, peak memory, or work from
// fixture cardinality or elapsed wall time.
type ExternalResourceReceipt struct {
	ReceiptID      string `json:"receipt_id"`
	Metric         string `json:"metric"`
	Value          uint64 `json:"value"`
	Unit           string `json:"unit"`
	ProviderDigest string `json:"provider_digest"`
	ObserverDigest string `json:"observer_digest"`
	SnapshotDigest string `json:"snapshot_digest"`
	SourceDigest   string `json:"source_digest"`
	BindingDigest  string `json:"binding_digest"`
	Present        bool   `json:"present"`
	Independent    bool   `json:"independent"`
	State          string `json:"state"`
}

type ResourceObservation struct {
	CPUCoreNS       uint64 `json:"cpu_core_ns"`
	PeakMemoryBytes uint64 `json:"peak_memory_bytes"`
	WorkUnits       uint64 `json:"work_units"`
}

type CouplingReceipt struct {
	ReceiptID                   string      `json:"receipt_id"`
	SurfaceID                   string      `json:"surface_id"`
	SemanticOwnerID             string      `json:"semantic_owner_id"`
	CodeSymbolID                string      `json:"code_symbol_id"`
	SourceMapBindingDigest      string      `json:"source_map_binding_digest"`
	SnapshotDigest              string      `json:"snapshot_digest"`
	RegistryDigest              string      `json:"registry_digest"`
	ToolchainDigest             string      `json:"toolchain_digest"`
	ProfileDigest               string      `json:"profile_digest"`
	BeforeIRDigest              string      `json:"before_ir_digest"`
	AfterIRDigest               string      `json:"after_ir_digest"`
	AuthoritySourceBeforeDigest string      `json:"authority_source_before_digest"`
	AuthoritySourceAfterDigest  string      `json:"authority_source_after_digest"`
	ChangeClaim                 ChangeClaim `json:"change_claim"`
	ReceiptKind                 ReceiptKind `json:"receipt_kind"`
	SemanticDelta               string      `json:"semantic_delta,omitempty"`
	SemanticDeltaDigest         string      `json:"semantic_delta_digest,omitempty"`
	AuthoritativeSourceRef      string      `json:"authoritative_source_ref,omitempty"`
	OriginPathID                string      `json:"origin_path_id"`
	ClaimRecordID               string      `json:"claim_record_id"`
	EvidenceRefs                []string    `json:"evidence_refs"`
	State                       string      `json:"state"`
}

type Input struct {
	Schema                string                    `json:"schema"`
	FixtureID             string                    `json:"fixture_id"`
	RegistryDigest        string                    `json:"registry_digest"`
	Config                EvaluationConfig          `json:"config"`
	Manifest              SourceManifest            `json:"manifest"`
	ResourceRegistry      ResourceBindingConfig     `json:"resource_registry"`
	AuthoritySourceBefore string                    `json:"authority_source_before"`
	AuthoritySourceAfter  string                    `json:"authority_source_after"`
	SemanticBefore        SemanticIR                `json:"semantic_before"`
	SemanticAfter         SemanticIR                `json:"semantic_after"`
	Registry              []CodeBinding             `json:"registry"`
	Changes               []CodeChange              `json:"changes"`
	Receipts              []CouplingReceipt         `json:"receipts"`
	ResourceReceipts      []ExternalResourceReceipt `json:"resource_receipts"`
	Roots                 []string                  `json:"roots"`
	Path                  semantic.InferencePathV1  `json:"path"`
}

type ObservationCounts struct {
	RegistryBindings      uint64 `json:"registry_bindings"`
	ChangedCodeSurfaces   uint64 `json:"changed_code_surfaces"`
	ChangedRegistered     uint64 `json:"changed_registered_surfaces"`
	ReceiptRecords        uint64 `json:"receipt_records"`
	ValidReceipts         uint64 `json:"valid_receipts"`
	OrphanReceipts        uint64 `json:"orphan_receipts"`
	DuplicateReceipts     uint64 `json:"duplicate_receipts"`
	PathEdges             uint64 `json:"path_edges"`
	PathClaims            uint64 `json:"path_claims"`
	PathEvidence          uint64 `json:"path_evidence"`
	CandidateObservations uint64 `json:"candidate_observations"`
	AcceptedLifts         uint64 `json:"accepted_lifts"`
	AddedSemanticFacts    uint64 `json:"added_semantic_facts"`
	RemovedSemanticFacts  uint64 `json:"removed_semantic_facts"`
	ResourceReceipts      uint64 `json:"resource_receipts"`
}

type Output struct {
	Schema                string              `json:"schema"`
	FixtureID             string              `json:"fixture_id"`
	InputDigest           string              `json:"input_digest"`
	Decision              Decision            `json:"decision"`
	Reason                Reason              `json:"reason"`
	AcceptedSurfaces      []string            `json:"accepted_surfaces"`
	ChangedSurfaces       []string            `json:"changed_surfaces"`
	ReceiptSurfaces       []string            `json:"receipt_surfaces"`
	SemanticBeforeDigest  string              `json:"semantic_before_digest"`
	SemanticAfterDigest   string              `json:"semantic_after_digest"`
	SemanticDeltaDigest   string              `json:"semantic_delta_digest,omitempty"`
	PathClosureDigest     string              `json:"path_closure_digest,omitempty"`
	ObservationCounts     ObservationCounts   `json:"observation_counts"`
	Resources             ResourceObservation `json:"resources"`
	CanonicalOutputDigest string              `json:"canonical_output_digest"`
	ReplayDigest          string              `json:"replay_digest"`
}

// BaselineResult is a separately implemented, typed-config/full-suite
// comparison. It intentionally does not reuse oracle validation helpers.
type BaselineResult struct {
	Decision          Decision            `json:"decision"`
	Reason            Reason              `json:"reason"`
	LocalizedSurfaces []string            `json:"localized_surfaces"`
	FullSuite         bool                `json:"full_suite"`
	ObservationCounts ObservationCounts   `json:"observation_counts"`
	Resources         ResourceObservation `json:"resources"`
	WorkUnits         uint64              `json:"work_units"`
}

type Comparison struct {
	Oracle            Output         `json:"oracle"`
	Baseline          BaselineResult `json:"baseline"`
	OutcomeMatch      bool           `json:"outcome_match"`
	ReasonMatch       bool           `json:"reason_match"`
	LocalizationMatch bool           `json:"localization_match"`
	Finding           string         `json:"finding"`
}

type FixtureExpectation struct {
	Decision          Decision            `json:"decision"`
	Reason            Reason              `json:"reason"`
	AcceptedSurfaces  []string            `json:"accepted_surfaces"`
	ChangedSurfaces   []string            `json:"changed_surfaces"`
	ReceiptSurfaces   []string            `json:"receipt_surfaces"`
	ObservationCounts ObservationCounts   `json:"observation_counts"`
	Resources         ResourceObservation `json:"resources"`
}

type CorpusCase struct {
	Name     string             `json:"name"`
	Input    Input              `json:"input"`
	Expected FixtureExpectation `json:"expected"`
}

type Corpus struct {
	Schema          string       `json:"schema"`
	CanonicalDigest string       `json:"canonical_digest"`
	Cases           []CorpusCase `json:"cases"`
}
