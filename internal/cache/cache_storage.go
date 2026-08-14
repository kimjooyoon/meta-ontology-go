package cache

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func (c *Cache) getLocked(key Key) ([]byte, Metadata, error) {
	if err := c.validatePathKey(key); err != nil {
		return nil, Metadata{}, err
	}
	c.filesystemMu.RLock()
	defer c.filesystemMu.RUnlock()
	return c.readObject(key)
}

func (c *Cache) putLocked(key Key, data []byte, info EntryInfo) error {
	if err := validateFullKey(key); err != nil {
		return err
	}
	if c.maxEntrySize > 0 && int64(len(data)) > c.maxEntrySize {
		return fmt.Errorf("cache: %w: %d > %d", ErrEntryTooLarge, len(data), c.maxEntrySize)
	}
	if err := validateEntryInfo(info); err != nil {
		return err
	}
	if info.Projection != "" && info.Projection != key.Projection {
		return fmt.Errorf("%w: EntryInfo projection is not key identity", ErrInvalidKey)
	}
	path, err := c.objectPath(key)
	if err != nil {
		return err
	}
	c.filesystemMu.Lock()
	defer c.filesystemMu.Unlock()
	return c.writeObject(path, key, data, info)
}

func (c *Cache) writeObject(path string, key Key, data []byte, info EntryInfo) error {
	if _, _, err := c.readObject(key); err == nil {
		return nil
	} else if !errors.Is(err, ErrNotFound) && !errors.Is(err, ErrCorrupt) {
		return err
	}
	shardPath := filepath.Dir(path)
	if err := ensureDirectory(shardPath, 0o700); err != nil {
		return fmt.Errorf("cache: prepare shard: %w", err)
	}
	temporary, err := os.MkdirTemp(shardPath, "."+key.String()+".tmp-")
	if err != nil {
		return fmt.Errorf("cache: create staging directory: %w", err)
	}
	defer func() { _ = removeCacheEntryBestEffort(temporary) }()
	metadata := makeMetadata(key, data, info)
	if err := writeObjectFiles(temporary, data, metadata); err != nil {
		return err
	}
	bestEffortSyncDirectory(temporary)
	if err := commitObject(temporary, path, key, c); err != nil {
		return err
	}
	bestEffortSyncDirectory(shardPath)
	return nil
}

func makeMetadata(key Key, data []byte, info EntryInfo) Metadata {
	metadata := Metadata{
		FormatVersion:         metadataVersion,
		Key:                   key.String(),
		KeyVersion:            key.Version,
		Domain:                key.Domain,
		Namespace:             key.Namespace,
		ArtifactKind:          key.ArtifactKind,
		ToolVersion:           key.ToolVersion,
		Toolchain:             key.Toolchain,
		Target:                key.Target,
		HostStage:             key.HostStage,
		InputDigest:           key.InputDigest,
		SemanticClosureDigest: key.SemanticClosureDigest,
		DependencyRoot:        key.DependencyRoot,
		PolicySchemaDigest:    key.PolicySchemaDigest,
		BuildTagsDigest:       key.BuildTagsDigest,
		OptionsDigest:         key.OptionsDigest,
		DependencyDigest:      key.DependencyDigest,
		ProvenanceDigest:      key.ProvenanceDigest,
		ArtifactType:          info.ArtifactType,
		Projection:            key.Projection,
		Reconstructable:       true,
		Size:                  int64(len(data)),
		ContentDigest:         HashBytes(data),
		CreatedAt:             time.Now().UTC(),
	}
	metadata.MetadataDigest = digestMetadata(metadata)
	return metadata
}

func digestMetadata(metadata Metadata) Digest {
	metadata.MetadataDigest = ""
	data, err := CanonicalBytes(metadata)
	if err != nil {
		return ""
	}
	return HashBytes(data)
}

func writeObjectFiles(directory string, data []byte, metadata Metadata) error {
	if err := writeDurableFile(filepath.Join(directory, dataFileName), data); err != nil {
		return fmt.Errorf("cache: write projection: %w", err)
	}
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("cache: encode metadata: %w", err)
	}
	if err := writeDurableFile(filepath.Join(directory, metaFileName), metadataBytes); err != nil {
		return fmt.Errorf("cache: write metadata: %w", err)
	}
	return nil
}

func commitObject(temporary, path string, key Key, cache *Cache) error {
	if err := os.Rename(temporary, path); err == nil {
		return nil
	} else if _, _, existingErr := cache.readObject(key); existingErr == nil {
		return nil
	}
	if info, statErr := os.Lstat(path); statErr == nil {
		if removeErr := removeCacheEntry(path, info); removeErr != nil {
			return fmt.Errorf("cache: replace corrupt entry: %w", removeErr)
		}
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return fmt.Errorf("cache: inspect existing entry: %w", statErr)
	}
	if err := os.Rename(temporary, path); err != nil {
		if _, _, existingErr := cache.readObject(key); existingErr == nil {
			return nil
		}
		return fmt.Errorf("cache: commit projection: %w", err)
	}
	return nil
}

func (c *Cache) readObject(key Key) ([]byte, Metadata, error) {
	path, err := c.objectPath(key)
	if err != nil {
		return nil, Metadata{}, err
	}
	objectInfo, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, Metadata{}, fmt.Errorf("%w: %s", ErrNotFound, key)
	}
	if err != nil {
		return nil, Metadata{}, fmt.Errorf("cache: inspect %s: %w", key, err)
	}
	if objectInfo.Mode()&os.ModeSymlink != 0 || !objectInfo.IsDir() {
		return nil, Metadata{}, corruptError(key, "object is not a directory")
	}
	metadata, err := readMetadataAt(filepath.Join(path, metaFileName))
	if err != nil {
		return nil, Metadata{}, corruptError(key, "invalid metadata: "+err.Error())
	}
	if err := validateMetadataForKey(metadata, key); err != nil {
		return nil, Metadata{}, corruptError(key, err.Error())
	}
	data, err := readDataFile(filepath.Join(path, dataFileName), c.maxEntrySize)
	if err != nil {
		if errors.Is(err, ErrEntryTooLarge) {
			return nil, Metadata{}, err
		}
		return nil, Metadata{}, corruptError(key, "invalid projection: "+err.Error())
	}
	if int64(len(data)) != metadata.Size {
		return nil, Metadata{}, corruptError(key, fmt.Sprintf("size mismatch: metadata=%d actual=%d", metadata.Size, len(data)))
	}
	if HashBytes(data) != metadata.ContentDigest {
		return nil, Metadata{}, corruptError(key, "content digest mismatch")
	}
	return append([]byte(nil), data...), metadata, nil
}

func readMetadataAt(path string) (Metadata, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return Metadata{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return Metadata{}, fmt.Errorf("metadata is not a regular file")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Metadata{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var metadata Metadata
	if err := decoder.Decode(&metadata); err != nil {
		return Metadata{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Metadata{}, fmt.Errorf("multiple JSON values")
		}
		return Metadata{}, err
	}
	return metadata, nil
}

func readDataFile(path string, maxSize int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("projection is not a regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if maxSize > 0 {
		data, err := io.ReadAll(io.LimitReader(file, maxSize+1))
		if err != nil {
			return nil, err
		}
		if int64(len(data)) > maxSize {
			return nil, ErrEntryTooLarge
		}
		return data, nil
	}
	return io.ReadAll(file)
}

func writeDurableFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	closeFile := true
	defer func() {
		if closeFile {
			_ = file.Close()
		}
	}()
	n, err := file.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return io.ErrShortWrite
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	closeFile = false
	return nil
}
