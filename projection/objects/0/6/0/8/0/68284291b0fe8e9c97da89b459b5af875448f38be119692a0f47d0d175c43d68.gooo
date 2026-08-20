package cache

import (
	"fmt"
)

// ProjectionIdentity is the typed identity needed before a projection hit can
// be returned. EvidenceFreshness remains separate and proves freshness only.
type ProjectionIdentity struct {
	SemanticClosureDigest Digest
	SourceDigest          Digest
	IRDigest              Digest
	OptionsDigest         Digest
	Toolchain             string
	ToolchainDigest       Digest
}

// Validate rejects identity values that cannot be compared deterministically.
func (i ProjectionIdentity) Validate() error {
	for _, field := range []struct {
		label string
		value Digest
	}{
		{"semantic closure", i.SemanticClosureDigest}, {"source", i.SourceDigest},
		{"IR", i.IRDigest}, {"options", i.OptionsDigest}, {"toolchain", i.ToolchainDigest},
	} {
		if !field.value.Known() {
			return fmt.Errorf("%w: unknown %s identity", ErrUnknownFreshness, field.label)
		}
	}
	if err := validateKeyComponent("toolchain", i.Toolchain, true); err != nil {
		return fmt.Errorf("%w: %v", ErrUnknownFreshness, err)
	}
	return nil
}
func (i ProjectionIdentity) matchesKey(key Key) bool {
	return i.SemanticClosureDigest == key.SemanticClosureDigest &&
		i.OptionsDigest == key.OptionsDigest && i.Toolchain == key.Toolchain
}
func (i ProjectionIdentity) matchesEvidence(evidence EvidenceFreshness) bool {
	return i.SourceDigest == evidence.SourceDigest && i.IRDigest == evidence.IRDigest &&
		i.ToolchainDigest == evidence.ToolchainDigest
}
