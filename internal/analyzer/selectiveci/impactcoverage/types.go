// Package impactcoverage observes changed source blobs and their explicit
// stable-ID bindings for selective-CI admission.
package impactcoverage

import "github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"

const SchemaV1 = "gooo/selective-ci-impact-coverage/v1"

type Decision string

const (
	DecisionExact   Decision = "EXACT"
	DecisionUnknown Decision = "UNKNOWN"
)

type Reason string

const (
	ReasonComplete        Reason = "COMPLETE"
	ReasonNoChange        Reason = "NO_CHANGE"
	ReasonMissingBinding  Reason = "MISSING_BINDING"
	ReasonAuthorityDrift  Reason = "AUTHORITY_DRIFT"
	ReasonInvalidSnapshot Reason = "INVALID_SNAPSHOT"
)

// Input contains only the two validated source snapshots. Nil is an explicit
// missing authority value and is never treated as an empty snapshot.
type Input struct {
	Schema string                `json:"schema"`
	Base   *selectiveci.Snapshot `json:"base"`
	Head   *selectiveci.Snapshot `json:"head"`
}

// Result is a deterministic observation. Stable IDs are intentionally empty
// whenever DecisionUnknown so callers cannot use partial analysis to select.
type Result struct {
	Schema                    string   `json:"schema"`
	Decision                  Decision `json:"decision"`
	Reason                    Reason   `json:"reason"`
	FullSuiteRequired         bool     `json:"full_suite_required"`
	ChangedBlobCount          uint64   `json:"changed_blob_count"`
	CoveredChangedBlobCount   uint64   `json:"covered_changed_blob_count"`
	UncoveredChangedBlobCount uint64   `json:"uncovered_changed_blob_count"`
	ChangedBindingCount       uint64   `json:"changed_binding_count"`
	DeterministicWorkUnits    uint64   `json:"deterministic_work_units"`
	ChangedStableIDs          []string `json:"changed_stable_ids"`
	UncoveredPaths            []string `json:"uncovered_paths"`
	BaseSnapshotDigest        string   `json:"base_snapshot_digest"`
	HeadSnapshotDigest        string   `json:"head_snapshot_digest"`
	BaseSourceMapDigest       string   `json:"base_source_map_digest"`
	HeadSourceMapDigest       string   `json:"head_source_map_digest"`
	BaseRegistryDigest        string   `json:"base_registry_digest"`
	HeadRegistryDigest        string   `json:"head_registry_digest"`
	InputDigest               string   `json:"input_digest"`
	OutputDigest              string   `json:"output_digest"`
}

type Output = Result

// NewInput is the explicit constructor used by non-JSON callers.
func NewInput(base, head *selectiveci.Snapshot) Input {
	return Input{Schema: SchemaV1, Base: base, Head: head}
}
