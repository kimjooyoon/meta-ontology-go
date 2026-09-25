package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

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
