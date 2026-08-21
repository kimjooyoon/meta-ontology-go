package coupling

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
