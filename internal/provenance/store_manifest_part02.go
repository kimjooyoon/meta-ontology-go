package provenance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func writeManifest(path string, manifest ledgerManifest, recovery bool) error {
	directory := filepath.Dir(manifestPath(path))
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create provenance manifest directory: %w", err)
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("marshal provenance manifest: %w", err)
	}
	payload = append(payload, '\n')
	points := manifestPoints(manifest, recovery)
	temporary, err := os.CreateTemp(directory, ".provenance-manifest-*")
	if err != nil {
		return fmt.Errorf("create provenance manifest: %w", err)
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	if err := writeFullAt(temporary, payload, points.write); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write provenance manifest: %w", err)
	}
	if err := syncFile(temporary, points.sync); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync provenance manifest: %w", err)
	}
	if err := closeFile(temporary, points.close); err != nil {
		return fmt.Errorf("close provenance manifest: %w", err)
	}
	if err := failStorageOperation(points.rename); err != nil {
		return fmt.Errorf("commit provenance manifest: %w", err)
	}
	if err := os.Rename(temporaryName, manifestPath(path)); err != nil {
		return fmt.Errorf("commit provenance manifest: %w", err)
	}
	if err := syncDirectory(directory, points.directorySync); err != nil {
		return fmt.Errorf("sync provenance manifest directory: %w", err)
	}
	return nil
}
