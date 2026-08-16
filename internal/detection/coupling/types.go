package coupling

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

const (
	RegistrySchemaV1         = "gooo/code-semantic-coupling-registry/v1"
	ManifestSchemaV1         = "gooo/code-semantic-coupling-manifest/v1"
	ConfigSchemaV1           = "gooo/code-semantic-coupling-config/v1"
	InputSchemaV1            = "gooo/code-semantic-coupling-input/v1"
	ReceiptSchemaV1          = "gooo/code-semantic-coupling-receipt/v1"
	ResultSchemaV1           = "gooo/code-semantic-coupling-result/v1"
	BaselineSchemaV1         = "gooo/code-semantic-coupling-baseline/v1"
	ResourceSchemaV1         = "gooo/external-resource-receipt/v1"
	AuthorityContextSchemaV1 = "gooo/code-semantic-coupling-authority/v1"
)

type Status string

const (
	StatusPass       Status = "PASS"
	StatusFailClosed Status = "FAIL_CLOSED"
	StatusUnknown    Status = "UNKNOWN"
)

type ReasonCode string

const (
	ReasonMalformedBinding        ReasonCode = "MALFORMED_BINDING"
	ReasonRequiredInputMissing    ReasonCode = "REQUIRED_INPUT_MISSING"
	ReasonDuplicateSurface        ReasonCode = "DUPLICATE_SURFACE"
	ReasonSurfaceUnregistered     ReasonCode = "SURFACE_UNREGISTERED"
	ReasonDuplicateReceipt        ReasonCode = "DUPLICATE_RECEIPT"
	ReasonOrphanReceipt           ReasonCode = "ORPHAN_RECEIPT"
	ReasonStaleInput              ReasonCode = "STALE_INPUT"
	ReasonDigestMismatch          ReasonCode = "DIGEST_MISMATCH"
	ReasonSourceMapMismatch       ReasonCode = "SOURCE_MAP_MISMATCH"
	ReasonContradictoryReceipt    ReasonCode = "CONTRADICTORY_RECEIPT"
	ReasonDeltaWithoutSource      ReasonCode = "DELTA_WITHOUT_SOURCE"
	ReasonNoDeltaWithoutEquality  ReasonCode = "NO_DELTA_WITHOUT_EQUALITY"
	ReasonCandidateOnlyPath       ReasonCode = "CANDIDATE_ONLY_PATH"
	ReasonInferencePathMalformed  ReasonCode = "MALFORMED_INFERENCE_PATH"
	ReasonMissingAuthorityPath    ReasonCode = "AUTHORITY_PATH_MISSING"
	ReasonMissingVerification     ReasonCode = "INDEPENDENT_VERIFICATION_MISSING"
	ReasonExternalReceiptMissing  ReasonCode = "EXTERNAL_RECEIPT_MISSING"
	ReasonAuthorityInputSelfBound ReasonCode = "COUPLING_AUTHORITY_INPUT_SELF_BOUND"
)

type Reason struct {
	Code   ReasonCode `json:"code"`
	Detail string     `json:"detail"`
}

type SourceMapBinding struct {
	SourceMapID   semantic.ID `json:"source_map_id"`
	BindingDigest string      `json:"binding_digest"`
	PackageLabel  string      `json:"package_label,omitempty"`
	FileLabel     string      `json:"file_label,omitempty"`
	SourceSpan    string      `json:"source_span,omitempty"`
}

type Surface struct {
	SurfaceID         semantic.ID      `json:"surface_id"`
	CodeSymbolID      semantic.ID      `json:"code_symbol_id"`
	SemanticOwnerID   semantic.ID      `json:"semantic_owner_id"`
	Binding           SourceMapBinding `json:"binding"`
	PresentationLabel string           `json:"presentation_label,omitempty"`
}

type Registry struct {
	Schema   string    `json:"schema"`
	Digest   string    `json:"digest"`
	Surfaces []Surface `json:"surfaces"`
}

type ManifestEntry struct {
	SurfaceID           semantic.ID `json:"surface_id"`
	CodeSymbolID        semantic.ID `json:"code_symbol_id"`
	SemanticOwnerID     semantic.ID `json:"semantic_owner_id"`
	BeforeBindingDigest string      `json:"before_binding_digest"`
	AfterBindingDigest  string      `json:"after_binding_digest"`
	BeforeBlobDigest    string      `json:"before_blob_digest"`
	AfterBlobDigest     string      `json:"after_blob_digest"`
	BeforeSourcePath    string      `json:"before_source_path"`
	AfterSourcePath     string      `json:"after_source_path"`
}

type ChangeManifest struct {
	Schema               string          `json:"schema"`
	Complete             bool            `json:"complete"`
	ZeroChange           bool            `json:"zero_change"`
	RegistryDigest       string          `json:"registry_digest"`
	ToolchainDigest      string          `json:"toolchain_digest"`
	ProfileDigest        string          `json:"profile_digest"`
	BeforeSnapshotDigest string          `json:"before_snapshot_digest"`
	AfterSnapshotDigest  string          `json:"after_snapshot_digest"`
	Entries              []ManifestEntry `json:"entries"`
	Digest               string          `json:"digest"`
}

type BaselineConfig struct {
	Schema            string `json:"schema"`
	FullSuiteRequired bool   `json:"full_suite_required"`
	Digest            string `json:"digest"`
}

type Config struct {
	Schema                  string         `json:"schema"`
	RegistryDigest          string         `json:"registry_digest"`
	ToolchainDigest         string         `json:"toolchain_digest"`
	ProfileDigest           string         `json:"profile_digest"`
	SnapshotDigest          string         `json:"snapshot_digest"`
	ExpectedProviderDigest  string         `json:"expected_provider_digest"`
	ExpectedObserverDigest  string         `json:"expected_observer_digest"`
	Baseline                BaselineConfig `json:"baseline"`
	ExternalReceiptRequired bool           `json:"external_receipt_required"`
}

// ApplicabilityProof is evaluator-owned evidence that an empty registry is
// applicable to one exact immutable snapshot and policy tuple.
type ApplicabilityProof struct {
	Schema          string
	RegistryDigest  string
	ToolchainDigest string
	ProfileDigest   string
	SnapshotDigest  string
	AllowsEmpty     bool
	Digest          string
}

// AuthorityContext is supplied by the evaluator owner, never decoded from an
// Input packet. Its values define the registry, policy, snapshot, applicability
// and resource obligations against which producer claims are compared.
type AuthorityContext struct {
	Schema                  string
	Registry                Registry
	ToolchainDigest         string
	ProfileDigest           string
	SnapshotDigest          string
	ExpectedProviderDigest  string
	ExpectedObserverDigest  string
	Baseline                BaselineConfig
	Applicability           *ApplicabilityProof
	ExternalReceiptRequired bool
}

type AuthoritySource struct {
	SourceID semantic.ID `json:"source_id"`
	Path     string      `json:"path"`
	Span     string      `json:"span,omitempty"`
}

const (
	ReceiptSemanticDelta   = "SEMANTIC_DELTA"
	ReceiptNoSemanticDelta = "NO_SEMANTIC_DELTA"
	ReceiptStateCurrent    = "CURRENT"
)

type ChangeClaim string

const (
	ChangeClaimDelta   ChangeClaim = "DELTA"
	ChangeClaimNoDelta ChangeClaim = "NO_DELTA"
)

type CouplingReceipt struct {
	Schema                        string                       `json:"schema"`
	ReceiptID                     semantic.ID                  `json:"receipt_id"`
	SurfaceID                     semantic.ID                  `json:"surface_id"`
	SemanticOwnerID               semantic.ID                  `json:"semantic_owner_id"`
	CodeSymbolID                  semantic.ID                  `json:"code_symbol_id"`
	SourceMapBindingDigest        string                       `json:"source_map_binding_digest"`
	SnapshotDigest                string                       `json:"snapshot_digest"`
	RegistryDigest                string                       `json:"registry_digest"`
	ToolchainDigest               string                       `json:"toolchain_digest"`
	ProfileDigest                 string                       `json:"profile_digest"`
	BeforeBlobDigest              string                       `json:"before_blob_digest"`
	AfterBlobDigest               string                       `json:"after_blob_digest"`
	BeforeAuthoritySourceDigest   string                       `json:"before_authority_source_digest"`
	AfterAuthoritySourceDigest    string                       `json:"after_authority_source_digest"`
	BeforeCanonicalSemanticDigest string                       `json:"before_canonical_semantic_digest"`
	AfterCanonicalSemanticDigest  string                       `json:"after_canonical_semantic_digest"`
	ChangeClaim                   ChangeClaim                  `json:"change_claim"`
	ReceiptKind                   semantic.SemanticChangeKind  `json:"receipt_kind"`
	CanonicalDelta                string                       `json:"canonical_delta,omitempty"`
	DeltaDigest                   string                       `json:"delta_digest,omitempty"`
	AuthoritativeSource           *AuthoritySource             `json:"authoritative_source,omitempty"`
	OriginPathIDs                 []semantic.ID                `json:"origin_path_ids"`
	InferenceClaimID              semantic.ID                  `json:"inference_claim_id"`
	EvidenceRefs                  []semantic.EvidenceReference `json:"evidence_refs"`
	State                         string                       `json:"state"`
}

type ExternalResourceReceipt struct {
	Schema                 string  `json:"schema"`
	SnapshotDigest         string  `json:"snapshot_digest"`
	ProviderDigest         string  `json:"provider_digest"`
	ObserverDigest         string  `json:"observer_digest"`
	CPUWorkUnits           *uint64 `json:"cpu_work_units,omitempty"`
	PeakMemoryBytes        *uint64 `json:"peak_memory_bytes,omitempty"`
	DeterministicWorkUnits *uint64 `json:"deterministic_work_units,omitempty"`
	Digest                 string  `json:"digest"`
}

type Input struct {
	Schema          string                   `json:"schema"`
	Config          Config                   `json:"config"`
	Registry        Registry                 `json:"registry"`
	Manifest        ChangeManifest           `json:"manifest"`
	Receipts        []CouplingReceipt        `json:"receipts"`
	InferencePath   semantic.InferencePathV1 `json:"inference_path"`
	ExternalReceipt *ExternalResourceReceipt `json:"external_receipt,omitempty"`
	WorkspaceRoot   string                   `json:"workspace_root,omitempty"`
}

type CountDimension struct {
	Known bool   `json:"known"`
	Value uint64 `json:"value"`
}

type ObservationVector struct {
	ChangedSurfaces   CountDimension `json:"changed_surfaces"`
	Receipts          CountDimension `json:"receipts"`
	InferenceRecords  CountDimension `json:"inference_records"`
	InferencePaths    CountDimension `json:"inference_paths"`
	DeterministicWork CountDimension `json:"deterministic_work"`
	ResourceWork      CountDimension `json:"resource_work"`
	CPU               CountDimension `json:"cpu"`
	Memory            CountDimension `json:"memory"`
}

type Result struct {
	Schema             string            `json:"schema"`
	Status             Status            `json:"status"`
	AcceptedSurfaceIDs []semantic.ID     `json:"accepted_surface_ids"`
	Reasons            []Reason          `json:"reasons"`
	Observation        ObservationVector `json:"observation"`
	FullSuiteRequired  bool              `json:"full_suite_required"`
	InputDigest        string            `json:"input_digest"`
	Digest             string            `json:"digest"`
}
