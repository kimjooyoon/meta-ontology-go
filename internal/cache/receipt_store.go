package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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

func (c *Cache) readReceiptsLocked() ([]CacheReceipt, error) {
	file, err := openReceiptRead(c.receipts)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cache: open receipts: %w", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	result := make([]CacheReceipt, 0)
	seenRuns := make(map[string]struct{})
	seenBundles := make(map[Digest]struct{})
	seenEventRefs := make(map[evidenceAttempt]struct{})
	seenEventIDs := make(map[evidenceAttempt]struct{})
	for {
		var receipt CacheReceipt
		err := decoder.Decode(&receipt)
		if errors.Is(err, io.EOF) {
			return result, nil
		}
		if err != nil {
			return nil, fmt.Errorf("cache: decode receipt: %w", err)
		}
		if err := validateSealedReceipt(receipt); err != nil {
			return nil, err
		}
		if _, exists := seenRuns[receipt.Evidence.RunID]; exists {
			return nil, ErrReceiptReplay
		}
		if _, exists := seenBundles[receipt.Evidence.BundleDigest]; exists {
			return nil, ErrReceiptReplay
		}
		eventRef := evidenceAttempt{ID: receipt.Evidence.EventRef, Attempt: receipt.Evidence.Attempt}
		eventID := evidenceAttempt{ID: receipt.Evidence.EventID, Attempt: receipt.Evidence.Attempt}
		if _, exists := seenEventRefs[eventRef]; exists {
			return nil, ErrReceiptReplay
		}
		if _, exists := seenEventIDs[eventID]; exists {
			return nil, ErrReceiptReplay
		}
		seenRuns[receipt.Evidence.RunID] = struct{}{}
		seenBundles[receipt.Evidence.BundleDigest] = struct{}{}
		seenEventRefs[eventRef] = struct{}{}
		seenEventIDs[eventID] = struct{}{}
		result = append(result, receipt)
	}
}

type evidenceAttempt struct {
	ID      string
	Attempt uint64
}

func validateSealedReceipt(receipt CacheReceipt) error {
	if err := receipt.Validate(); err != nil {
		return err
	}
	if !receipt.ReceiptDigest.Known() {
		return fmt.Errorf("%w: missing receipt digest", ErrInvalidReceipt)
	}
	copy := receipt
	copy.Evidence = canonicalEvidence(copy.Evidence)
	copy.EvidenceRefs = append([]EvidenceRef(nil), copy.Evidence.EvidenceRefs...)
	copy.ReceiptDigest = ""
	digest, err := DigestOf(copy)
	if err != nil || digest != receipt.ReceiptDigest {
		return fmt.Errorf("%w: receipt digest mismatch", ErrInvalidReceipt)
	}
	return nil
}

func validateReceiptFile(file *os.File) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return ErrUnsafeReceiptLog
	}
	return nil
}
