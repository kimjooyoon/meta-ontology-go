package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
	OriginPathIDs               []string    `json:"origin_path_ids"`
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
