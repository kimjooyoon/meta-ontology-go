package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAnalyzeRejectsStaleGeneratedMarkersWithoutWrite(t *testing.T) {
	_, generated := billingAnalyzeFiles(t, billingAnalyzeAuthority)
	source, err := os.ReadFile(generated)
	if err != nil {
		t.Fatal(err)
	}
	mutated := bytes.Replace(source, []byte(`//gooo:generated:end id="billing://entity/order" kind="entity"`), []byte(`//gooo:generated:end id="billing://entity/stale" kind="entity"`), 1)
	if bytes.Equal(source, mutated) {
		t.Fatal("stale marker mutation did not apply")
	}
	if err := os.WriteFile(generated, mutated, 0o640); err != nil {
		t.Fatal(err)
	}
	before, info := snapshotFile(t, generated)
	authority := filepath.Join(filepath.Dir(generated), "authority.gooo")
	if err := os.WriteFile(authority, []byte(billingAnalyzeAuthority), 0o640); err != nil {
		t.Fatal(err)
	}
	output, code, stderr := runAnalyzePaths(authority, generated)
	if code != exitFailure || len(output) != 0 || !strings.Contains(stderr, "generated region") {
		t.Fatalf("stale marker result = code %d, stdout=%q, stderr=%q", code, output, stderr)
	}
	after, afterInfo := snapshotFile(t, generated)
	if !bytes.Equal(before, after) || !os.SameFile(info, afterInfo) || info.ModTime() != afterInfo.ModTime() {
		t.Fatal("stale marker rejection mutated Go input")
	}
}
func billingAnalyzeFiles(t *testing.T, authoritySource string) (string, string) {
	t.Helper()
	authority, _ := writeAnalyzeFile(t, "main.gooo", authoritySource)
	outputDir := filepath.Join(filepath.Dir(authority), "generated")
	var stdout, stderr bytes.Buffer
	if code := runGenerate([]string{authority, "--out", outputDir}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr); code != exitOK || stderr.Len() != 0 {
		t.Fatalf("generate billing analyze fixture = %d, stdout=%q, stderr=%q", code, stdout.String(), stderr.String())
	}
	return authority, filepath.Join(outputDir, generatedFileName)
}
func writeAnalyzeFile(t *testing.T, name, source string) (string, []byte) {
	t.Helper()
	path := name
	if !filepath.IsAbs(path) {
		path = filepath.Join(t.TempDir(), name)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	data := []byte(source)
	if err := os.WriteFile(path, data, 0o640); err != nil {
		t.Fatal(err)
	}
	return path, data
}
func runAnalyzePaths(authority, generated string) ([]byte, int, string) {
	var stdout, stderr bytes.Buffer
	code := runAnalyze([]string{authority, "--go", generated}, OSFileReader{}, SyntaxSourceParser{}, &stdout, &stderr)
	return stdout.Bytes(), code, stderr.String()
}
