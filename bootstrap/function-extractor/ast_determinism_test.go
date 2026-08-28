package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGenericExtractionReplayIsByteIdentical(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.test\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := []byte("package p\n\nfunc First() {}\n\nfunc Second() {}\n")
	if err := os.WriteFile(filepath.Join(root, "x.go"), source, 0o644); err != nil {
		t.Fatal(err)
	}
	first, paths, err := genericASTExtraction(root, "x.go")
	if err != nil {
		t.Fatal(err)
	}
	second, secondPaths, err := genericASTExtraction(root, "x.go")
	if err != nil || len(paths) != len(secondPaths) {
		t.Fatal(err)
	}
	for _, path := range paths {
		if !bytes.Equal(first[path], second[path]) {
			t.Fatalf("nondeterministic output for %s", path)
		}
	}
}
