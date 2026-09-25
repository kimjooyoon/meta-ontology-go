package freshness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LoadManifest decodes one strict JSON snapshot. If Root is omitted, paths
// are resolved relative to the manifest directory.
func LoadManifest(path string) (Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, fmt.Errorf("read freshness manifest: %w", err)
	}
	if err := rejectDuplicateKeys(data); err != nil {
		return Snapshot{}, fmt.Errorf("decode freshness manifest: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var snapshot *Snapshot
	if err := decoder.Decode(&snapshot); err != nil {
		return Snapshot{}, fmt.Errorf("decode freshness manifest: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return Snapshot{}, fmt.Errorf("decode freshness manifest: %w", err)
	}
	if snapshot == nil {
		return Snapshot{}, fmt.Errorf("decode freshness manifest: snapshot must be an object")
	}
	if snapshot.Root == "" {
		absolute, err := filepath.Abs(filepath.Dir(path))
		if err != nil {
			return Snapshot{}, fmt.Errorf("resolve freshness manifest root: %w", err)
		}
		snapshot.Root = absolute
	}
	return *snapshot, nil
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
func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}
func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(bytes.TrimSpace(data)))
	decoder.UseNumber()
	if err := walkJSON(decoder); err != nil {
		return err
	}
	return requireEOF(decoder)
}
