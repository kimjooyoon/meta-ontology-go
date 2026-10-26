// Package recoveryorigin binds a declared recovery scope to retained observations.
// The caller owns collection/authenticity of the original GitHub API snapshots.
package recoveryorigin

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Read returns both the classified observation and its original file digests.
// Missing/unreadable input remains visible to the Gooo policy through read_error.
// It performs no network, registry, model or mutation operations.
func Read(root, role string, jobID int64) (R_ReceiptOrigin, map[string]string) {
	origin, digests := R_ReceiptOrigin{}, map[string]string{}
	if role != "publication" && role != "readback" {
		origin.F_read_error = "unknown receipt role"
		return origin, digests
	}
	if err := readOrigin(root, role, jobID, &origin, digests); err != nil {
		origin.F_read_error = err.Error()
	}
	return origin, digests
}

func readOrigin(root, role string, jobID int64, out *R_ReceiptOrigin, digests map[string]string) error {
	var run runRecord
	var jobs jobRecords
	var artifact artifactRecord
	for _, part := range []struct {
		suffix string
		target any
	}{{"run.json", &run}, {"jobs.json", &jobs}, {"artifact.json", &artifact}} {
		raw, err := readBounded(root, role+"-"+part.suffix, digests)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, part.target); err != nil {
			return fmt.Errorf("%s-%s: %w", role, part.suffix, err)
		}
	}
	if err := copyOrigin(run, jobs, artifact, jobID, out); err != nil {
		return err
	}
	if _, err := readBounded(root, "workflow.source", digests); err != nil {
		return err
	}
	out.F_workflow_sha256 = digests["workflow.source"]
	archive, err := readBounded(root, role+".zip", digests)
	if err != nil {
		return err
	}
	out.F_archive_sha256 = "sha256:" + digests[role+".zip"]
	return readImage(archive, role, out)
}

func readBounded(root, name string, digests map[string]string) ([]byte, error) {
	f, err := os.Open(filepath.Join(root, name))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	defer f.Close()
	const limit = 1 << 20
	raw, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	if len(raw) > limit {
		return nil, fmt.Errorf("%s: exceeds %d bytes", name, limit)
	}
	digests[name] = fmt.Sprintf("%x", sha256.Sum256(raw))
	return raw, nil
}
