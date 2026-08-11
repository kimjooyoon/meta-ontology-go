// Package semantic contains the parser-independent semantic kernel for gooo.
//
// The package deliberately keeps semantic identity separate from display names
// and namespaces. A name is a lookup convenience; an ID is the identity of a
// semantic node and must be supplied explicitly by an authoritative view.
package semantic

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

var (
	ErrInvalidIdentity  = errors.New("invalid semantic identity")
	ErrInvalidNamespace = errors.New("invalid semantic namespace")
)

// ID is an absolute, stable, URI-like semantic identity.
//
// ID intentionally does not expose a namespace derived from the URI. A
// namespace is an independent semantic boundary and must be carried by the
// declaration that owns the ID.
type ID string

// Identity is an alias kept for callers that prefer the vocabulary used in the
// language specification.
type Identity = ID

// ParseIdentity validates and canonicalizes a URI-like identity. URI scheme
// and host casing are normalized because those components are case-insensitive
// by URI convention; the rest of the identifier remains unchanged.
func ParseIdentity(raw string) (ID, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", fmt.Errorf("%w: empty value", ErrInvalidIdentity)
	}
	if strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", fmt.Errorf("%w: whitespace is not allowed", ErrInvalidIdentity)
	}

	u, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidIdentity, err)
	}
	if u.Scheme == "" {
		return "", fmt.Errorf("%w: identity must have a URI scheme", ErrInvalidIdentity)
	}
	// A value using the authority delimiter must have an authority. This
	// catches accidental values such as "billing://" while still allowing
	// opaque URI forms such as "urn:gooo:entity:order".
	if strings.Contains(value, "://") && u.Host == "" {
		return "", fmt.Errorf("%w: URI authority is empty", ErrInvalidIdentity)
	}

	u.Scheme = strings.ToLower(u.Scheme)
	if u.Host != "" {
		u.Host = strings.ToLower(u.Host)
	}
	canonical := u.String()
	if canonical == "" {
		return "", fmt.Errorf("%w: empty canonical value", ErrInvalidIdentity)
	}
	return ID(canonical), nil
}

// NewIdentity is a descriptive alias for ParseIdentity.
func NewIdentity(raw string) (ID, error) {
	return ParseIdentity(raw)
}

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

// NameRef is a fully qualified display-name lookup key. It is never used as
// semantic identity and therefore cannot accidentally merge equal names from
// different namespaces.
type NameRef struct {
	Namespace Namespace
	Name      string
}

func NewNameRef(namespace Namespace, name string) (NameRef, error) {
	ns, err := ParseNamespace(namespace.String())
	if err != nil {
		return NameRef{}, err
	}
	canonicalName, err := normalizeName(name)
	if err != nil {
		return NameRef{}, err
	}
	return NameRef{Namespace: ns, Name: canonicalName}, nil
}
