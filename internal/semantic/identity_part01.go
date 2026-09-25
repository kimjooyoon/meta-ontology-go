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
