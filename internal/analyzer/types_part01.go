package analyzer

import (
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
