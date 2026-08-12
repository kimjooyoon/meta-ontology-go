package analyzer

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const registrySchema = "analyzer-registry/v1"

// Registry contains semantic symbols allowed to cross the Go / semantic
// boundary. Callers provide the stable symbol keys; the registry does not load
// packages or infer types.
type Registry struct {
	entries []Registration
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register adds a semantic symbol. Multiple identities for one Go reference
// are intentionally retained as ambiguity during analysis.
func (r *Registry) Register(registration Registration) error {
	if r == nil {
		return errNilRegistry
	}
	if err := validateRegistration(registration); err != nil {
		return err
	}
	for _, existing := range r.entries {
		if sameRegistration(existing, registration) {
			return nil
		}
	}
	r.entries = append(r.entries, registration)
	return nil
}

// MustRegister is Register for package initialization and tests.
func (r *Registry) MustRegister(registration Registration) {
	if err := r.Register(registration); err != nil {
		panic(err)
	}
}

// RegisterActivity registers an Activity reference.
func (r *Registry) RegisterActivity(ref SymbolRef, identity Identity) error {
	return r.Register(Registration{Ref: ref, Kind: KindActivity, Identity: identity})
}

// RegisterEntity registers an Entity reference.
func (r *Registry) RegisterEntity(ref SymbolRef, identity Identity) error {
	return r.Register(Registration{Ref: ref, Kind: KindEntity, Identity: identity})
}

func (r *Registry) all() []Registration {
	if r == nil {
		return nil
	}
	entries := append([]Registration(nil), r.entries...)
	sortRegistrations(entries)
	return entries
}

// Canonical returns the order-independent identity mapping used by the
// analyzer. Source spans are evidence, not registry identity.
func (r *Registry) Canonical() string {
	entries := r.all()
	var builder strings.Builder
	builder.WriteString(registrySchema)
	builder.WriteByte('\n')
	writeBindingField(&builder, intString(len(entries)))
	for _, entry := range entries {
		builder.WriteString("entry\n")
		writeBindingField(&builder, entry.Ref.PackagePath)
		writeBindingField(&builder, entry.Ref.PackageName)
		writeBindingField(&builder, entry.Ref.Receiver)
		writeBindingField(&builder, entry.Ref.Name)
		writeBindingField(&builder, string(entry.Kind))
		writeBindingField(&builder, entry.Identity.Namespace)
		writeBindingField(&builder, entry.Identity.ID)
	}
	return builder.String()
}

// Digest returns the versioned identity of the complete registry mapping.
// Nil and empty registries intentionally have the same deterministic digest.
func (r *Registry) Digest() string { return semantic.StableHashString(r.Canonical()) }

func validateRegistration(registration Registration) error {
	if !validKind(registration.Kind) {
		return invalidRegistrationError("kind must be activity or entity")
	}
	if strings.TrimSpace(registration.Ref.Name) == "" {
		return invalidRegistrationError("symbol name is required")
	}
	if !registration.Identity.Valid() {
		return invalidRegistrationError("semantic identity is required")
	}
	return nil
}

func validKind(kind SymbolKind) bool {
	return kind == KindActivity || kind == KindEntity
}

func sameRegistration(left, right Registration) bool {
	return left.Ref == right.Ref && left.Kind == right.Kind && left.Identity == right.Identity
}

type invalidRegistrationError string

func (e invalidRegistrationError) Error() string { return string(e) }

var errNilRegistry = invalidRegistrationError("registry is nil")
