package main

import (
	"bytes"
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"os"
	"path/filepath"
	"testing"
)

func TestEntityFieldsDeferredLSPRoutePublishesOnlySourceDiagnostic(t *testing.T) {
	uri := "file:///fields.gooo"
	input := lspTranscript(
		lspRequest(1, "initialize", nil),
		lspNotification("textDocument/didOpen", map[string]any{
			"textDocument": map[string]any{"uri": uri, "version": 1, "text": deferredEntityFieldsSource},
		}),
		lspRequest(2, "shutdown", nil),
		lspNotification("exit", nil),
	)
	output, code, stderr := runLSPTranscript(t, input)
	if code != exitOK || stderr != "" {
		t.Fatalf("deferred LSP route = code %d stderr=%q output=%q", code, stderr, output)
	}
	if !bytes.Contains(output, []byte("parse.entity-fields-deferred")) || bytes.Contains(output, []byte(`"result":{"symbols"`)) {
		t.Fatalf("deferred LSP output did not stay diagnostic-only: %s", output)
	}
}
func TestFieldlessBillingProjectionBytesAndEvidenceHashesRemainStable(t *testing.T) {
	root := t.TempDir()
	firstDir := filepath.Join(root, "first")
	secondDir := filepath.Join(root, "second")
	fixture := filepath.Join("..", "..", "examples", "billing", "main.gooo")
	if code := runGenerate([]string{fixture, "--out", firstDir}, OSFileReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitOK {
		t.Fatalf("first billing generate = %d", code)
	}
	if code := runGenerate([]string{fixture, "--out", secondDir}, OSFileReader{}, SyntaxSourceParser{}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitOK {
		t.Fatalf("second billing generate = %d", code)
	}
	firstSource, err := os.ReadFile(filepath.Join(firstDir, generatedFileName))
	if err != nil {
		t.Fatal(err)
	}
	secondSource, err := os.ReadFile(filepath.Join(secondDir, generatedFileName))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstSource, secondSource) || semantic.StableHash(firstSource) != "3c0ca7a65301c490a6732d4c8635c0dda5d934bb14a6cf645dddc792fffea5d6" {
		t.Fatalf("fieldless billing source changed: first=%s second=%s", semantic.StableHash(firstSource), semantic.StableHash(secondSource))
	}
	var firstManifest, secondManifest projectionManifest
	firstBytes, err := os.ReadFile(filepath.Join(firstDir, generatedManifestFileName))
	if err != nil {
		t.Fatal(err)
	}
	secondBytes, err := os.ReadFile(filepath.Join(secondDir, generatedManifestFileName))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(firstBytes, &firstManifest); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(secondBytes, &secondManifest); err != nil {
		t.Fatal(err)
	}
	if firstManifest.SemanticDigest != secondManifest.SemanticDigest || firstManifest.GeneratedDigest != secondManifest.GeneratedDigest || firstManifest.SourceMapDigest != secondManifest.SourceMapDigest || firstManifest.ResponseDigest != secondManifest.ResponseDigest || firstManifest.EvidenceManifest.PayloadSHA256 != secondManifest.EvidenceManifest.PayloadSHA256 {
		t.Fatalf("fieldless billing evidence replay diverged: first=%#v second=%#v", firstManifest, secondManifest)
	}
}
