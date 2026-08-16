// Package coupling adapts immutable semantic-coupling explanations to
// standard Language Server Protocol values. The input is a read-only,
// snapshot-bound projection; it is never a source of semantic authority.
package coupling

import (
	"context"
)

const SchemaVersion = "gooo/lsp-coupling-explanation/v1"

const diagnosticSource = "gooo/coupling"

// Outcome is the closed result set of the upstream explanation projection.
type Outcome string

const (
	OutcomePass       Outcome = "PASS"
	OutcomeUnknown    Outcome = "UNKNOWN"
	OutcomeFailClosed Outcome = "FAIL_CLOSED"
)

func (o Outcome) valid() bool {
	return o == OutcomePass || o == OutcomeUnknown || o == OutcomeFailClosed
}

// ChangeClaim is deliberately separate from inference edge kinds.
type ChangeClaim string

const (
	ClaimDelta   ChangeClaim = "DELTA"
	ClaimNoDelta ChangeClaim = "NO_DELTA"
)

func (c ChangeClaim) valid() bool { return c == ClaimDelta || c == ClaimNoDelta }

// Reason values are upstream failure partitions. The adapter does not
// collapse them into a successful no-op.
type Reason string

const (
	ReasonAmbiguous       Reason = "AMBIGUOUS"
	ReasonStaleSnapshot   Reason = "STALE_SNAPSHOT"
	ReasonUnregistered    Reason = "UNREGISTERED"
	ReasonMissing         Reason = "MISSING"
	ReasonUpstreamUnknown Reason = "UPSTREAM_UNKNOWN"
	ReasonUpstreamFail    Reason = "UPSTREAM_FAIL"
)

func (r Reason) valid() bool {
	switch r {
	case ReasonAmbiguous, ReasonStaleSnapshot, ReasonUnregistered, ReasonMissing,
		ReasonUpstreamUnknown, ReasonUpstreamFail:
		return true
	default:
		return false
	}
}

// Position and Range use the standard LSP zero-based UTF-16 representation.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// BoundLocation is input-only metadata. Stable IDs are required to validate
// the immutable explanation, but are intentionally not present in output LSP
// locations.
type BoundLocation struct {
	StableID        string `json:"stable_id"`
	SourceMapID     string `json:"source_map_id"`
	SourceMapDigest string `json:"source_map_digest"`
	URI             string `json:"uri"`
	Range           Range  `json:"range"`
	Label           string `json:"label,omitempty"`
}

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

// LocationLink is the standard LSP definition response shape.
type LocationLink struct {
	OriginSelectionRange *Range `json:"originSelectionRange,omitempty"`
	TargetURI            string `json:"targetUri"`
	TargetRange          Range  `json:"targetRange"`
	TargetSelectionRange Range  `json:"targetSelectionRange"`
}

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

type Diagnostic struct {
	Range              Range                          `json:"range"`
	Severity           int                            `json:"severity,omitempty"`
	Code               string                         `json:"code,omitempty"`
	Source             string                         `json:"source,omitempty"`
	Message            string                         `json:"message"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
}

const (
	DiagnosticError       = 1
	DiagnosticWarning     = 2
	DiagnosticInformation = 3
)

const (
	DiagnosticExplanation          = "gooo.coupling.explanation"
	DiagnosticMissingCancellation  = "gooo.coupling.missing-cancellation"
	DiagnosticMissingSnapshot      = "gooo.coupling.missing-snapshot"
	DiagnosticMissingVersion       = "gooo.coupling.missing-document-version"
	DiagnosticDocumentMismatch     = "gooo.coupling.document-mismatch"
	DiagnosticWrongVersion         = "gooo.coupling.wrong-document-version"
	DiagnosticStaleSnapshot        = "gooo.coupling.stale-snapshot"
	DiagnosticCancelled            = "gooo.coupling.cancelled"
	DiagnosticAmbiguous            = "gooo.coupling.ambiguous"
	DiagnosticUpstreamUnknown      = "gooo.coupling.upstream-unknown"
	DiagnosticUpstreamFail         = "gooo.coupling.upstream-fail-closed"
	DiagnosticNoBinding            = "gooo.coupling.no-binding"
	DiagnosticInvalidEnvelope      = "gooo.coupling.invalid-envelope"
	DiagnosticInvalidPosition      = "gooo.coupling.invalid-position"
	DiagnosticLiveInvalidEnvelope  = "gooo.coupling.query-invalid-envelope"
	DiagnosticLiveInvalidBinding   = "gooo.coupling.query-invalid-binding"
	DiagnosticLiveMissingLocations = "gooo.coupling.missing-source-locations"
	DiagnosticLiveMissingMaterial  = "gooo.coupling.missing-verified-material"
	DiagnosticLiveNotVerified      = "gooo.coupling.not-independently-verified"
)

// Result is an in-process aggregation. Its fields are independently
// serializable standard LSP values; Result itself is not a custom wire
// envelope.
type Result struct {
	Outcome     Outcome
	Links       []LocationLink
	Hover       *Hover
	Diagnostics []Diagnostic
}

type Adapter struct {
	envelope Envelope
	raw      []byte
}
