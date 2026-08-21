package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type entityFieldsMapReader map[string][]byte
type deferredCLIWorkspace struct {
	parent, outputDir, manifestPath, storePath, tempPath, sourcePath, cachePath string
}

func (reader entityFieldsMapReader) ReadFile(name string) ([]byte, error) {
	data, ok := reader[name]
	if !ok {
		return nil, fmt.Errorf("missing fixture %q", name)
	}
	return append([]byte(nil), data...), nil
}
func prepareDeferredCLIWorkspace(t *testing.T) deferredCLIWorkspace {
	t.Helper()
	workspace := deferredCLIWorkspace{parent: t.TempDir()}
	workspace.outputDir = filepath.Join(workspace.parent, "output")
	workspace.manifestPath = filepath.Join(workspace.parent, "evidence", "projection.jsonl")
	workspace.storePath = filepath.Join(workspace.parent, "evidence", "provenance.jsonl")
	workspace.tempPath = filepath.Join(workspace.outputDir, ".semantic.gooo.go.tmp-stale")
	workspace.sourcePath = filepath.Join(workspace.parent, "fields.gooo")
	workspace.cachePath = filepath.Join(workspace.parent, "cache", "semantic.cache")
	for path, data := range map[string][]byte{
		filepath.Join(workspace.outputDir, generatedFileName): []byte("old generated\n"),
		workspace.manifestPath:                                []byte("old manifest\n"),
		workspace.storePath:                                   []byte("old evidence\n"),
		workspace.tempPath:                                    []byte("old temp\n"),
		workspace.sourcePath:                                  []byte(deferredEntityFieldsSource),
		workspace.cachePath:                                   []byte("old cache\n"),
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o640); err != nil {
			t.Fatal(err)
		}
	}
	return workspace
}
func assertDeferredCLIWorkspace(t *testing.T, workspace deferredCLIWorkspace) {
	t.Helper()
	for path, want := range map[string][]byte{
		filepath.Join(workspace.outputDir, generatedFileName): []byte("old generated\n"),
		workspace.manifestPath:                                []byte("old manifest\n"),
		workspace.storePath:                                   []byte("old evidence\n"),
		workspace.tempPath:                                    []byte("old temp\n"),
		workspace.sourcePath:                                  []byte(deferredEntityFieldsSource),
		workspace.cachePath:                                   []byte("old cache\n"),
	} {
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("deferred CLI changed %s: got=%q err=%v", path, got, err)
		}
	}
}
