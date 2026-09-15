package cache

import (
	"errors"
)

var (
	// ErrNotFound identifies a cache miss. ErrMiss is an equivalent alias.
	ErrNotFound = errors.New("cache miss")
	ErrMiss     = ErrNotFound

	// ErrCorrupt means an entry exists but failed integrity validation.
	ErrCorrupt = errors.New("cache entry corrupt")

	ErrInvalidKey              = errors.New("invalid cache key")
	ErrInvalidSemanticIdentity = errors.New("invalid semantic identity")
	ErrInvalidFreshness        = errors.New("invalid cache freshness")
	ErrUnknownFreshness        = errors.New("cache freshness is unknown")
	ErrStale                   = errors.New("cache entry stale")
	ErrEmptyFilter             = errors.New("cache invalidation filter is empty")
	ErrNotReconstructable      = errors.New("cache entries must be reconstructable projections")
	ErrEntryTooLarge           = errors.New("cache entry exceeds configured size limit")
)

const (
	metadataVersion  = "v1"
	objectsDirName   = "objects"
	dataFileName     = "data"
	metaFileName     = "metadata.json"
	receiptsFileName = "receipts.jsonl"
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
