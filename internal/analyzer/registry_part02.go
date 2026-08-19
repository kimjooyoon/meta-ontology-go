package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

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
	if _, err := semantic.ParseIdentity(registration.Identity.ID); err != nil {
		return invalidRegistrationError("semantic identity must be a valid URI")
	}
	if strings.TrimSpace(registration.Identity.Namespace) != "" {
		if _, err := semantic.ParseNamespace(registration.Identity.Namespace); err != nil {
			return invalidRegistrationError("semantic namespace must be valid")
		}
	}
	return nil
}
func validateRegistryRegistration(registration Registration) error {
	if err := validateRegistration(registration); err != nil {
		return err
	}
	if _, err := semantic.ParseIdentity(registration.Identity.ID); err != nil {
		return invalidRegistrationError("semantic identity must be a valid URI")
	}
	if _, err := semantic.ParseNamespace(registration.Identity.Namespace); err != nil {
		return invalidRegistrationError("semantic namespace must be valid")
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
