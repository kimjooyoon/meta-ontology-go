package coupling

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
