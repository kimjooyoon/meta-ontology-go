package cache

import (
	"encoding/json"
	"fmt"
)

// AppendReceipt durably appends one sealed receipt and rejects run/bundle
// replay. There is intentionally no update or delete operation.
func (c *Cache) AppendReceipt(receipt CacheReceipt) (CacheReceipt, error) {
	sealed, err := receipt.Seal()
	if err != nil {
		return CacheReceipt{}, err
	}
	c.receiptMu.Lock()
	defer c.receiptMu.Unlock()
	release, err := acquireReceiptFileLock(c.receipts)
	if err != nil {
		return CacheReceipt{}, fmt.Errorf("cache: lock receipts: %w", err)
	}
	defer release()
	existing, err := c.readReceiptsLocked()
	if err != nil {
		return CacheReceipt{}, err
	}
	for _, prior := range existing {
		if prior.ReceiptDigest == sealed.ReceiptDigest || prior.Evidence.RunID == sealed.Evidence.RunID ||
			prior.Evidence.BundleDigest == sealed.Evidence.BundleDigest ||
			prior.Evidence.EventRef == sealed.Evidence.EventRef && prior.Evidence.Attempt == sealed.Evidence.Attempt ||
			prior.Evidence.EventID == sealed.Evidence.EventID && prior.Evidence.Attempt == sealed.Evidence.Attempt {
			return CacheReceipt{}, ErrReceiptReplay
		}
	}
	data, err := json.Marshal(sealed)
	if err != nil {
		return CacheReceipt{}, fmt.Errorf("cache: encode receipt: %w", err)
	}
	file, err := openReceiptAppend(c.receipts)
	if err != nil {
		return CacheReceipt{}, fmt.Errorf("cache: open receipts: %w", err)
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		_ = file.Close()
		return CacheReceipt{}, fmt.Errorf("cache: append receipt: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return CacheReceipt{}, fmt.Errorf("cache: sync receipt: %w", err)
	}
	if err := file.Close(); err != nil {
		return CacheReceipt{}, fmt.Errorf("cache: close receipt: %w", err)
	}
	return sealed, nil
}

// Receipts reads and validates the append-only receipt log.
func (c *Cache) Receipts() ([]CacheReceipt, error) {
	c.receiptMu.Lock()
	defer c.receiptMu.Unlock()
	release, err := acquireReceiptFileLock(c.receipts)
	if err != nil {
		return nil, fmt.Errorf("cache: lock receipts: %w", err)
	}
	defer release()
	return c.readReceiptsLocked()
}
