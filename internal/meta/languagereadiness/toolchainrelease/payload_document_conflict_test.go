package toolchainrelease

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestPayloadDocumentRejectsChangedModuleAndLicense(t *testing.T) {
	root := payloadTestRoot(t)
	original, err := os.ReadFile(filepath.Join("../../../..", payloadDocumentPath))
	if err != nil {
		t.Fatal(err)
	}
	document := filepath.Join(root, payloadDocumentPath)
	if err := os.MkdirAll(filepath.Dir(document), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(document, original, 0o644); err != nil {
		t.Fatal(err)
	}
	modulePath := filepath.Join(root, "go.mod")
	module, err := os.ReadFile(modulePath)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := readPayloadPolicy(root, Targets()[0])
	if err != nil {
		t.Fatal(err)
	}
	changedModule := bytes.Replace(module, []byte(policy.Members.RuntimeVersion),
		[]byte("v0.0.0-gooo-mismatch"), 1)
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
