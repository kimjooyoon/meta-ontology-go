package authorization

import (
	"bytes"
	"testing"
)

func TestGeneratedManifestIgnoresOnlyOutputPath(t *testing.T) {
	first := []byte(`{"schema":"v1","generated_file":"/tmp/first/policy.go","digest":"same"}`)
	replay := []byte(`{"schema":"v1","generated_file":"/tmp/replay/policy.go","digest":"same"}`)
	left, err := normalizeGenerated("semantic.gooo.manifest.jsonl", first)
	if err != nil {
		t.Fatal(err)
	}
	right, err := normalizeGenerated("semantic.gooo.manifest.jsonl", replay)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(left, right) {
		t.Fatalf("normalized manifests differ: %s != %s", left, right)
	}
}
