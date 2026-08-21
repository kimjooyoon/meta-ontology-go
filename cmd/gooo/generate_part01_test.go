package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGenerateWritesDeterministicProjection(t *testing.T) {
	outputDir := t.TempDir()
	var stdout, stderr bytes.Buffer
	args := []string{"billing.gooo", "--out", outputDir}
	code := runGenerate(args, fixtureReader{source: validSource}, SyntaxSourceParser{}, &stdout, &stderr)
	if code != exitOK || stderr.Len() != 0 {
		t.Fatalf("generate result = code %d, stdout %q, stderr %q", code, stdout.String(), stderr.String())
	}
	path := filepath.Join(outputDir, "semantic.gooo.go")
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(first), `//gooo:generated:start id="billing://activity/pay-order"`) {
		t.Fatalf("generated source lacks activity marker:\n%s", first)
	}
	if code := runGenerate(args, fixtureReader{source: validSource}, SyntaxSourceParser{}, &bytes.Buffer{}, &bytes.Buffer{}); code != exitOK {
		t.Fatalf("second generate code = %d", code)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("identical source did not generate identical output")
	}
	if _, err := os.Stat(filepath.Join(outputDir, generatedManifestFileName)); err != nil {
		t.Fatalf("generate did not publish manifest: %v", err)
	}
}
