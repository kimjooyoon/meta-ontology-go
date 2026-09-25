package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEntityFieldsDeferredAnalyzeRouteHasNoFixPlanOutput(t *testing.T) {
	var stdout, stderr strings.Builder
	reader := entityFieldsMapReader{"fields.gooo": []byte(deferredEntityFieldsSource)}
	if code := runAnalyze([]string{"fields.gooo"}, reader, SyntaxSourceParser{}, &stdout, &stderr); code != exitFailure {
		t.Fatalf("deferred analyze code = %d", code)
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "parse.entity-fields-deferred") {
		t.Fatalf("deferred analyze output = stdout %q stderr %q", stdout.String(), stderr.String())
	}
}
func TestGenerateArtifactTransactionRestoresBothOutputsOnLaterRenameFailure(t *testing.T) {
	root := t.TempDir()
	outputPath := filepath.Join(root, generatedFileName)
	manifestPath := filepath.Join(root, generatedManifestFileName)
	oldOutput, oldManifest := []byte("old generated\n"), []byte("old manifest\n")
	if err := os.WriteFile(outputPath, oldOutput, 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, oldManifest, 0o640); err != nil {
		t.Fatal(err)
	}
	ops := defaultAtomicFileOps()
	renames := 0
	ops.rename = func(oldPath, newPath string) error {
		renames++
		if renames == 2 {
			return errors.New("injected second rename failure")
		}
		return os.Rename(oldPath, newPath)
	}
	err := writeAtomicFilesWithOps([]atomicWrite{{path: outputPath, data: []byte("new generated\n")}, {path: manifestPath, data: []byte("new manifest\n")}}, ops)
	if err == nil || !strings.Contains(err.Error(), "injected second rename failure") {
		t.Fatalf("transaction error = %v", err)
	}
	gotOutput, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	gotManifest, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotOutput, oldOutput) || !bytes.Equal(gotManifest, oldManifest) {
		t.Fatalf("transaction rollback changed pre-state: output=%q manifest=%q", gotOutput, gotManifest)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("transaction left temporary files: %#v", entries)
	}
}
