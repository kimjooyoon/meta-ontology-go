package cache

import "fmt"

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

// GetProjectionIfFresh returns bytes only when the projection identity and a
// sealed, ledger-recorded evidence receipt both match the requested tuple.
// Rejections are read-only and never repair, rewrite, or append cache state.
func (c *Cache) GetProjectionIfFresh(key Key, identity ProjectionIdentity,
	current EvidenceFreshness, receipt CacheReceipt) ([]byte, Metadata, error) {
	if err := identity.Validate(); err != nil {
		return nil, Metadata{}, err
	}
	if err := validateFullKey(key); err != nil {
		return nil, Metadata{}, err
	}
	if !identity.matchesKey(key) {
		return nil, Metadata{}, fmt.Errorf("%w: projection identity differs from key", ErrStale)
	}
	if err := current.Validate(); err != nil {
		return nil, Metadata{}, err
	}
	if !identity.matchesEvidence(current) {
		return nil, Metadata{}, fmt.Errorf("%w: source, IR, or toolchain identity differs", ErrStale)
	}
	if err := receipt.ValidateForKey(key); err != nil {
		return nil, Metadata{}, err
	}
	if !receipt.hasArtifact() || !receipt.Evidence.Matches(current) {
		return nil, Metadata{}, fmt.Errorf("%w: evidence receipt is stale", ErrStale)
	}
	sealed, err := receipt.Seal()
	if err != nil || sealed.ReceiptDigest != receipt.ReceiptDigest {
		return nil, Metadata{}, fmt.Errorf("%w: receipt is not sealed", ErrInvalidReceipt)
	}
	recorded, err := c.receiptRecorded(receipt.ReceiptDigest)
	if err != nil {
		return nil, Metadata{}, err
	}
	if !recorded {
		return nil, Metadata{}, fmt.Errorf("%w: receipt is not recorded", ErrInvalidReceipt)
	}
	data, metadata, err := c.Get(key)
	if err != nil {
		return nil, Metadata{}, err
	}
	if err := receipt.ValidateForData(data); err != nil {
		return nil, Metadata{}, err
	}
	return data, metadata, nil
}

func (c *Cache) receiptRecorded(digest Digest) (bool, error) {
	receipts, err := c.Receipts()
	if err != nil {
		return false, err
	}
	for _, receipt := range receipts {
		if receipt.ReceiptDigest == digest {
			return true, nil
		}
	}
	return false, nil
}
