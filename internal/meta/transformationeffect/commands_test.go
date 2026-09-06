package transformationeffect

import (
	"bytes"
	"testing"
)

func TestCanonicalMetricsPayloadNormalizesWorkspacePath(t *testing.T) {
	root := "/tmp/metrics-workspace-123"
	payload := []byte(`{"root":"/tmp/metrics-workspace-123","storage_root":"/tmp/metrics-workspace-123","other":"/tmp/other-workspace"}`)
	want := []byte(`{"root":"<workspace>","storage_root":"<workspace>","other":"/tmp/other-workspace"}`)
	if got := canonicalMetricsPayload(payload, root); !bytes.Equal(got, want) {
		t.Fatalf("canonical metrics payload = %s, want %s", got, want)
	}
}
