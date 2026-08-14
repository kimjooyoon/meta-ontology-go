package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCanonicalizeBindsContentAndIgnoresWorkspaceRoot(t *testing.T) {
	left, right := t.TempDir(), t.TempDir()
	manifest := func(root string, generated []byte) []byte {
		value := strings.Repeat("1", 64)
		return []byte(fmt.Sprintf(`{"generated_file":%q,"semantic_digest":"%s","source_digest":"%s","generated_digest":%q,"source_map_digest":"%s","response_digest":"%s","evidence_manifest":{"payload_sha256":"%s"}}`+"\n", filepath.Join(root, "semantic.gooo.go"), value, strings.Repeat("2", 64), digest(generated), strings.Repeat("4", 64), strings.Repeat("5", 64), strings.Repeat("6", 64)))
	}
	for _, root := range []string{left, right} {
		generated := []byte("//gooo:generated\n")
		if err := os.WriteFile(filepath.Join(root, "semantic.gooo.go"), generated, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "semantic.gooo.manifest.jsonl"), manifest(root, generated), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	leftData, err := Canonicalize(left, mustRead(t, filepath.Join(left, "semantic.gooo.manifest.jsonl")))
	if err != nil {
		t.Fatal(err)
	}
	rightData, err := Canonicalize(right, mustRead(t, filepath.Join(right, "semantic.gooo.manifest.jsonl")))
	if err != nil {
		t.Fatal(err)
	}
	if string(leftData) != string(rightData) {
		t.Fatal("workspace-specific generated_file changed canonical manifest bytes")
	}
}

func TestCanonicalizeRejectsEscapedPathAndTamperedGeneratedDigest(t *testing.T) {
	root := t.TempDir()
	generated := []byte("//gooo:generated\n")
	if err := os.WriteFile(filepath.Join(root, "semantic.gooo.go"), generated, 0o600); err != nil {
		t.Fatal(err)
	}
	valid := fmt.Sprintf(`{"generated_file":%q,"semantic_digest":"%s","source_digest":"%s","generated_digest":%q,"source_map_digest":"%s","response_digest":"%s","evidence_manifest":{"payload_sha256":"%s"}}`+"\n", filepath.Join(root, "semantic.gooo.go"), strings.Repeat("1", 64), strings.Repeat("2", 64), digest(generated), strings.Repeat("4", 64), strings.Repeat("5", 64), strings.Repeat("6", 64))
	if _, err := Canonicalize(root, []byte(strings.Replace(valid, filepath.Join(root, "semantic.gooo.go"), filepath.Join(filepath.Dir(root), "escape.go"), 1))); err == nil {
		t.Fatal("manifest path escaping the generated root was accepted")
	}
	if _, err := Canonicalize(root, []byte(strings.Replace(valid, digest(generated), strings.Repeat("3", 64), 1))); err == nil {
		t.Fatal("tampered embedded generated digest was accepted")
	}
}

func mustRead(t *testing.T, filename string) []byte {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
