package analyzer

import (
	"fmt"
	"sort"
	"strings"
)

// SymbolKind is the semantic role registered for a Go symbol.
type SymbolKind string

const (
	KindActivity SymbolKind = "activity"
	KindEntity   SymbolKind = "entity"

	// Activity and Entity are concise aliases for callers constructing
	// registrations.
	Activity = KindActivity
	Entity   = KindEntity
)

// Relation is a semantic relation produced by the analyzer.
type Relation string

const (
	RelationInvokes    Relation = "invokes"
	RelationUses       Relation = "uses"
	RelationGenerates  Relation = "generates"
	RelationReferences Relation = "references"
)

// Identity is the stable semantic identity of a symbol. Namespace is kept as
// a separate field so callers can enforce namespace policy without parsing an
// ID. ID is the authoritative value; Go names are not identity.
type Identity struct {
	Namespace string
	ID        string
}

// NewIdentity is a convenience constructor for a stable semantic identity.
func NewIdentity(namespace, id string) Identity {
	return Identity{Namespace: namespace, ID: id}
}

// Valid reports whether the identity can participate in a fact.
func (i Identity) Valid() bool {
	return strings.TrimSpace(i.ID) != ""
}

// String returns the stable identity used by callers and diagnostics.
func (i Identity) String() string {
	return i.ID
}

// SymbolRef identifies a Go symbol without assigning semantic meaning to it.
// PackagePath is preferred for resolution. PackageName is retained for source
// files and registries that do not have an import path available.
type SymbolRef struct {
	PackagePath string
	PackageName string
	Receiver    string
	Name        string
}

func (r SymbolRef) canonical() string {
	return strings.Join([]string{r.PackagePath, r.PackageName, r.Receiver, r.Name}, "\x00")
}

// Registration binds one Go symbol to one semantic identity.
type Registration struct {
	Ref      SymbolRef
	Kind     SymbolKind
	Identity Identity
	Span     Span
}

// SourceFile is one Go source view supplied to AnalyzePackage.
type SourceFile struct {
	Filename    string
	PackagePath string
	Source      []byte
}

// DiagnosticCode identifies a deterministic analyzer diagnostic.
type DiagnosticCode string

const (
	DiagInvalidAnnotation     DiagnosticCode = "analyzer.invalid-annotation"
	DiagConflictingAnnotation DiagnosticCode = "analyzer.conflicting-annotation"
)

// Diagnostic describes a source-backed semantic-analysis problem. Invalid
// annotations never become registrations or semantic facts.
type Diagnostic struct {
	Code    DiagnosticCode
	Message string
	Span    Span
}

// String formats a diagnostic deterministically.
func (d Diagnostic) String() string {
	return d.Span.String() + ": " + string(d.Code) + ": " + d.Message
}

// Error implements error for convenient single-diagnostic reporting.
func (d Diagnostic) Error() string { return d.String() }

// Diagnostics is an ordered list of analyzer diagnostics.
type Diagnostics []Diagnostic

// SortBySpan returns a detached, deterministic source order. Filename is part
// of the ordering so package analysis does not depend on input file order.
func (d Diagnostics) SortBySpan() Diagnostics {
	result := append(Diagnostics(nil), d...)
	sort.SliceStable(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.Span.Filename != right.Span.Filename {
			return left.Span.Filename < right.Span.Filename
		}
		if left.Span.Start.Offset != right.Span.Start.Offset {
			return left.Span.Start.Offset < right.Span.Start.Offset
		}
		if left.Code != right.Code {
			return left.Code < right.Code
		}
		return left.Message < right.Message
	})
	return result
}

// HasErrors reports whether at least one diagnostic was emitted.
func (d Diagnostics) HasErrors() bool { return len(d) > 0 }

// Error returns all diagnostics as one deterministic error value.
func (d Diagnostics) Error() error {
	if !d.HasErrors() {
		return nil
	}
	lines := make([]string, 0, len(d))
	for _, diagnostic := range d.SortBySpan() {
		lines = append(lines, diagnostic.String())
	}
	return diagnosticError(strings.Join(lines, "\n"))
}

type diagnosticError string

func (e diagnosticError) Error() string { return string(e) }

// ObservationOrigin distinguishes contract-shaped signature facts from
// implementation observations in a generated Go projection.
type ObservationOrigin string

const (
	OriginSignature      ObservationOrigin = "signature"
	OriginImplementation ObservationOrigin = "implementation"
)

// Position is a zero-based byte offset plus one-based line and column, using
// token.FileSet's standard Go position conventions.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Span is an inclusive-start, exclusive-end source range.
type Span struct {
	Filename string
	Start    Position
	End      Position
}

// String formats a source span for diagnostics without requiring source text.
func (s Span) String() string {
	location := fmt.Sprintf("%d:%d", s.Start.Line, s.Start.Column)
	if s.Filename != "" {
		location = s.Filename + ":" + location
	}
	if s.Start.Offset == s.End.Offset {
		return location
	}
	return fmt.Sprintf("%s-%d:%d", location, s.End.Line, s.End.Column)
}

// Fact is a deterministic semantic relation produced from one source reference.
type Fact struct {
	Subject  Identity
	Relation Relation
	Object   Identity
	Span     Span
	Origin   ObservationOrigin
}

// Candidate is a potentially semantic relation that could not be selected
// deterministically. Options are sorted by stable identity.
type Candidate struct {
	Subject   Identity
	Relation  Relation
	Reference string
	Options   []Identity
	Span      Span
	Reason    string
	Origin    ObservationOrigin
}

// IdentityState explains why an implementation observation stayed deferred.
// These states are never semantic fact statuses and cannot enter candidates.
type IdentityState string

const (
	IdentityUnresolved IdentityState = "unresolved"
	IdentityAmbiguous  IdentityState = "ambiguous"
	IdentityInvalid    IdentityState = "invalid"
)

func (s IdentityState) valid() bool {
	return s == IdentityUnresolved || s == IdentityAmbiguous || s == IdentityInvalid
}

// ImplementationDetail records a source observation that stayed in the Go
// view because it has no usable registered semantic identity.
type ImplementationDetail struct {
	Reference     string        `json:"reference"`
	Span          Span          `json:"span"`
	Reason        string        `json:"reason"`
	IdentityState IdentityState `json:"identity_state"`
}

func (d ImplementationDetail) normalized() ImplementationDetail {
	if d.IdentityState == "" {
		d.IdentityState = IdentityUnresolved
	}
	return d
}

// SemanticDelta is the output of one analysis. Added contains only
// source-backed deterministic facts; ambiguity and implementation details are
// kept in separate collections.
type SemanticDelta struct {
	Added                 []Fact
	Candidates            []Candidate
	ImplementationDetails []ImplementationDetail
}

// DeterministicFacts returns a copy of the facts in the delta.
func (d SemanticDelta) DeterministicFacts() []Fact {
	return append([]Fact(nil), d.Added...)
}

// Result contains the semantic delta and source-local registrations discovered
// from annotations.
type Result struct {
	Delta         SemanticDelta
	Registrations []Registration
	Diagnostics   Diagnostics
}
