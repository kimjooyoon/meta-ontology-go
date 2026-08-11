package freshness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadManifest decodes a JSON snapshot. If Root is omitted, paths are
// resolved relative to the manifest directory.
func LoadManifest(path string) (Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, fmt.Errorf("read freshness manifest: %w", err)
	}
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("decode freshness manifest: %w", err)
	}
	if snapshot.Root == "" {
		absolute, err := filepath.Abs(filepath.Dir(path))
		if err != nil {
			return Snapshot{}, fmt.Errorf("resolve freshness manifest root: %w", err)
		}
		snapshot.Root = absolute
	}
	return snapshot, nil
}

// CheckManifest loads path and checks its snapshot. Decode and I/O failures
// returned by LoadManifest are fatal; record-level missing/stale states are
// represented in the returned report.
func CheckManifest(path string) (Report, error) {
	snapshot, err := LoadManifest(path)
	if err != nil {
		return Report{}, err
	}
	return Check(snapshot), nil
}
