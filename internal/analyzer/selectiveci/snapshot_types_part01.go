package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
)

const (
	// SchemaV1 is the versioned canonical JSON schema for Snapshot.
	SchemaV1 = "gooo/selective-ci-snapshot/v1"

	// FormatVersion is an alias for callers that use format vocabulary.
	FormatVersion = SchemaV1
)

// Status is the state of a source-backed selective-CI result.
type Status string

const (
	StatusBound   Status = "BOUND"
	StatusUnknown Status = "UNKNOWN"
)

// ErrorCode identifies a fail-closed construction or validation failure.
type ErrorCode string

const (
	CodeInput             ErrorCode = "selectiveci.input"
	CodeMissingBinding    ErrorCode = "selectiveci.missing-binding"
	CodeAmbiguousBinding  ErrorCode = "selectiveci.ambiguous-binding"
	CodeUnregisteredID    ErrorCode = "selectiveci.unregistered-id"
	CodeDuplicateBinding  ErrorCode = "selectiveci.duplicate-binding"
	CodeMalformedPath     ErrorCode = "selectiveci.malformed-path"
	CodeMalformedDigest   ErrorCode = "selectiveci.malformed-digest"
	CodeCandidateIdentity ErrorCode = "selectiveci.candidate-identity"
	CodeDerivedIdentity   ErrorCode = "selectiveci.derived-identity"
	CodeStaleSnapshot     ErrorCode = "selectiveci.stale-snapshot"
	CodeInvalidBinding    ErrorCode = "selectiveci.invalid-binding"
	CodeInvalidStatus     ErrorCode = "selectiveci.invalid-status"
	CodeInvalidSchema     ErrorCode = "selectiveci.invalid-schema"
)

// Error is a deterministic fail-closed result. A construction or diff error
// always requires the caller to use its full-suite fallback.
type Error struct {
	Code              ErrorCode
	Message           string
	FullSuiteFallback bool
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return string(e.Code) + ": " + e.Message
}

// SourceInput is one explicit repository source and its semanticbinding
// records. Path is a repository-relative source-map path; no name lookup is
// performed to associate a binding with a source.
type SourceInput struct {
	Path       string
	BlobDigest string
	// Bindings must be non-nil. An explicit empty slice is a validated
	// unbound source; nil means the binding authority was not supplied.
	Bindings []semanticbinding.Binding
}
