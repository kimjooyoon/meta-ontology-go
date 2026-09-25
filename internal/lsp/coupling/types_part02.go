package coupling

import (
	"context"
)

// CausalSpan is an exact contributing source span from the immutable query
// result. Ordinal preserves the query's causal ordering; StableID binds the
// span without making it a custom LSP wire field.
type CausalSpan struct {
	StableID        string `json:"stable_id"`
	SourceMapID     string `json:"source_map_id"`
	SourceMapDigest string `json:"source_map_digest"`
	URI             string `json:"uri"`
	Range           Range  `json:"range"`
	Ordinal         int    `json:"ordinal"`
	Message         string `json:"message,omitempty"`
}
type Document struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

// Explanation is one source-bound coupling explanation. Names and labels are
// presentation only; selection is performed by its exact origin range.
type Explanation struct {
	CodeSymbolID    string        `json:"code_symbol_id"`
	SemanticOwnerID string        `json:"semantic_owner_id"`
	Label           string        `json:"label,omitempty"`
	Origin          BoundLocation `json:"origin"`
	Target          BoundLocation `json:"target"`
	CausalSpans     []CausalSpan  `json:"causal_spans"`
	Claim           ChangeClaim   `json:"claim"`
	Status          Outcome       `json:"status"`
	Reason          Reason        `json:"reason,omitempty"`
}

// Envelope is the future immutable query/coupling explanation byte contract
// consumed by this package. It is intentionally a projection envelope, not a
// semantic authority or a writable document model.
type Envelope struct {
	Schema               string        `json:"schema"`
	SnapshotDigest       string        `json:"snapshot_digest"`
	RegistryDigest       string        `json:"registry_digest"`
	ToolchainDigest      string        `json:"toolchain_digest"`
	ProfileDigest        string        `json:"profile_digest"`
	DetectorResultDigest string        `json:"detector_result_digest"`
	OracleResultDigest   string        `json:"oracle_result_digest"`
	EvidenceDigest       string        `json:"evidence_digest"`
	Document             Document      `json:"document"`
	Status               Outcome       `json:"status"`
	Reason               Reason        `json:"reason,omitempty"`
	Explanations         []Explanation `json:"explanations"`
}

// Request carries every freshness and cancellation input required for a
// read. A nil Context, empty snapshot digest, or non-positive version is not
// treated as an implicit current value.
type Request struct {
	Context         context.Context
	DocumentURI     string
	DocumentVersion int
	Position        Position
	SnapshotDigest  string
}

// Location is the standard LSP location shape.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}
