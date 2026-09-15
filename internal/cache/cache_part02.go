package cache

import (
	"sync"
	"time"
)

// Metadata is the integrity and provenance envelope for a cached projection.
// CreatedAt is observational metadata and is not part of the content address.
type Metadata struct {
	FormatVersion         string    `json:"format_version"`
	Key                   string    `json:"key"`
	KeyVersion            string    `json:"key_version"`
	Domain                string    `json:"domain"`
	Namespace             string    `json:"namespace"`
	ArtifactKind          string    `json:"artifact_kind"`
	ToolVersion           string    `json:"tool_version"`
	Toolchain             string    `json:"toolchain"`
	Target                string    `json:"target"`
	HostStage             HostStage `json:"host_stage"`
	InputDigest           Digest    `json:"input_digest"`
	SemanticClosureDigest Digest    `json:"semantic_closure_digest"`
	DependencyRoot        Digest    `json:"dependency_root"`
	PolicySchemaDigest    Digest    `json:"policy_schema_digest"`
	BuildTagsDigest       Digest    `json:"build_tags_digest"`
	OptionsDigest         Digest    `json:"options_digest"`
	DependencyDigest      Digest    `json:"dependency_digest"`
	ProvenanceDigest      Digest    `json:"provenance_digest"`
	ArtifactType          string    `json:"artifact_type,omitempty"`
	Projection            string    `json:"projection,omitempty"`
	Reconstructable       bool      `json:"reconstructable"`
	Size                  int64     `json:"size"`
	ContentDigest         Digest    `json:"content_digest"`
	MetadataDigest        Digest    `json:"metadata_digest"`
	CreatedAt             time.Time `json:"created_at"`
}

// InvalidationFilter selects cached projections to remove. Empty fields are
// wildcards, but at least one field must be set. Use Clear to remove all.
type InvalidationFilter struct {
	Namespace    string
	KeyVersion   string
	ToolVersion  string
	ArtifactType string
	Projection   string
}

// Cache is a content-addressed, filesystem-backed projection cache.
type Cache struct {
	root         string
	objects      string
	maxEntrySize int64
	filesystemMu sync.RWMutex
	locksMu      sync.Mutex
	locks        map[string]*entryLock
	receiptMu    sync.Mutex
	receipts     string
}
type entryLock struct {
	mu   sync.Mutex
	refs int
}

// New creates or opens a cache rooted at root.
func New(root string) (*Cache, error) { return Open(root) }
