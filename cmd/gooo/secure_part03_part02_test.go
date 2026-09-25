package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// CLI-PATH-007: output targets must be regular files when they already exist.
func TestCLIPATH007RejectsNonRegularOutput(t *testing.T) {
	outputDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(outputDir, generatedFileName), 0o755); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	code := runGenerate([]string{"billing.gooo", "--out", outputDir}, fixtureReader{source: validSource}, SyntaxSourceParser{}, &bytes.Buffer{}, &stderr)
	if code != exitFailure || !strings.Contains(stderr.String(), "not a regular file") {
		t.Fatalf("non-regular output = code %d, stderr %q", code, stderr.String())
	}
}
