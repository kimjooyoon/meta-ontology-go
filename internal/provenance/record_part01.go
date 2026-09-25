package provenance

// SchemaVersion is the wire schema used by the ledger.
const SchemaVersion = 1

// EvidenceKind classifies the fact without granting it authority.
type EvidenceKind string

const (
	KindCompilerRun  EvidenceKind = "compiler-run"
	KindVerification EvidenceKind = "verification"
	KindComparison   EvidenceKind = "comparison"
	KindObservation  EvidenceKind = "observation"
)

// EvidenceStatus is deliberately separate from EvidenceKind. Candidate and
// deferred observations can be stored, but neither can satisfy a verified
// claim.
type EvidenceStatus string

const (
	StatusVerified  EvidenceStatus = "verified"
	StatusCandidate EvidenceStatus = "candidate"
	StatusDeferred  EvidenceStatus = "deferred"
	StatusFailed    EvidenceStatus = "failed"
	StatusRejected  EvidenceStatus = "rejected"
)

// Position identifies a source location. Offsets are zero-based; lines and
// columns are one-based, matching the compiler's source-span convention.
type Position struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

// SourceSpan binds an observation to the source that produced it.
type SourceSpan struct {
	URI   string   `json:"uri"`
	File  string   `json:"-"`
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// DigestLink is one immutable predecessor reference in the ledger chain.
type DigestLink struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
}
