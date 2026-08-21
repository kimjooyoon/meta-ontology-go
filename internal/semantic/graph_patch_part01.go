package semantic

import (
	"errors"
)

const (
	GraphPatchSchemaVersion                     = "gooo-graph-patch/v1"
	GraphPatchSetNodeField                      = "set-node-field"
	GraphPatchAddFact                           = "add-fact"
	PatchInvalidRequest       PatchConflictCode = "invalid_request"
	PatchStaleGraphHash       PatchConflictCode = "stale_graph_hash"
	PatchUnknownNode          PatchConflictCode = "unknown_node"
	PatchNodeHashMismatch     PatchConflictCode = "node_hash_mismatch"
	PatchUnknownField         PatchConflictCode = "unknown_field"
	PatchImmutableField       PatchConflictCode = "immutable_field"
	PatchFieldHashMismatch    PatchConflictCode = "field_hash_mismatch"
	PatchInvalidRelation      PatchConflictCode = "invalid_relation"
	PatchUnknownEndpoint      PatchConflictCode = "unknown_endpoint"
	PatchEndpointKindMismatch PatchConflictCode = "endpoint_kind_mismatch"
	PatchBaseTupleMismatch    PatchConflictCode = "base_tuple_mismatch"
	PatchIntentMissing        PatchConflictCode = "intent_missing"
	PatchLocalityMissing      PatchConflictCode = "locality_missing"
)

var ErrGraphPatch = errors.New("semantic graph patch rejected")

// PatchConflictCode identifies a deterministic, pre-write rejection reason.
type PatchConflictCode string

// GraphPatchBase is the trusted source/IR tuple observed by the validator.
type GraphPatchBase struct {
	SourceDigest string
	IRDigest     string
}

// GraphPatchRequest carries preconditions for one typed, not-yet-applied edit.
// This API validates only; it never changes graph, source, or store state.
type GraphPatchRequest struct {
	SchemaVersion        string
	Operation            string
	ExpectedGraphHash    string
	NodeID               ID
	ExpectedNodeHash     string
	Field                string
	ExpectedFieldHash    string
	Subject              ID
	Predicate            Relation
	Object               ID
	ExpectedSourceDigest string
	ExpectedIRDigest     string
	AllowedIntent        string
	Locality             string
}

// GraphPatchMutation is the in-memory payload applied after preconditions.
type GraphPatchMutation struct {
	Name    string
	Aliases []string
	Fact    *Fact
}

// GraphPatchConflict is a stable error code and human-readable detail.
type GraphPatchConflict struct {
	Code   PatchConflictCode
	Detail string
}
