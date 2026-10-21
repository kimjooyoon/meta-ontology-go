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
	original, err := os.ReadFile(filepath.Join("../../../..", payloadDocumentPath))
	if err != nil {
		t.Fatal(err)
	}
	stale := []byte(strings.Replace(string(original), "v0.2.51-experimental` /", "v0.2.50-experimental` /", 1))
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
	modulePath := filepath.Join(root, "go.mod")
	module, err := os.ReadFile(modulePath)
	if err != nil {
		t.Fatal(err)
	}
	changedModule := bytes.Replace(module, []byte("v0.2.51-experimental"), []byte("v0.2.50-experimental"), 1)
	if bytes.Equal(changedModule, module) {
		t.Fatal("direct module fixture was not found")
	}
	if err := os.WriteFile(modulePath, changedModule, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ProjectPayloadModuleDocument(root, true); err == nil {
		t.Fatal("Gooo policy and go.mod mismatch still projected documentation")
	}
	if err := os.WriteFile(modulePath, module, 0o644); err != nil {
		t.Fatal(err)
	}
	licensePath := filepath.Join(root, "licenses/gooo-jev/LICENSE")
	license, err := os.ReadFile(licensePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(licensePath, append(license, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ProjectPayloadModuleDocument(root, true); err == nil {
		t.Fatal("changed dependency license still projected its old hash")
	}
}
