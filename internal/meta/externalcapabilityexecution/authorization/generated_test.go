package authorization

import (
	"bytes"
	"testing"
)

func TestGeneratedManifestIgnoresOnlyNonSemanticMetadata(t *testing.T) {
	first := []byte(`{"schema":"v1","generated_file":"/tmp/first/policy.go","analysis_provenance":{"source_digest":"first"},"digest":"same"}`)
	replay := []byte(`{"schema":"v1","generated_file":"/tmp/replay/policy.go","analysis_provenance":{"source_digest":"replay"},"digest":"same"}`)
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

func TestGeneratedManifestPreservesSemanticDifferences(t *testing.T) {
	baseline := []byte(`{"schema":"v1","analysis_provenance":{"source_digest":"same"},"digest":"same"}`)
	tampered := []byte(`{"schema":"v1","analysis_provenance":{"source_digest":"same"},"digest":"changed"}`)
	left, err := normalizeGenerated("semantic.gooo.manifest.jsonl", baseline)
	if err != nil {
		t.Fatal(err)
	}
	right, err := normalizeGenerated("semantic.gooo.manifest.jsonl", tampered)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(left, right) {
		t.Fatal("semantic manifest mutation was normalized away")
	}
}
