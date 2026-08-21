package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunGeneratePreservesPreviousGoAndPublishesManifest(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "main.gooo")
	if err := os.WriteFile(sourcePath, []byte(validSource), 0o640); err != nil {
		t.Fatal(err)
	}
	firstDir := filepath.Join(root, "first")
	var stdout, stderr bytes.Buffer
	if code := runGenerate([]string{sourcePath, "--out", firstDir}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("initial generate = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	firstPath := filepath.Join(firstDir, generatedFileName)
	first, err := os.ReadFile(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	userBody := "return Order{\n\t\t// user-owned\n\t}"
	previous := bytes.Replace(first, []byte("return Order{}"), []byte(userBody), 1)
	if bytes.Equal(first, previous) {
		t.Fatal("fixture did not create a user-owned previous body")
	}
	previousPath := filepath.Join(root, "previous.go")
	if err := os.WriteFile(previousPath, previous, 0o640); err != nil {
		t.Fatal(err)
	}
	secondDir := filepath.Join(root, "second")
	manifestPath := filepath.Join(root, "evidence", "projection.jsonl")
	stdout.Reset()
	stderr.Reset()
	if code := runGenerate([]string{sourcePath, "--out", secondDir, "--previous-go", previousPath, "--manifest", manifestPath}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK {
		t.Fatalf("previous-Go generate = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	second, err := os.ReadFile(filepath.Join(secondDir, generatedFileName))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(second, []byte("// user-owned")) || !bytes.Contains(second, []byte(userBody)) {
		t.Fatalf("previous Go slot was not preserved:\n%s", second)
	}
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest projectionManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("manifest is not JSONL JSON: %v", err)
	}
	if manifest.Schema != projectionManifestSchema || manifest.Status != "pass" || !manifest.PreviousGoProvided || manifest.PreviousGoDigest == "" || !manifest.ProtectedBytesEqual || manifest.ResponseDigest == "" || manifest.EvidenceManifest.PayloadSHA256 == "" {
		t.Fatalf("incomplete previous-Go manifest: %#v", manifest)
	}
	if got, err := os.ReadFile(sourcePath); err != nil || !bytes.Equal(got, []byte(validSource)) {
		t.Fatalf("source was modified: %v", err)
	}
}
