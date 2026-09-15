package coupling

import (
	"context"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/query/couplingexplain"
)

// SourceLocation is a source-map-backed location supplied by the caller. It
// is not inferred from a query path, name, alias, or presentation field.
// StableID and SourceMapID are validation inputs and never appear on the LSP
// wire.
type SourceLocation struct {
	StableID        string
	SourceMapID     string
	SourceMapDigest string
	URI             string
	Range           Range
	Label           string
	Message         string
}

// LocationSnapshot binds every source location used by the adapter to one
// exact snapshot and LSP document version. The production query envelope
// does not contain URI/range data, so a verified link cannot be emitted
// without this explicit projection.
type LocationSnapshot struct {
	SnapshotDigest  string
	DocumentURI     string
	DocumentVersion int
	Locations       []SourceLocation
}

// LiveRequest combines the LSP request with the immutable query request. The
// query request's control versions must be equal and must match the exact LSP
// document version before a link can be returned.
type LiveRequest struct {
	Context         context.Context
	DocumentURI     string
	DocumentVersion int
	Position        Position
	SnapshotDigest  string
	Query           couplingexplain.Request
	Locations       LocationSnapshot
}

// LiveAdapter consumes only the production couplingexplain envelope bytes.
// It has no write, rename, filesystem, graph, or semantic-authority surface.
type LiveAdapter struct {
	raw []byte
}

// NewLiveAdapter validates and copies one immutable production query envelope.
func NewLiveAdapter(data []byte) (*LiveAdapter, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("coupling query: empty envelope")
	}
	if _, err := couplingexplain.DecodeVerifiedEnvelope(data); err != nil {
		return nil, fmt.Errorf("coupling query: decode: %w", err)
	}
	return &LiveAdapter{raw: append([]byte(nil), data...)}, nil
}
