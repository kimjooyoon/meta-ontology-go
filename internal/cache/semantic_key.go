package cache

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// SemanticDigest returns the content identity of normalized semantic meaning.
// Presentation fields such as names, aliases, and source spans are excluded by
// semantic.IR's stable canonical form; stable IDs, namespaces, and relations
// remain part of the digest.
func SemanticDigest(ir semantic.IR) (Digest, error) {
	normalized, err := normalizeSemanticIR(ir)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidSemanticIdentity, err)
	}
	digest := Digest(normalized.StableHash())
	if !digest.Known() {
		return "", fmt.Errorf("%w: semantic digest is unknown", ErrInvalidSemanticIdentity)
	}
	return digest, nil
}

// NewSemanticProjectionKey derives the semantic-closure component of a
// projection key from the normalized IR. This keeps callers from accidentally
// addressing a semantic projection with a source or presentation digest.
// Domain and Namespace are inferred from the IR when omitted and, when
// supplied, must identify the same semantic namespace.
func NewSemanticProjectionKey(ir semantic.IR, spec ProjectionKeySpec) (ProjectionKey, error) {
	normalized, err := normalizeSemanticIR(ir)
	if err != nil {
		return ProjectionKey{}, fmt.Errorf("%w: %w", ErrInvalidSemanticIdentity, err)
	}
	digest := Digest(normalized.StableHash())
	if !digest.Known() {
		return ProjectionKey{}, fmt.Errorf("%w: semantic digest is unknown", ErrInvalidSemanticIdentity)
	}
	if spec.SemanticClosureDigest != "" && spec.SemanticClosureDigest != digest {
		return ProjectionKey{}, fmt.Errorf("%w: semantic closure digest does not match IR", ErrInvalidKey)
	}
	namespace := normalized.Namespace.String()
	if namespace != "" {
		if spec.Domain != "" && spec.Domain != namespace {
			return ProjectionKey{}, fmt.Errorf("%w: key domain does not match IR namespace", ErrInvalidKey)
		}
		if spec.Namespace != "" && spec.Namespace != namespace {
			return ProjectionKey{}, fmt.Errorf("%w: key namespace does not match IR namespace", ErrInvalidKey)
		}
		spec.Domain = namespace
		spec.Namespace = namespace
	}
	spec.SemanticClosureDigest = digest
	return NewProjectionKey(spec)
}

func normalizeSemanticIR(ir semantic.IR) (semantic.IR, error) {
	return ir.Normalized()
}
