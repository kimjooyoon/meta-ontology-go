package transformationeffect

import (
	"bytes"
	"testing"
)

func TestCanonicalMetricsPayloadNormalizesWorkspacePath(t *testing.T) {
	root := "/tmp/metrics-workspace-123"
	first, err := canonicalMetricsPayload([]byte(`{"commit_sha":"head","root":"/tmp/metrics-workspace-123","storage_root":"/tmp/metrics-workspace-123","files":[{"path":"fixture.go","language":"go","lines":1}],"directories":[],"semantic_path":"/tmp/metrics-workspace-123"}`), root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := canonicalMetricsPayload([]byte(`{"commit_sha":"head","root":"/tmp/metrics-workspace-456","storage_root":"/tmp/metrics-workspace-456","files":[{"path":"fixture.go","language":"go","lines":1}],"directories":[],"semantic_path":"/tmp/metrics-workspace-123"}`), "/tmp/metrics-workspace-456")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) || !bytes.Contains(first, []byte(`"root":"<workspace>"`)) || !bytes.Contains(first, []byte(`"storage_root":"<workspace>"`)) || !bytes.Contains(first, []byte(`"semantic_path":"/tmp/metrics-workspace-123"`)) {
		t.Fatalf("canonical metrics payloads differ or omit normalized roots: first=%s second=%s", first, second)
	}

	changed, err := canonicalMetricsPayload([]byte(`{"commit_sha":"head","root":"/tmp/metrics-workspace-123","storage_root":"/tmp/metrics-workspace-123","files":[{"path":"fixture.go","language":"go","lines":1}],"directories":[],"semantic_path":"<workspace>"}`), root)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(first, changed) {
		t.Fatal("canonical metrics payload erased a substantive non-root value")
	}

	if _, err := canonicalMetricsPayload([]byte(`{"root":`), "/tmp/metrics-workspace-123"); err == nil {
		t.Fatal("malformed metrics payload unexpectedly canonicalized")
	}
}
