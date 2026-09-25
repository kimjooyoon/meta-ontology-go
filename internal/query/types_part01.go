package query

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

var (
	ErrInvalidID        = errors.New("invalid stable ID")
	ErrInvalidRelation  = errors.New("invalid PROV relation")
	ErrInvalidFact      = errors.New("invalid query fact")
	ErrInvalidQuery     = errors.New("invalid exact query")
	ErrInvalidTraversal = errors.New("invalid traversal options")
)

// StableID is a canonical, URI-like semantic identity. Display names are not
// accepted as substitutes because equal names may belong to different scopes.
type StableID string

// ID is the shorter spelling used by facts and query APIs.
type ID = StableID

// ParseID validates and canonicalizes a stable ID. URI scheme and host casing
// are normalized; path and opaque URI content retain their case.
func ParseID(raw string) (StableID, error) {
	value := strings.TrimSpace(raw)
	if value == "" || strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", fmt.Errorf("%w: empty or whitespace-containing value", ErrInvalidID)
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidID, err)
	}
	if parsed.Scheme == "" {
		return "", fmt.Errorf("%w: URI scheme is required", ErrInvalidID)
	}
	if strings.Contains(value, "://") && parsed.Host == "" {
		return "", fmt.Errorf("%w: URI authority is empty", ErrInvalidID)
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	canonical := parsed.String()
	if canonical == "" {
		return "", fmt.Errorf("%w: empty canonical value", ErrInvalidID)
	}
	return StableID(canonical), nil
}

// NewID is an alias for ParseID for callers constructing graph inputs.
func NewID(raw string) (ID, error) { return ParseID(raw) }
func (id StableID) String() string { return string(id) }

// Valid reports whether id is already a valid stable identity. It does not
// require callers to retain the canonical form when only checking input.
func (id StableID) Valid() bool {
	_, err := ParseID(id.String())
	return err == nil
}

// Relation is a directed PROV relation. Its subject and object follow the
// PROV-O direction, for example Activity prov:used Entity.
type Relation string
