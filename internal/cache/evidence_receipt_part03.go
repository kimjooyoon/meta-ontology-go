package cache

// Matches reports whether evidence belongs to the requested current base,
// head, run, and predecessor tuple.
func (e EvidenceFreshness) Matches(current EvidenceFreshness) bool {
	return e.Equal(current)
}

// CacheReceipt is an append-only immutable record of one cache decision.
type CacheReceipt struct {
	SchemaVersion         string            `json:"schema_version"`
	CacheKey              Digest            `json:"cache_key"`
	Domain                string            `json:"domain"`
	KeyVersion            string            `json:"key_version"`
	HostStage             HostStage         `json:"host_stage"`
	ArtifactKind          string            `json:"artifact_kind"`
	Projection            string            `json:"projection"`
	SemanticClosureDigest Digest            `json:"semantic_closure_digest"`
	DependencyRoot        Digest            `json:"dependency_root"`
	DirectDependencies    []Digest          `json:"direct_dependencies"`
	PolicySchemaDigest    Digest            `json:"policy_schema_digest"`
	Toolchain             string            `json:"toolchain"`
	Target                string            `json:"target"`
	BuildTagsDigest       Digest            `json:"build_tags_digest"`
	OptionsDigest         Digest            `json:"options_digest"`
	ContentDigest         Digest            `json:"content_digest"`
	Size                  int64             `json:"size"`
	Reconstructable       bool              `json:"reconstructable"`
	EvidenceRefs          []EvidenceRef     `json:"evidence_refs"`
	ProducerHost          string            `json:"producer_host"`
	Status                ReceiptStatus     `json:"status"`
	Evidence              EvidenceFreshness `json:"evidence"`
	ReceiptDigest         Digest            `json:"receipt_digest"`
}

// ReceiptStatus describes the observed cache outcome, never feature proof.
type ReceiptStatus string

const (
	ReceiptHit        ReceiptStatus = "hit"
	ReceiptMiss       ReceiptStatus = "miss"
	ReceiptRecomputed ReceiptStatus = "recomputed"
	ReceiptStale      ReceiptStatus = "stale"
	ReceiptCorrupt    ReceiptStatus = "corrupt"
)
