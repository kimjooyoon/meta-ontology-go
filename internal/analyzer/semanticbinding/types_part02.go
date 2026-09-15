package semanticbinding

// Obligation is one explicit gooo:obligation declaration.
type Obligation struct {
	ID              string
	Subject         string
	Pressure        string
	PackagePath     string
	DeclarationKey  string
	Span            Span
	DirectiveSpan   Span
	Digest          string
	CanonicalDigest string
}

// Result is the complete, sorted extraction result. Failed extraction returns
// UNKNOWN with FullSuiteFallback set and no partial records.
type Result struct {
	Status            Status
	Bindings          []Binding
	Obligations       []Obligation
	Digest            string
	CanonicalDigest   string
	Unknowns          []Unknown
	FullSuiteFallback bool
}

// Unknown retains a deterministic reason for a result that could not be
// established from the supplied source and explicit inputs.
type Unknown struct {
	Code              Code
	Message           string
	Span              Span
	FullSuiteFallback bool
}

// Code identifies a strict extraction failure.
type Code string

const (
	CodeInput              Code = "semanticbinding.input"
	CodeParse              Code = "semanticbinding.parse"
	CodeTypeCheck          Code = "semanticbinding.type-check"
	CodeDetachedComment    Code = "semanticbinding.detached-comment"
	CodeUnknownDirective   Code = "semanticbinding.unknown-directive"
	CodeMalformedDirective Code = "semanticbinding.malformed-directive"
	CodeUnknownField       Code = "semanticbinding.unknown-field"
	CodeDuplicateField     Code = "semanticbinding.duplicate-field"
	CodeMissingField       Code = "semanticbinding.missing-field"
	CodeInvalidRole        Code = "semanticbinding.invalid-role"
	CodeInvalidIdentity    Code = "semanticbinding.invalid-identity"
	CodeDuplicateID        Code = "semanticbinding.duplicate-id"
	CodeUnregisteredID     Code = "semanticbinding.unregistered-id"
	CodeAmbiguousBinding   Code = "semanticbinding.ambiguous-binding"
	CodeMissingObject      Code = "semanticbinding.missing-object"
)

// Error is a source-backed strict extraction error. Every Error is also an
// UNKNOWN result and requires the caller's full-suite fallback policy.
type Error struct {
	Code              Code
	Message           string
	Span              Span
	FullSuiteFallback bool
}
