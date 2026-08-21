package semantic

import (
	"fmt"
	"strings"
	"unicode"
)

// MustIdentity is intended for package-level declarations and tests. It
// panics only when the supplied identity is invalid.
func MustIdentity(raw string) ID {
	id, err := ParseIdentity(raw)
	if err != nil {
		panic(err)
	}
	return id
}

// String returns the canonical identity text.
func (id ID) String() string {
	return string(id)
}

// Valid reports whether id is a canonicalizable absolute identity.
func (id ID) Valid() bool {
	_, err := ParseIdentity(string(id))
	return err == nil
}

// Canonical returns the normalized identity. It returns an empty string for an
// invalid ID; callers that need diagnostics should use ParseIdentity.
func (id ID) Canonical() string {
	canonical, err := ParseIdentity(string(id))
	if err != nil {
		return ""
	}
	return canonical.String()
}

// Namespace is an explicit semantic scope. It is intentionally not inferred
// from an ID: two namespaces may use the same display name and may even use
// unrelated identity URI schemes.
type Namespace string

// ParseNamespace validates and trims a namespace label. Namespace labels are
// case-sensitive and may contain punctuation, but never whitespace.
func ParseNamespace(raw string) (Namespace, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("%w: empty value", ErrInvalidNamespace)
	}
	if strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", fmt.Errorf("%w: whitespace is not allowed", ErrInvalidNamespace)
	}
	return Namespace(value), nil
}

// NewNamespace is a descriptive alias for ParseNamespace.
func NewNamespace(raw string) (Namespace, error) {
	return ParseNamespace(raw)
}
func (ns Namespace) String() string {
	return string(ns)
}
func (ns Namespace) Valid() bool {
	_, err := ParseNamespace(string(ns))
	return err == nil
}
