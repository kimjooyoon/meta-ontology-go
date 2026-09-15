package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
)

const (
	SchemaVersion         = "gooo/selective-ci/v1"
	ManifestSchemaVersion = "gooo/selective-ci-manifest/v1"
	RegistrySchemaVersion = "gooo/selective-ci-registry/v1"
)

type Status string

const (
	StatusSelective         Status = "SELECTIVE"
	StatusFullSuiteFallback Status = "FULL_SUITE_FALLBACK"
	SELECTIVE                      = StatusSelective
	FULL_SUITE_FALLBACK            = StatusFullSuiteFallback
)
const (
	ReasonNone               = ""
	ReasonUnsupportedSchema  = "UNSUPPORTED_SCHEMA"
	ReasonInvalidInput       = "INVALID_INPUT"
	ReasonUnknownPath        = "UNKNOWN_PATH"
	ReasonMissingBinding     = "MISSING_BINDING"
	ReasonMismatchedDigest   = "MISMATCHED_DIGEST"
	ReasonDuplicateID        = "DUPLICATE_ID"
	ReasonDanglingReference  = "DANGLING_REFERENCE"
	ReasonCycle              = "CYCLE"
	ReasonAmbiguousPath      = "AMBIGUOUS_PATH"
	ReasonInvalidArgv        = "INVALID_ARGV"
	ReasonResourceReceipt    = "INVALID_RESOURCE_RECEIPT"
	ReasonResourceArithmetic = "RESOURCE_ARITHMETIC_ERROR"
	ReasonResourceLimit      = "RESOURCE_LIMIT_EXCEEDED"
	ReasonEvaluatorError     = "EVALUATOR_ERROR"
	ReasonFrontierBlocked    = "FRONTIER_BLOCKED"
	ReasonMissingObligation  = "MISSING_OBLIGATION"
	ReasonMissingCommand     = "MISSING_COMMAND"
	ReasonDanglingCommand    = "DANGLING_COMMAND"
	ReasonUnknownRoot        = "UNKNOWN_ROOT"
	ReasonDuplicateRoot      = "DUPLICATE_ROOT"
	ReasonStaleGraph         = "STALE_GRAPH"
	ReasonStaleRegistry      = "STALE_REGISTRY"
	ReasonStaleSnapshot      = "STALE_SNAPSHOT"
	ReasonWorkOverflow       = "WORK_OVERFLOW"
)

type SnapshotFile struct {
	Path        string   `json:"path"`
	BlobDigest  string   `json:"blob_digest"`
	SemanticIDs []string `json:"semantic_ids"`
}
type SnapshotManifest struct {
	SchemaVersion string         `json:"schema_version"`
	Digest        string         `json:"snapshot_digest"`
	Files         []SnapshotFile `json:"files"`
}
type DependencyEdge struct {
	From string               `json:"from"`
	To   string               `json:"to"`
	Kind impactgraph.EdgeKind `json:"kind"`
}
type Command struct {
	ID           string   `json:"id"`
	Argv         []string `json:"argv"`
	WorkingDir   string   `json:"working_directory"`
	CPUWorkUnits uint64   `json:"cpu_work_units"`
	MemoryBytes  uint64   `json:"memory_bytes"`
}
