package semantic

import (
	"fmt"
)

func (e Evidence) Normalized() (Evidence, error) {
	id, err := ParseIdentity(e.ID.String())
	if err != nil {
		return Evidence{}, fmt.Errorf("%w: id: %v", ErrInvalidEvidence, err)
	}
	producer, err := ParseIdentity(e.Producer.String())
	if err != nil {
		return Evidence{}, fmt.Errorf("%w: producer: %v", ErrInvalidEvidence, err)
	}
	fact, err := normalizeFactKey(e.Fact)
	if err != nil {
		return Evidence{}, fmt.Errorf("%w: fact: %v", ErrInvalidEvidence, err)
	}
	if !e.Kind.Valid() {
		return Evidence{}, fmt.Errorf("%w: unknown kind %q", ErrInvalidEvidence, e.Kind)
	}
	status := e.Status
	if status == 0 {
		status = FactDeterministic
	}
	if status != FactDeterministic && status != FactCandidate {
		return Evidence{}, fmt.Errorf("%w: unknown status %d", ErrInvalidEvidence, status)
	}
	if status == FactCandidate && e.Kind != CompilerRunEvidence {
		return Evidence{}, fmt.Errorf("%w: candidate evidence must be %q, got %q", ErrInvalidEvidence, CompilerRunEvidence, e.Kind)
	}
	digest, err := normalizeDigest(e.Digest)
	if err != nil {
		return Evidence{}, err
	}
	span := e.Span.Normalized()
	if err := span.Validate(); err != nil {
		return Evidence{}, fmt.Errorf("%w: span: %v", ErrInvalidEvidence, err)
	}
	e.ID = id
	e.Producer = producer
	e.Fact = fact
	e.Status = status
	e.Digest = digest
	e.Span = span
	return e, nil
}
func (e Evidence) Validate() error {
	_, err := e.Normalized()
	return err
}
func (e Evidence) WithSpan(span Span) Evidence {
	e.Span = span
	return e
}
