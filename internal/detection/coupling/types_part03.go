package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
