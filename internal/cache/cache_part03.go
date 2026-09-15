package cache

import (
	"fmt"
	"path/filepath"
)

// Open creates or opens a cache. The variadic form permits Open(path) and
// Open(path, Options{...}) without a second constructor.
func Open(root string, options ...Options) (*Cache, error) {
	configured, err := parseOptions(root, options)
	if err != nil {
		return nil, err
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("cache: resolve root: %w", err)
	}
	if err := ensureDirectory(absoluteRoot, 0o700); err != nil {
		return nil, fmt.Errorf("cache: prepare root: %w", err)
	}
	objects := filepath.Join(absoluteRoot, objectsDirName)
	if err := ensureDirectory(objects, 0o700); err != nil {
		return nil, fmt.Errorf("cache: prepare objects directory: %w", err)
	}
	if err := cleanupTemporaryEntries(objects); err != nil {
		return nil, fmt.Errorf("cache: clean temporary entries: %w", err)
	}
	return &Cache{root: absoluteRoot, objects: objects, maxEntrySize: configured.MaxEntrySize,
		locks: make(map[string]*entryLock), receipts: filepath.Join(absoluteRoot, receiptsFileName)}, nil
}
func parseOptions(root string, options []Options) (Options, error) {
	if len(options) > 1 {
		return Options{}, fmt.Errorf("cache: expected at most one Options value")
	}
	if root == "" {
		return Options{}, fmt.Errorf("cache: root must not be empty")
	}
	var configured Options
	if len(options) == 1 {
		configured = options[0]
	}
	if configured.MaxEntrySize < 0 {
		return Options{}, fmt.Errorf("cache: MaxEntrySize must not be negative")
	}
	return configured, nil
}

// Root returns the absolute cache root selected at construction time.
func (c *Cache) Root() string { return c.root }

// Get returns a verified projection and its metadata.
func (c *Cache) Get(key Key) ([]byte, Metadata, error) {
	release, err := c.acquireKey(key)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer release()
	return c.getLocked(key)
}

// GetMetadata returns verified metadata without exposing projection bytes.
func (c *Cache) GetMetadata(key Key) (Metadata, error) {
	_, metadata, err := c.Get(key)
	return metadata, err
}
