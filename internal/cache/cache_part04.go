package cache

import (
	"context"
	"errors"
	"fmt"
)

// Has reports whether key is a valid cache hit. Corrupt entries are misses.
func (c *Cache) Has(key Key) (bool, error) {
	_, _, err := c.Get(key)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, ErrNotFound) || errors.Is(err, ErrCorrupt) {
		return false, nil
	}
	return false, err
}

// Put stores data as a reconstructable projection using default entry info.
func (c *Cache) Put(key Key, data []byte) error {
	return c.PutWithInfo(key, data, EntryInfo{})
}

// PutWithInfo stores data and projection descriptors. The complete object
// directory becomes visible with one rename after both files are synced.
func (c *Cache) PutWithInfo(key Key, data []byte, info EntryInfo) error {
	release, err := c.acquireKey(key)
	if err != nil {
		return err
	}
	defer release()
	return c.putLocked(key, data, info)
}

// ComputeFunc constructs a projection after a cache miss.
type ComputeFunc func() ([]byte, error)

// GetOrCompute returns a hit when key is present, otherwise computes and
// stores it. Same-key calls on one Cache instance are serialized.
func (c *Cache) GetOrCompute(ctx context.Context, key Key, compute ComputeFunc) ([]byte, Metadata, bool, error) {
	if compute == nil {
		return nil, Metadata{}, false, fmt.Errorf("cache: nil compute function")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, Metadata{}, false, err
	}
	release, err := c.acquireKey(key)
	if err != nil {
		return nil, Metadata{}, false, err
	}
	defer release()
	return c.computeLocked(ctx, key, compute)
}
