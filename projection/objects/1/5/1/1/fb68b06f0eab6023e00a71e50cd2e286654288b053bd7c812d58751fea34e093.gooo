package semantic

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// ValidateAgainst ensures evidence supports a fact already present in the IR.
func (e Evidence) ValidateAgainst(graph Graph) error {
	normalized, err := e.Normalized()
	if err != nil {
		return err
	}
	if normalized.Status == FactCandidate {
		if !graph.HasCandidate(normalized.Fact) && !graph.HasFact(normalized.Fact) {
			return fmt.Errorf("%w: candidate fact is not present", ErrInvalidEvidence)
		}

		return nil
	}
	if !graph.HasFact(normalized.Fact) {
		return fmt.Errorf("%w: deterministic fact is not present", ErrInvalidEvidence)
	}
	return nil
}

// ValidateFresh checks that the evidence digest still names the pinned
// payload used to produce the record. It does not decide whether the claim is
// authoritative; that policy remains with the independent Go verifier.
func (e Evidence) ValidateFresh(payload []byte) error {
	normalized, err := e.Normalized()
	if err != nil {
		return err
	}
	expected := StableHash(payload)
	if normalized.Digest != expected {
		return fmt.Errorf("%w: got %s, want %s", ErrStaleEvidence, normalized.Digest, expected)
	}
	return nil
}
func normalizeDigest(raw string) (string, error) {
	digest := strings.ToLower(strings.TrimSpace(raw))
	if len(digest) != sha256.Size*2 {
		return "", fmt.Errorf("%w: digest must be a SHA-256 hex value", ErrInvalidEvidence)
	}
	if _, err := hex.DecodeString(digest); err != nil {
		return "", fmt.Errorf("%w: digest: %v", ErrInvalidEvidence, err)
	}
	return digest, nil
}
func (e Evidence) Canonical() string {
	if normalized, err := e.Normalized(); err == nil {
		e = normalized
	}
	var b strings.Builder
	b.WriteString("evidence\t")
	writeCanonicalField(&b, e.ID.String())
	writeCanonicalField(&b, e.Producer.String())
	writeCanonicalField(&b, e.Kind.String())
	writeCanonicalField(&b, e.Status.String())
	writeCanonicalField(&b, e.Fact.Subject.String())
	writeCanonicalField(&b, e.Fact.Predicate.String())
	writeCanonicalField(&b, e.Fact.Object.String())
	writeCanonicalField(&b, e.Digest)
	writeCanonicalSpan(&b, e.Span)
	return b.String()
}
