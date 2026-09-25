package semanticdelta

import (
	"fmt"
	"strings"
)

// Normalized returns a validated, sorted, detached request.
func (r Request) Normalized() (Request, error) {
	version := strings.TrimSpace(r.Version)
	if version == "" {
		version = FormatVersion
	}
	if version != FormatVersion {
		return Request{}, fmt.Errorf("unsupported semanticdelta version %q", version)
	}
	scope, err := r.Allowed.Normalized()
	if err != nil {
		return Request{}, err
	}
	delta, err := r.Delta.Normalized()
	if err != nil {
		return Request{}, err
	}
	return Request{Version: version, Allowed: scope, Delta: delta}, nil
}

// Normalize validates and canonicalizes the request in place.
func (r *Request) Normalize() error {
	if r == nil {
		return fmt.Errorf("cannot normalize a nil request")
	}
	normalized, err := r.Normalized()
	if err != nil {
		return err
	}
	*r = normalized
	return nil
}

// Normalized returns a validated, sorted, detached scope.
func (s Scope) Normalized() (Scope, error) {
	ids, err := normalizeValues("scope ID", s.IDs)
	if err != nil {
		return Scope{}, err
	}
	prefixes, err := normalizeValues("scope prefix", s.Prefixes)
	if err != nil {
		return Scope{}, err
	}
	predicates, err := normalizeValues("scope predicate", s.Predicates)
	if err != nil {
		return Scope{}, err
	}
	return Scope{IDs: ids, Prefixes: prefixes, Predicates: predicates}, nil
}

// Normalize validates and canonicalizes the scope in place.
func (s *Scope) Normalize() error {
	if s == nil {
		return fmt.Errorf("cannot normalize a nil scope")
	}
	normalized, err := s.Normalized()
	if err != nil {
		return err
	}
	*s = normalized
	return nil
}
