package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

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
