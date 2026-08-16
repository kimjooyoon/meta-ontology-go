package semanticbinding

import (
	"fmt"
	"strings"
)

// Role is the implementation role of a Go declaration.
type Role string

const (
	RoleHandwrittenImpl Role = "HANDWRITTEN_IMPL"
	RoleGeneratedImpl   Role = "GENERATED_IMPL"
	RoleAdapter         Role = "ADAPTER"

	HandwrittenImpl = RoleHandwrittenImpl
	GeneratedImpl   = RoleGeneratedImpl
	Adapter         = RoleAdapter
)

// Status describes whether extraction produced a complete result.
type Status string

const (
	StatusBound   Status = "BOUND"
	StatusUnknown Status = "UNKNOWN"
)

// SourceFile is one explicit Go source input. PackagePath is required unless
// Input.PackagePath supplies the same path for every source.
type SourceFile struct {
	Filename    string
	PackagePath string
	Source      []byte
}

// Source is a concise alias for SourceFile.
type Source = SourceFile

// Input supplies one Go package and, optionally, an explicit ID registry.
// A non-nil RegisteredIDs slice makes registry membership part of validation;
// an ID absent from it produces UNKNOWN rather than an inferred binding.
type Input struct {
	Sources       []SourceFile
	Files         []SourceFile
	PackagePath   string
	RegisteredIDs []string
}

// Position is a token.FileSet position with a zero-based byte offset and
// one-based line and column numbers.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Span is a half-open source range.
type Span struct {
	Filename string
	Start    Position
	End      Position
}

// SourceSpan is a vocabulary alias for Span.
type SourceSpan = Span

// Binding is one explicit gooo:bind declaration binding.
type Binding struct {
	ID              string
	Role            Role
	PackagePath     string
	DeclarationKey  string
	Span            Span
	DirectiveSpan   Span
	Digest          string
	CanonicalDigest string
}

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

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Span.Filename == "" {
		return string(e.Code) + ": " + e.Message
	}
	return fmt.Sprintf("%s: %s: %s", e.Span.Filename, e.Span, e.Message)
}

func (s Span) String() string {
	if s.Filename == "" {
		return fmt.Sprintf("%d:%d-%d:%d", s.Start.Line, s.Start.Column, s.End.Line, s.End.Column)
	}
	return fmt.Sprintf("%s:%d:%d-%d:%d", s.Filename, s.Start.Line, s.Start.Column, s.End.Line, s.End.Column)
}

func (r Role) valid() bool {
	return r == RoleHandwrittenImpl || r == RoleGeneratedImpl || r == RoleAdapter
}

func (s Span) valid() bool {
	return s.Filename != "" && s.Start.Offset >= 0 && s.End.Offset >= s.Start.Offset
}

func (i Input) sourceInputs() ([]SourceFile, error) {
	if len(i.Sources) > 0 && len(i.Files) > 0 {
		return nil, &Error{Code: CodeInput, Message: "use Sources or Files, not both", FullSuiteFallback: true}
	}
	sources := i.Sources
	if len(sources) == 0 {
		sources = i.Files
	}
	if len(sources) == 0 {
		return nil, &Error{Code: CodeInput, Message: "at least one source file is required", FullSuiteFallback: true}
	}
	result := make([]SourceFile, len(sources))
	copy(result, sources)
	for index := range result {
		if result[index].PackagePath == "" {
			result[index].PackagePath = i.PackagePath
		}
		if strings.TrimSpace(result[index].Filename) == "" || strings.TrimSpace(result[index].PackagePath) == "" {
			return nil, &Error{Code: CodeInput, Message: "filename and package path are required", FullSuiteFallback: true}
		}
		if result[index].Source == nil {
			return nil, &Error{Code: CodeInput, Message: "source bytes are required", FullSuiteFallback: true}
		}
		result[index].Filename = strings.TrimSpace(result[index].Filename)
		result[index].PackagePath = strings.TrimSpace(result[index].PackagePath)
	}
	return result, nil
}
