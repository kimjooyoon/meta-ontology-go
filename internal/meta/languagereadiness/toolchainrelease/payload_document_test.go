package toolchainrelease

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPayloadDocumentMatchesGoooSource(t *testing.T) {
	result, err := ProjectPayloadModuleDocument("../../../..", false)
	if err != nil || result.Changed || result.Written || result.Provenance != nil {
		t.Fatalf("committed release document is stale: %+v, %v", result, err)
	}
}

func TestPayloadDocumentProjectsOnlyMarkedTable(t *testing.T) {
	root := payloadTestRoot(t)
	policy, err := readPayloadPolicy(root, Targets()[0])
	if err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(filepath.Join("../../../..", payloadDocumentPath))
	if err != nil {
		t.Fatal(err)
	}
	stale := []byte(strings.Replace(string(original), policy.Members.RuntimeVersion+"` /",
		"v0.0.0-gooo-mismatch` /", 1))
	if bytes.Equal(stale, original) {
		t.Fatal("release module row fixture was not found")
	}
	path := filepath.Join(root, payloadDocumentPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, stale, 0o644); err != nil {
		t.Fatal(err)
	}
	check, err := ProjectPayloadModuleDocument(root, false)
	if err != nil || !check.Changed || check.Written || check.Provenance != nil {
		t.Fatalf("check mode did not expose stale text without writing: %+v, %v", check, err)
	}
	unchanged, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(unchanged, stale) {
		t.Fatal("check mode changed the source document")
	}
	written, err := ProjectPayloadModuleDocument(root, true)
	if err != nil || !written.Changed || !written.Written || written.Provenance["@graph"] == nil {
		t.Fatalf("write mode did not record the actual projection: %+v, %v", written, err)
	}
	graph := written.Provenance["@graph"].([]any)
	if len(graph) != 7 || len(graph[5].(map[string]any)["prov:used"].([]any)) != 5 {
		t.Fatal("projection omitted a checked source or target preimage")
	}
	actual, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(actual, original) || written.OutputSHA256 != payloadDigest(original) {
		t.Fatal("projection failed to restore the exact Gooo-derived table")
	}
	if _, err := replacePayloadModuleTable([]byte("no markers"), "table"); err == nil {
		t.Fatal("unmarked document was accepted for projection")
	}
}
