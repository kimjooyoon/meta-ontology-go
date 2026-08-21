package cache

import (
	"fmt"
)

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
