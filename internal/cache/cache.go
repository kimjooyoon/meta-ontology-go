package cache

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"
)

var (
	// ErrNotFound identifies a cache miss. ErrMiss is an equivalent alias.
	ErrNotFound = errors.New("cache miss")
	ErrMiss     = ErrNotFound

	// ErrCorrupt means an entry exists but failed integrity validation.
	ErrCorrupt = errors.New("cache entry corrupt")

	ErrInvalidKey         = errors.New("invalid cache key")
	ErrEmptyFilter        = errors.New("cache invalidation filter is empty")
	ErrNotReconstructable = errors.New("cache entries must be reconstructable projections")
	ErrEntryTooLarge      = errors.New("cache entry exceeds configured size limit")
)

const (
	metadataVersion = "v1"
	objectsDirName  = "objects"
	dataFileName    = "data"
	metaFileName    = "metadata.json"
)

// Options configures a cache opened on a filesystem directory.
type Options struct {
	// MaxEntrySize rejects writes and reads larger than this number of bytes.
	// Zero means no explicit limit.
	MaxEntrySize int64
}

// EntryInfo describes the reconstructable projection stored in an entry.
type EntryInfo struct {
	ArtifactType string
	Projection   string
}

// Metadata is the integrity and provenance envelope for a cached projection.
// CreatedAt is observational metadata and is not part of the content address.
type Metadata struct {
	FormatVersion   string    `json:"format_version"`
	Key             string    `json:"key"`
	KeyVersion      string    `json:"key_version"`
	Namespace       string    `json:"namespace"`
	ToolVersion     string    `json:"tool_version"`
	HostStage       HostStage `json:"host_stage"`
	InputDigest     Digest    `json:"input_digest"`
	OptionsDigest   Digest    `json:"options_digest"`
	ArtifactType    string    `json:"artifact_type,omitempty"`
	Projection      string    `json:"projection,omitempty"`
	Reconstructable bool      `json:"reconstructable"`
	Size            int64     `json:"size"`
	ContentDigest   Digest    `json:"content_digest"`
	MetadataDigest  Digest    `json:"metadata_digest"`
	CreatedAt       time.Time `json:"created_at"`
}

// InvalidationFilter selects cached projections to remove. Empty fields are
// wildcards, but at least one field must be set. Use Clear to remove all.
type InvalidationFilter struct {
	Namespace    string
	KeyVersion   string
	ToolVersion  string
	ArtifactType string
	Projection   string
}

// Cache is a content-addressed, filesystem-backed projection cache.
type Cache struct {
	root         string
	objects      string
	maxEntrySize int64
	filesystemMu sync.RWMutex
	locksMu      sync.Mutex
	locks        map[string]*entryLock
}

type entryLock struct {
	mu   sync.Mutex
	refs int
}

// New creates or opens a cache rooted at root.
func New(root string) (*Cache, error) { return Open(root) }

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
		locks: make(map[string]*entryLock)}, nil
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
