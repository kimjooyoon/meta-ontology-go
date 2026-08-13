package cache

import (
	"context"
	"errors"
	"fmt"
)

func (c *Cache) computeLocked(ctx context.Context, key Key, compute ComputeFunc) ([]byte, Metadata, bool, error) {
	data, metadata, err := c.getLocked(key)
	if err == nil {
		return data, metadata, true, nil
	}
	if !errors.Is(err, ErrNotFound) && !errors.Is(err, ErrCorrupt) {
		return nil, Metadata{}, false, err
	}
	if err := ctx.Err(); err != nil {
		return nil, Metadata{}, false, err
	}
	data, err = compute()
	if err != nil {
		return nil, Metadata{}, false, fmt.Errorf("cache: compute %s: %w", key, err)
	}
	if err := ctx.Err(); err != nil {
		return nil, Metadata{}, false, err
	}
	if err := c.putLocked(key, data, EntryInfo{}); err != nil {
		return nil, Metadata{}, false, err
	}
	data, metadata, err = c.getLocked(key)
	if err != nil {
		return nil, Metadata{}, false, err
	}
	return data, metadata, false, nil
}
