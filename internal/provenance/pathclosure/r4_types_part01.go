package pathclosure

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// R4SchemaVersion identifies the finite, receipt-bound path contract. It is
// deliberately separate from the semantic inference-path schema.
const R4SchemaVersion = "gooo-path-closure-r4/v1"

type R4Phase string

const (
	R4CompilePhase R4Phase = "compile"
	R4RuntimePhase R4Phase = "runtime"
)
const (
	CodeInvalidPath            = "INVALID_PATH"
	CodeConflictingReceipt     = "CONFLICTING_RECEIPT"
	CodeMissingProvider        = "MISSING_PROVIDER_BINDING"
	CodeMissingProviderBinding = CodeMissingProvider
	CodePhaseMismatch          = "PHASE_MISMATCH"
	CodeMissingObserver        = "MISSING_OBSERVER"
	CodeOpenWorld              = "OPEN_WORLD"
	CodeUnexhaustedBoundary    = "UNEXHAUSTED_BOUNDARY"
	CodeMissingRequiredPaths   = "MISSING_REQUIRED_PATHS"
)

// R4Record is one canonical producer record. Its ReceiptID is an explicit
// binding to the observer/provider receipt; it is never inferred by lookup.
type R4Record struct {
	ID             semantic.ID
	SubjectID      semantic.ID
	ObjectID       semantic.ID
	ProviderID     semantic.ID
	ProviderDigest string
	Phase          R4Phase
	PhaseDigest    string
	PredecessorID  semantic.ID
	ReceiptID      semantic.ID
	Writes         bool
	Effect         string
}

// R4Receipt is append-only evidence for exactly one record. EventID is kept
// distinct from ReceiptID so replay cannot silently reuse a physical event.
type R4Receipt struct {
	ID             semantic.ID
	EventID        semantic.ID
	RecordID       semantic.ID
	ProviderID     semantic.ID
	ProviderDigest string
	Phase          R4Phase
	PhaseDigest    string
	ObserverID     semantic.ID
	Writes         bool
	Effect         string
}

// R4Path is an ordered, finite claim. RecordBytes are the exact canonical JSON
// bytes that must be recomputed for the ordered RecordIDs.
type R4Path struct {
	ID          semantic.ID
	StartID     semantic.ID
	EndID       semantic.ID
	RecordIDs   []semantic.ID
	RecordBytes []string
}
type R4Boundary struct {
	RequiredPathIDs []semantic.ID
	Exhausted       bool
	OpenWorld       bool
}
